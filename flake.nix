{
  description = "dev";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
    # Shared development environment (provides pinned Go + Zig 0.15.1 + common tools)
    dev-env.url = "github:spyderorg/dev-env";
    gomod2nix.url = "github:nix-community/gomod2nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      dev-env,
      flake-utils,
      gomod2nix,
      ...
    }:
    let
      flakeForSystem =
        nixpkgs: system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          # Pull toolchains / common tooling from centralized dev-env flake
          inherit (dev-env.packages.${system}) zigPackages goPackages commonPackages;
          buildGoApp = gomod2nix.legacyPackages.${system}.buildGoApplication;

          # Use GOPRIVATE only (default to go.spyder.org)
          goPrivate =
            let
              gpr = builtins.getEnv "GOPRIVATE";
            in
            if gpr != "" then gpr else "go.spyder.org";

          # A git-askpass helper that fetches token from env or gh CLI at build time
          gitAskpass = pkgs.writeShellScript "git-askpass.sh" ''
            prompt="$1"
            get_token() {
              if [ -n "$GITHUB_TOKEN" ]; then
                printf "%s" "$GITHUB_TOKEN"; return 0
              fi
              if [ -n "$GH_TOKEN" ]; then
                printf "%s" "$GH_TOKEN"; return 0
              fi
              if command -v gh >/dev/null 2>&1; then
                gh auth token 2>/dev/null || true; return 0
              fi
              printf ""
            }
            case "$prompt" in
              *'Username for https://github.com'*) echo "x-access-token" ;;
              *'Username for https://x-access-token@github.com'*) echo "x-access-token" ;;
              *'Password for https://x-access-token@github.com'*) get_token ;;
              *'Password for https://github.com'*) get_token ;;
              *) echo "" ;;
            esac
          '';

          gitAuthEnv = {
            GIT_TERMINAL_PROMPT = "0";
            GIT_ASKPASS = toString gitAskpass;
          };

          preBuildWithGh = ''
            export HOME=$TMPDIR
            # Hydrate token from gh if env is empty
            if [ -z "$GITHUB_TOKEN" ] && [ -z "$GH_TOKEN" ] && command -v gh >/dev/null 2>&1; then
              TOK=$(gh auth token 2>/dev/null || true)
              if [ -n "$TOK" ]; then
                export GITHUB_TOKEN="$TOK"
                export GH_TOKEN="$TOK"
              fi
            fi
            TOKEN=""
            if [ -n "$GITHUB_TOKEN" ]; then
              TOKEN="$GITHUB_TOKEN"
            elif [ -n "$GH_TOKEN" ]; then
              TOKEN="$GH_TOKEN"
            fi
            if [ -n "$TOKEN" ]; then
              git config --global url."https://x-access-token:$TOKEN@github.com/".insteadOf "https://github.com/"
              git config --global url."https://x-access-token:$TOKEN@github.com/".insteadOf "git@github.com:"
            fi
          '';
        in
        rec {
          devShell = pkgs.mkShell {
            shellHook = ''
              export GOPRIVATE=${goPrivate}
              export GONOPROXY=${goPrivate}
              export GONOSUMDB=${goPrivate}
              # Optionally hydrate tokens from gh if available (dev only)
              if command -v gh >/dev/null 2>&1; then
                TOK=$(gh auth token 2>/dev/null || true)
                if [ -n "$TOK" ]; then
                  export GITHUB_TOKEN="$TOK"
                  export GH_TOKEN="$TOK"
                fi
              fi
              echo "✅ Canary development environment loaded"
              echo "   - canary binary available in PATH"
              echo "   - Run 'canary --help' to get started"
            '';
            env = {
              GOPRIVATE = goPrivate;
              GONOPROXY = goPrivate;
              GONOSUMDB = goPrivate;
            }
            // gitAuthEnv;
            packages =
              # Base developer toolchain (Go + Zig + shared utilities)
              [
                goPackages
                commonPackages
              ]
              # Add the canary binary to the development environment
              ++ [ packages.canary ];
          };

          packages = {
            canary = buildGoApp (
              {
                pname = "canary";
                version = "dev";
                src = self;
                modules = ./gomod2nix.toml;
                subPackages = [ "cmd/canary" ];
                GOPRIVATE = goPrivate;
                GONOPROXY = goPrivate;
                GONOSUMDB = goPrivate;
                nativeBuildInputs = [
                  pkgs.gh
                  pkgs.git
                ];
                preBuild = preBuildWithGh;
              }
              // gitAuthEnv
            );
          };

          defaultPackage = packages.canary;
        };
    in
    flake-utils.lib.eachDefaultSystem (system: flakeForSystem nixpkgs system);

  nixConfig = {
    extra-substituters = [ "https://spyder.cachix.org" ];
    extra-trusted-public-keys = [
      "spyder.cachix.org-1:xZG2a8INH6yNyOAwtN2Ojjys0GO9D1pXEe3PNriT04E="
    ];
  };
}
