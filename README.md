# stowy

`stowy` is a small, experimental Go tool for installing Arch Linux packages into a user-local prefix instead of writing to system-owned directories. It is designed as a learning project for understanding how package archives can be downloaded, extracted, and exposed from a home-directory layout without requiring `sudo`.

This is not a replacement for `pacman`, and it is not intended to be a hardened production package manager. It is a prototype for experimenting with user-space package installation patterns.

## What it does

- Searches Arch repository databases by package name or description
- Downloads repository metadata from Arch mirrors
- Installs packages from Arch repositories or local `.pkg.tar.zst` archives
- Extracts package contents into a managed prefix under `~/.local/stowy`
- Exposes programs from that prefix via launchers placed in `~/.local/bin`
- Removes packages previously installed by `stowy`
- Rejects unsafe archive entries and unsupported package paths

## Why this exists

The project explores a simple idea: keep package metadata and extracted files under a user-owned prefix, make binaries available on the `PATH`, and do it without modifying the root filesystem. That makes it useful for experiments, local tooling, and package-learning work outside of a full system package manager.

## Project layout

- `main.go` — CLI entry point and command dispatch
- `repo.go` — repo database download and retrieval logic
- `search.go` — repository search helpers
- `meta.go` and `meta/meta.go` — package metadata parsing
- `install/` — archive extraction and installation logic
- `testdata/` — fixture data used in tests

## Installation

From the repository root, build the binary:

```bash
go build ./...
```

This produces a `stowy` executable in the working directory.

## Usage

### Search a repository

```bash
./stowy s <repo> <term>
```

Example:

```bash
./stowy s core ripgrep
```

This downloads the repository database for the selected mirror and prints matching packages.

### Install a package from Arch repos

```bash
./stowy i <package>
```

Example:

```bash
./stowy i ripgrep
```

`stowy` checks the `core`, `extra`, and `multilib` repositories and installs the best match it finds.

### Install a local package archive

```bash
./stowy l <package.tar.zst>
```

Example:

```bash
./stowy l /var/cache/pacman/pkg/ripgrep-14.1.0-1-x86_64.pkg.tar.zst
```

### Remove a package

```bash
./stowy uninstall <package>
```

Aliases supported by the CLI include:

```bash
./stowy d ripgrep
./stowy remove ripgrep
./stowy delete ripgrep
```

## Install target

Package files are unpacked under:

```bash
~/.local/stowy
```

The CLI also creates launchers in:

```bash
~/.local/bin
```

These wrappers help make installed executables available on your `PATH` while giving the launched process access to the user-local library directory used by the package install.

## Safety and limitations

Warning: this tool is intentionally a prototype, not a trusted package management workflow.

At the moment, it does not verify package integrity before extracting files. It does not validate cryptographic signatures, compare checksums against trusted metadata, or defend against compromised mirrors or malicious archives in a production-grade way. It expects the repository metadata and package sources it downloads to be trusted by the user.

The current implementation also does not provide:

- automatic dependency resolution
- package database tracking
- authentic signature verification
- robust conflict handling
- hardened production-security guarantees

Use it only with packages you are comfortable inspecting and testing in a controlled environment. Even though it installs into a user-owned prefix instead of system directories, it can still be risky if a package source is compromised or unexpectedly modified.

## Quick start

```bash
go build ./...
./stowy i ripgrep
ripgrep --version
```

## License

This repository does not currently include a license file. If you plan to distribute or publish the project, add an explicit license before doing so.
