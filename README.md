# stowy

`stowy` is an experimental, user-local package manager prototype written in Go. It is designed to explore how to install Arch Linux package archives into a home-directory layout without requiring `sudo`.

The project downloads package metadata from Arch repositories, searches for packages, installs a selected archive into `~/.local/stowy/root`, and exposes executables from `usr/bin` as symlinks in `~/.local/bin`.

This project is intentionally a prototype and is not a replacement for `pacman` or a hardened production package manager.

## Features

- Search Arch repositories by package name or description
- Download and inspect package metadata from `.db` files
- Install local `.pkg.tar.zst` archives into a user-local prefix
- Extract regular files, directories, and relative symlinks safely
- Expose installed binaries from `usr/bin` into `~/.local/bin`
- Remove installed package links and extracted files
- Reject unsafe archive paths and unsupported archive entries

## Project layout

- `main.go` — command handling and CLI entry point
- `meta.go` / `meta/meta.go` — package metadata parsing
- `install/extr.go` — archive extraction and binary exposure
- `repo.go` — repository database download helpers
- `search.go` — repository search logic

## Install

From the repository root:

```bash
go build ./...
```

This produces a binary named `stowy` in the current directory.

## Usage

### Search a repository

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
./stowy i <repo> <term>
```

Example:

```bash
./stowy i extra fd
```

This downloads the package archive, installs it under a user-local prefix, and creates symlinks for executables in `~/.local/bin`.

### Install a local package archive

```bash
./stowy l <package.tar.zst>
```

Example:

```bash
./stowy l /var/cache/pacman/pkg/ripgrep-14.1.0-1-x86_64.pkg.tar.zst
```

### Remove an installed package

```bash
./stowy d <package>
```

Example:

```bash
./stowy d ripgrep
```

## Installation target

Packages are extracted under:

```bash
~/.local/stowy/root
```

Executable links are exposed in:

```bash
~/.local/bin
```

## Safety and limitations

This is a prototype for learning and experimentation. It does not yet provide:

- dependency resolution
- package database tracking
- signature or checksum verification
- conflict handling for existing files
- complete uninstall management
- a hardened or production-safe installation model

Use it only with packages you are comfortable inspecting and testing in an isolated environment.

## Why this exists

The goal is to explore the mechanics of user-local package installation patterns similar to those used in Linux distributions: read package metadata, extract files into a managed prefix, and make commands available without altering system-wide packages.

## License

This project does not currently declare a license file. If you plan to distribute or reuse it, check the repository state and add an explicit license before publication.
