# stowy

An experimental, user-local package manager prototype written in Go.

`stowy` is a small project for exploring how to install Arch Linux package archives into a home-directory layout without `sudo`. The idea is to unpack packages under `~/.local/stowy/` and expose executables from `usr/bin` into `~/.local/bin`.

This is intentionally a prototype, not a replacement for `pacman` or a hardened package manager. It is best treated as a learning tool and a sandbox for experimenting with package extraction and user-local installation patterns.

## What it does today

- Reads package metadata from an Arch package archive's `.PKGINFO` section
- Prints the package name, version, architecture, and declared dependencies
- Extracts archive contents into `~/.local/stowy/<package-name>/`
- Handles regular files, directories, and relative symlinks
- Exposes runnable files from `usr/bin` as symlinks in `~/.local/bin`
- Rejects unsafe archive paths and unsupported entry types

## Example usage

From the repository root:

```sh
go run . /path/to/package.pkg.tar.zst
```

For example, on an Arch system with a cached package archive:

```sh
go run . /var/cache/pacman/pkg/go-tools-4:0.52.0-1-x86_64.pkg.tar.zst
```

The tool will print the package metadata and then extract the package contents under `~/.local/stowy/`.

## Current limitations

This project is still intentionally incomplete. At the moment it does not provide:

- dependency resolution
- uninstall support
- package database tracking
- signature or checksum verification
- conflict handling for already-installed files
- a safe or general-purpose installation model for production use

## Project status

The code is a prototype for experimenting with:

- parsing Arch package metadata
- reading `.pkg.tar.zst` archives
- extracting files into a user-local prefix
- exposing binaries from package contents into a user `bin` directory

Use it only with packages you are comfortable inspecting or testing in an isolated environment.
