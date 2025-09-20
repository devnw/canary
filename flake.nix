{
  description = "dev";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
    # Shared development environment (provides pinned Go + Zig 0.15.1 + common tools)
    dev-env.url = "github:spyderorg/dev-env";
    zcap.url = "github:spyderorg/zcap";
    gomod2nix.url = "github:nix-community/gomod2nix";
  };

  outputs =
    {
      self,
      nixpkgs,
      dev-env,
      flake-utils,
      zcap,
      gomod2nix,
      ...
    }:
    let
      flakeForSystem =
        nixpkgs: system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          zcapPkg = zcap.packages.${system}.zcap;
          zcapLib = zcap.packages.${system}.zcap-static;

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
              # Provide zcap prefixes for Makefile and GoReleaser
              export ZCAP_PREFIX=${zcapPkg}
              export ZCAP_STATIC_PREFIX=${zcapLib}
              # Provide Zig lib dir for static compiler runtime when needed
              export ZIG_LIB_DIR=${pkgs.zig}/lib/zig
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
                zigPackages
                commonPackages
              ]
              ++ [
                zcapPkg
                zcapLib
                pkgs.gh
                pkgs.tshark
                pkgs.suricata
                pkgs.tcpdump
              ];
          };

          packages = {
            pcapinfo = buildGoApp (
              {
                pname = "pcapinfo";
                version = "dev";
                src = self;
                modules = ./gomod2nix.toml;
                subPackages = [ "cmd/pcapinfo" ];
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

            pcapnginfo = buildGoApp (
              {
                pname = "pcapnginfo";
                version = "dev";
                src = self;
                modules = ./gomod2nix.toml;
                subPackages = [ "cmd/pcapnginfo" ];
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

            pcapconv = buildGoApp (
              {
                pname = "pcapconv";
                version = "dev";
                src = self;
                modules = ./gomod2nix.toml;
                subPackages = [ "cmd/pcapconv" ];
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

            pcapfilter = buildGoApp (
              {
                pname = "pcapfilter";
                version = "dev";
                src = self;
                modules = ./gomod2nix.toml;
                subPackages = [ "cmd/pcapfilter" ];
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

            pcapdump = buildGoApp (
              {
                pname = "pcapdump";
                version = "dev";
                src = self;
                modules = ./gomod2nix.toml;
                subPackages = [ "cmd/pcapdump" ];
                CGO_ENABLED = 1;
                tags = [ "zcap" ];
                buildInputs = [
                  zcapPkg
                  zcapLib
                ];
                nativeBuildInputs = [
                  pkgs.pkg-config
                  zigPackages
                  pkgs.gh
                  pkgs.git
                ];
                CC = "${zigPackages}/bin/zig cc";
                CXX = "${zigPackages}/bin/zig c++";
                CGO_CFLAGS = "-I${zcapLib}/include -I${zcapPkg}/include";
                CGO_LDFLAGS = "-Wl,-Bstatic -L${zcapLib}/lib -lzcap -Wl,-Bdynamic -Wl,-rpath,${zcapPkg}/lib";
                GOPRIVATE = goPrivate;
                GONOPROXY = goPrivate;
                GONOSUMDB = goPrivate;
                preBuild = preBuildWithGh;
              }
              // gitAuthEnv
            );

            # Fully static MUSL build using zig (Linux x86_64 only)
            pcapdump-static = buildGoApp (
              {
                pname = "pcapdump-static";
                version = "dev";
                src = self;
                modules = ./gomod2nix.toml;
                subPackages = [ "cmd/pcapdump" ];
                CGO_ENABLED = 1;
                tags = [
                  "zcap"
                  "netgo"
                  "osusergo"
                ];
                buildInputs = [ zcapLib ];
                nativeBuildInputs = [
                  zigPackages
                  pkgs.gh
                  pkgs.git
                ];
                # Restrict to Linux amd64/aarch64 via target flags; zig will static link musl
                CC = "${zigPackages}/bin/zig cc -target x86_64-linux-musl";
                CXX = "${zigPackages}/bin/zig c++ -target x86_64-linux-musl";
                CGO_CFLAGS = "-I${zcapLib}/include";
                CGO_LDFLAGS = "-static -L${zcapLib}/lib -lzcap -L${zigPackages}/lib/zig -l:libcompiler_rt.a";
                ldflags = [
                  "-s"
                  "-w"
                  "-linkmode=external"
                ];
                GOPRIVATE = goPrivate;
                GONOPROXY = goPrivate;
                GONOSUMDB = goPrivate;
                preBuild = preBuildWithGh;
              }
              // gitAuthEnv
            );
          };

          defaultPackage = packages.pcapdump;
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
