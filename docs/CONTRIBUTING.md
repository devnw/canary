# Contributing

We'd love to accept your patches and contributions to this project through the
process of creating a [pull request][https://github.com/devnw/canary] (**PR**). This document details the
process of submitting a PR so that it can be reviewed and merged into the
codebase. It also contains some guidelines for writing good commits, reporting
issues, and guidelines for project maintainers.

---

## Reporting issues

Bugs, feature requests, and development-related questions should be directed to
the specific project's issue tracker or discussion board.

### Bugs

If reporting a bug, please submit an issue and provide as much context as
possible such as your operating system, architecture, library release version,
Go version (if applicable), and anything else that might be relevant to the bug.

Fill out as much information as possible in the form provided by the issue
template.

#### SECURITY BUGS

We take security bugs ***VERY*** seriously!

Please promptly report security related bugs to <security@devnw.com>. Please
follow [responsible disclosure guidelines][SECURITY.md] when publicizing any security related
information, ensuring that maintainers are aware of the issue and are able to
address it promptly.

Please include:

1. Information about the vulnerability
1. Associated CVEs (if any)
1. Affected release(s)
1. Affected package(s)

### Feature Requests

For feature requests, please explain what you're trying to do, and
how the requested feature would help you do that.

Security related bugs can either be reported in the issue tracker, or if they
are more sensitive, emailed to <security@devnw.com>.

[responsible disclosure guidelines]: https://cheatsheetseries.owasp.org/cheatsheets/Vulnerability_Disclosure_Cheat_Sheet.html

---

## Submitting a Pull Request

  1. It's generally best to start by opening a new issue describing the bug or
     feature you're intending to fix. Even if you think it's relatively minor,
     it's helpful to know what people are working on. Mention in the initial
     issue that you are planning to work on that bug or feature so that it can
     be assigned to you.

  1. Follow the normal process of [forking][https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/fork-a-repo] the project, and setup a new
     branch to work in. It's important that each group of changes be done in
     separate branches in order to ensure that a pull request only includes the
     commits related to that bug or feature.
    
  1. This project uses `nix` and `direnv` to manage development environments.
     Please ensure you have both installed and configured on your system.
     See the [development environment documentation](./DEVELOPMENT.md) for more
     information.
---

## Maintainer's Guide

(These notes are mostly only for people merging in pull requests.)

It is the responsibility of the maintainer to ensure that the code is passing
all checks and tests. The maintainer should also ensure that the code is
consistent with the project's [code style][] and [documentation][]. The
maintainer should also ensure that the code is well-documented and tested
before merging in a pull request.

[git-aliases]: https://github.com/willnorris/dotfiles/blob/d640d010c23b1116bdb3d4dc12088ed26120d87d/git/.gitconfig#L13-L15
[rebase-comment]: https://github.com/google/go-github/pull/277#issuecomment-183035491
[modified-comment]: https://github.com/google/go-github/pull/280#issuecomment-184859046
