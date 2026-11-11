package extract

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		// build target path in the pkgdir
		targ := filepath.Join(pkgdir, hdr.Name)

		// create parent dirs
		if e := os.MkdirAll(filepath.Dir(targ), 0o755); e != nil {
			return fmt.Errorf("I/O: could not create directory %v\nSTDERR: %w\n", filepath.Dir(targ), e)
		}

		switch hdr.Typeflag {

		case tar.TypeSymlink:
			target, err := localSymlinkTarget(hdr.Name, hdr.Linkname)
			if err != nil {
				return fmt.Errorf("I/O: unsafe symlink %q > %q: %w", hdr.Name, hdr.Linkname, err)
			}
			symlinks = append(symlinks, pendsymlink{
				name:   hdr.Name,
				target: target,
			})
		case tar.TypeDir:
			// create directories as a clone from tar
			if err := os.MkdirAll(targ, 0o755); err != nil {
				return fmt.Errorf("I/O: could not create directory %v: %w", targ, err)
			}
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
		default:
			return fmt.Errorf("E: %q has an unsupported archive entry of type %d\n", hdr.Name, hdr.Typeflag)

		}

	}

}

func expose(pkgdir string) error {
	pkgbin := filepath.Join(pkgdir, "usr", "bin")
	entries, e := os.ReadDir(pkgbin)
	if os.IsNotExist(e) {
		return fmt.Errorf("I/O: directory does not exist")
	} // damn even go devs dont have grammar
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
		if _, e := os.Lstat(link); e == nil {
			if _, e := os.Lstat(link); e == nil {
				existingTarget, readErr := os.Readlink(link)
				if readErr == nil && existingTarget == src {
					continue
				}
				return fmt.Errorf("E: %s already exists", link)
			} else if !os.IsNotExist(e) {
				return fmt.Errorf("I/O: could not inspect %s: %w", link, e)
			}
			return fmt.Errorf("E: %s already exists", link)
		} else if !os.IsNotExist(e) {
			return fmt.Errorf("I/O: could not inspect %s: %w", link, e)
		}
		if e := os.Symlink(src, link); e != nil {
			return fmt.Errorf("E: link %q: %w", link, e)
		}
		logStep("link", "exposed command %s", link)
	}
	logStep("bin", "command directory is %s; if zsh cannot find a newly installed command, run: rehash", userbin)
	return nil

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
