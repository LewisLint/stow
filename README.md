# stow

A small, experimental per-user package manager prototype written in Go.

## Aim

The goal is to install software into a user's home directory without `sudo`, keeping package files under `~/.local/stow/`. This is not intended to replace `pacman` or manage the system package database. The current prototype reads Arch Linux package archives (`.pkg.tar.zst`); running the program on another distribution would not make Arch packages compatible with that distribution.

## Current status

The command accepts one package archive. It reads `.PKGINFO`, prints basic package information, and extracts package contents into `~/.local/stow/<package-name>/`.

It currently handles regular files, directories, and relative symlinks. It rejects archive paths that are not local and unsupported archive entry types. This is still a learning prototype, not a safe general-purpose installer.

## Try it

From the repository root, run:

```sh
go run . /path/to/package.pkg.tar.zst
```

For example, on an Arch system with that package cached:

```sh
go run . /var/cache/pacman/pkg/go-tools-4:0.52.0-1-x86_64.pkg.tar.zst
```

The command prints the parsed metadata and extracts files under `~/.local/stow/`. Do not use it on packages or directories you care about yet.

## Not implemented yet

- Exposing package commands through `~/.local/bin` and other user-local paths
- Tracking installed files and uninstalling packages
- Dependency resolution and package compatibility checks
- Signature or checksum verification
- Complete protection against unsafe archive contents and existing-file conflicts
- ... Installing packages 
