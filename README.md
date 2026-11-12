# stowy

`stowy` is an experimental Go-based package manager for Arch Linux that installs software into a user-local prefix instead of requiring `sudo`.

It is designed to explore how package archives can be downloaded, extracted, and exposed through a home-directory layout while keeping system packages untouched. The project reads package metadata from Arch repositories, installs matching `.pkg.tar.zst` archives under `~/.local/stowy`, and exposes binaries from the package install tree through `~/.local/bin`.

This project is intentionally a prototype and is not a replacement for `pacman` or a hardened production package manager.

## Features

- Search Arch repositories by package name or description
- Download and inspect repository database metadata
- Install packages from Arch repos or local archive files
- Extract regular files, directories, and relative symlinks safely
- Expose installed executables through launchers in `~/.local/bin`
- Add the user-local `usr/lib` directory to each launcher's shared-library search path
- Remove packages that were installed through `stowy`
- Reject unsafe archive paths and unsupported package entries

## Why this project exists

The goal is to explore the mechanics behind user-local package installation patterns commonly used on Linux systems: read metadata, unpack a package into a managed prefix, and make commands available without modifying system-wide files.

## Project layout

- `main.go` — CLI entry point and command dispatch
- `repo.go` — repository database download logic
- `search.go` — repository search helpers
- `meta.go` and `meta/meta.go` — package metadata parsing
- `install/` — archive extraction and installation logic

## Installation

From the repository root, build the binary:

```bash
go build ./...
```

This produces a `stowy` binary in the current directory.

## Usage

### Search for a package in a repository

```bash
./stowy s <repo> <term>
```

Example:

```bash
./stowy s core ripgrep
```

This downloads the repository database and searches for matching packages.

### Install a package from a repository

```bash
./stowy i <term>
```

Example:

```bash
./stowy i ripgrep
```

`stowy` tries the `core`, `extra`, and `multilib` repositories and installs the best match.

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

Examples:

```bash
./stowy d ripgrep
./stowy remove ripgrep
./stowy delete ripgrep
```

`d`, `delete`, and `remove` are aliases for `uninstall`.

## Install target

Packages are extracted under:

```bash
~/.local/stowy
```

Executables are exposed through launchers in:

```bash
~/.local/bin
```

These launchers add `~/.local/stowy/root/usr/lib` to `LD_LIBRARY_PATH` for the
launched program and its child processes. This lets programs find shared
libraries installed by stowy without changing system linker configuration.

## Safety and limitations

Warning: this project is intentionally a learning prototype and is not a secure package installation workflow.

This tool does not verify cryptographic hashes for downloaded package archives, does not compare SHA-256 against metadata, and currently trusts the repository metadata and package mirrors provided by the Arch ecosystem, including pacman-managed package sources. In other words, package integrity is not independently validated before extraction.

This is a prototype for learning and experimentation. It does not yet provide:

- automatic dependency resolution; install required dependency packages separately
- package database tracking
- signature verification or checksum validation
- SHA-256 comparison against trusted digests
- conflict handling for existing files
- hardened production-grade security guarantees

Use it only with packages you are comfortable inspecting and testing in an isolated environment. Because this runs in user space, it can still be dangerous if a package or mirror is compromised, malicious, or unexpectedly modified, even though it does not install to system-owned paths.

The project is not intended to be used as a trusted package installation mechanism for unreviewed or attacker-controlled software.

## Quick start

```bash
go build ./...
./stowy i ripgrep
ripgrep --version
```

## License

This repository does not currently include a license file. If you plan to publish or redistribute the project, add an explicit license before doing so.


*(note, this README.md was written by AI and verified by me, as I have a life. This project, however, was not written with the help of Artificial Intelligence)*
