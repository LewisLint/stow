package extract

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// btw i use arch
func Expkg(pkgname, archpath string) error {
	// find user's ~ dir
	home, e := os.UserHomeDir()
	if e != nil {
		return fmt.Errorf("E: user is a homeless streetbegger.\nE: cannot find ~ dir\nSTDERR: %w\n", e)
	}
	if !validPackageName(pkgname) {
		return fmt.Errorf("E: invalid package name %q", pkgname)
	}

	// create a directory for each individual package
	pkgdir := filepath.Join(home, ".local", "stowy", "root")
	logStep("extract", "extracting package archive %s", archpath)
	logStep("extract", "writing package files to %s", pkgdir)

	if e := os.MkdirAll(pkgdir, 0o755); e != nil {
		return fmt.Errorf("I/O: could not create directory %s of perms 755\nSTDERR: %w\n", pkgdir, e)
	}

	// open pkg archive
	f, e := os.Open(archpath)
	if e != nil {
		return fmt.Errorf("I/O: could not open %s\nSTDERR: %w\n", archpath, e)
	}
	defer f.Close()

	// extract zstd data while its read.
	// 'tr' will read tar entries
	// from the decompressed stream
	zr, e := zstd.NewReader(f)
	if e != nil {
		return fmt.Errorf("R: could not execute zstd reader for %v\nSTDERR: %w\n", f, e)
	}
	defer zr.Close()

	// create tar reader
	tr := tar.NewReader(zr) // pacman packages compress via tar AND zstd.

	type pendsymlink struct {
		name   string
		target string
	}

	var symlinks []pendsymlink
	var files []string
	var dirs []string

	// read every package entry
	for {
		hdr, e := tr.Next()

		if e == io.EOF {
			// we dont need to create an error message
			// this just stops the loop if we reached the end of the archive
			for _, l := range symlinks {
				lp := filepath.Join(pkgdir, l.name)
				if e := os.Symlink(l.target, lp); e != nil {
					return fmt.Errorf("I/O: create symlink %q: %w", lp, e)
				}
			}
			if e := writeManifest(home, pkgname, files, dirs); e != nil {
				return fmt.Errorf("I/O: could not record installed package %q: %w", pkgname, e)
			}
			if e := expose(pkgdir); e != nil {
				return fmt.Errorf("I/O: could not correctly expose to path\nSTDERR: %w\n", e)
			}
			logStep("done", "package extraction complete")
			return nil
		}
		if e != nil {
			return fmt.Errorf("R: Error reading %v:\nSTDERR: %w", f, e)
		}

		// skip pkg meta
		if hdr.Name == ".PKGINFO" {
			// we already previously read package metadata seperately
			continue

		}

		if !filepath.IsLocal(hdr.Name) {
			return fmt.Errorf("P: unsafe archive path %q\n", hdr.Name)
		}
		name := filepath.Clean(hdr.Name)
		if name == "." && hdr.Typeflag == tar.TypeDir {
			continue
		}
		// build target path in the pkgdir
		targ := filepath.Join(pkgdir, name)

		// create parent dirs
		if e := os.MkdirAll(filepath.Dir(targ), 0o755); e != nil {
			return fmt.Errorf("I/O: could not create directory %v\nSTDERR: %w\n", filepath.Dir(targ), e)
		}

		switch hdr.Typeflag {

		case tar.TypeSymlink:
			target, err := localSymlinkTarget(name, hdr.Linkname)
			if err != nil {
				return fmt.Errorf("I/O: unsafe symlink %q > %q: %w", name, hdr.Linkname, err)
			}
			symlinks = append(symlinks, pendsymlink{
				name:   name,
				target: target,
			})
			files = append(files, name)
		case tar.TypeDir:
			// create directories as a clone from tar
			if err := os.MkdirAll(targ, 0o755); err != nil {
				return fmt.Errorf("I/O: could not create directory %v: %w", targ, err)
			}
			dirs = append(dirs, name)
		case tar.TypeReg, tar.TypeRegA:
			// regulr files should create copy with its contents
			out, e := os.OpenFile(targ, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if e != nil {
				return fmt.Errorf("I/o: could not create extracted files\nSTDERR: %w", e)
			}
			if _, e := io.Copy(out, tr); e != nil {
				out.Close()

				return fmt.Errorf("I/O: could not copy extracted files\nSTDERR: %w\n", e)

			}

			if e := out.Close(); e != nil {
				return fmt.Errorf("I/O: could not close %v\nSTDERR: %w", out, e)
			}
			files = append(files, name)
		default:
			return fmt.Errorf("E: %q has an unsupported archive entry of type %d\n", hdr.Name, hdr.Typeflag)

		}

	}

}

type packageManifest struct {
	Package string   `json:"package"`
	Files   []string `json:"files"`
	Dirs    []string `json:"dirs"`
}

func validPackageName(pkg string) bool {
	return pkg != "" && pkg != "." && pkg != ".." &&
		filepath.IsLocal(pkg) && filepath.Base(pkg) == pkg
}

func manifestPath(home, pkg string) string {
	return filepath.Join(home, ".local", "stowy", "installed", pkg+".json")
}

func writeManifest(home, pkg string, files, dirs []string) error {
	sort.Strings(files)
	sort.Strings(dirs)
	data, err := json.Marshal(packageManifest{Package: pkg, Files: files, Dirs: dirs})
	if err != nil {
		return err
	}
	path := manifestPath(home, pkg)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".manifest-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

func expose(pkgdir string) error {
	pkgbin := filepath.Join(pkgdir, "usr", "bin")
	entries, e := os.ReadDir(pkgbin)
	// this block is an absolute motherfucker
	if os.IsNotExist(e) {
		logStep("expose", "package has no usr/bin dir")
		return nil
	}
	// do you  KNOW how many times i got  "could not bind to usr/bin?"
	// burn in hell

	if e != nil {

		return fmt.Errorf("E: read pkg commands: %w", e)
	}
	home, _ := os.UserHomeDir()
	userbin := filepath.Join(home, ".local", "bin")
	if e := os.MkdirAll(userbin, 0o755); e != nil {
		return fmt.Errorf("I/O: could not create cmd dir: %w", e)
	}

	logStep("link", "creating command links in %s", userbin)
	for _, i := range entries {
		src := filepath.Join(pkgbin, i.Name())
		info, e := os.Stat(src)
		if e != nil {
			return fmt.Errorf("E: failed to find %q: %v\n", src, e)
		}
		if !info.Mode().IsRegular() ||
			info.Mode().Perm()&0o111 == 0 {
			continue
		}
		link := filepath.Join(userbin, i.Name())
		launcher := commandLauncher(i.Name())
		existing, e := os.Lstat(link)
		if e == nil {
			if existing.Mode()&os.ModeSymlink != 0 {
				target, readErr := os.Readlink(link)
				if readErr != nil {
					return fmt.Errorf("I/O: could not inspect %s: %w", link, readErr)
				}
				if target != src {
					return fmt.Errorf("E: %s already exists", link)
				}
			} else if existing.Mode().IsRegular() && existing.Size() == int64(len(launcher)) {
				data, readErr := os.ReadFile(link)
				if readErr != nil {
					return fmt.Errorf("I/O: could not inspect %s: %w", link, readErr)
				}
				if string(data) == launcher {
					continue
				}
				return fmt.Errorf("E: %s already exists", link)
			} else {
				return fmt.Errorf("E: %s already exists", link)
			}
		} else if !os.IsNotExist(e) {
			return fmt.Errorf("I/O: could not inspect %s: %w", link, e)
		}
		temp, e := os.CreateTemp(userbin, ".stowy-launcher-*")
		if e != nil {
			return fmt.Errorf("I/O: create launcher for %q: %w", link, e)
		}
		tempName := temp.Name()
		defer os.Remove(tempName)
		if e := temp.Chmod(0o755); e != nil {
			temp.Close()
			return fmt.Errorf("I/O: set launcher permissions for %q: %w", link, e)
		}
		if _, e := temp.WriteString(launcher); e != nil {
			temp.Close()
			return fmt.Errorf("I/O: write launcher for %q: %w", link, e)
		}
		if e := temp.Close(); e != nil {
			return fmt.Errorf("I/O: close launcher for %q: %w", link, e)
		}
		if e := os.Rename(tempName, link); e != nil {
			return fmt.Errorf("E: install launcher %q: %w", link, e)
		}
		logStep("link", "exposed command %s", link)
	}
	logStep("bin", "command directory is %s; if zsh cannot find a newly installed command, run: rehash", userbin)
	return nil

}

func commandLauncher(command string) string {
	return "#!/bin/sh\n" +
		"STOWY_ROOT=\"${HOME:?}/.local/stowy/root\"\n" +
		"export LD_LIBRARY_PATH=\"$STOWY_ROOT/usr/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}\"\n" +
		"exec \"$STOWY_ROOT/usr/bin/\"" + shellQuote(command) + " \"$@\"\n"
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func logStep(label, format string, args ...interface{}) {
	fmt.Printf("\033[1m\033[36m%s\033[0m %s\n", label, fmt.Sprintf(format, args...))
}

func localSymlinkTarget(name, linkname string) (string, error) {
	var target string

	if filepath.IsAbs(linkname) {
		// Treat an archive's absolute path as rooted at stowy's shared install root.
		target = filepath.Clean(strings.TrimLeft(linkname, string(filepath.Separator)))
	} else {
		target = filepath.Clean(filepath.Join(filepath.Dir(name), linkname))
	}

	if !filepath.IsLocal(target) {
		return "", fmt.Errorf("target escapes install root")
	}

	return filepath.Rel(filepath.Dir(name), target)
}
