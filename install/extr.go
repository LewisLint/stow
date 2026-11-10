package extract

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
)

// btw i use arch

func Expkg(pkgname, archpath string) error {
	// find user's ~ dir
	// put bullshit in ~/.local

	home, e := os.UserHomeDir()
	if e != nil {
		return fmt.Errorf("E: user is a homeless streetbegger.\nE: cannot find ~ dir\n")
		//lmfao
	}

	// create a directory for each individual package
	pkgdir := filepath.Join(home, ".local", "stowy", pkgname)

	if e := os.MkdirAll(pkgdir, 0o755); e != nil {
		return fmt.Errorf("I/O: could not create directory %s of perms 755\n", pkgdir)
	}

	// open pkg archive
	f, e := os.Open(archpath)
	if e != nil {
		return fmt.Errorf("I/O: could not open %s\n", archpath)
	}
	defer f.Close()

	// extract zstd data while its read
	// 'tr' will read tar entries
	// from the decompressed stream
	zr, e := zstd.NewReader(f)
	if e != nil {
		return fmt.Errorf("R: could not execute zstd reader for %v\n", f)
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
				return fmt.Errorf("I/O: could not correctly expose to path\n")
			}
			return nil
		}
		if e != nil {
			return fmt.Errorf("R: Error reading %v:\n	%v", f, e)
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
			return fmt.Errorf("I/O: could not create directory %v\n", filepath.Dir(targ))
		}

		switch hdr.Typeflag {

		case tar.TypeSymlink:
			resolv := filepath.Clean(filepath.Join(filepath.Dir(hdr.Name), hdr.Linkname))
			if filepath.IsAbs(hdr.Linkname) ||
				!filepath.IsLocal(resolv) {
				return fmt.Errorf("I/O: unsafe symlink %q > %q\n", hdr.Name, hdr.Linkname)
			}
			symlinks = append(symlinks, pendsymlink{
				name:   hdr.Name,
				target: hdr.Linkname,
			})
		case tar.TypeDir:
			// create directories as a clone from tar
			if err := os.MkdirAll(targ, 0o755); err != nil {
				return fmt.Errorf("I/O: could not create directory %v\n", filepath.Dir(targ))
			}
		case tar.TypeReg, tar.TypeRegA:
			// regulr files should create copy with its contents
			out, e := os.OpenFile(targ, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if e != nil {
				return fmt.Errorf("I/o: could not create extracted files\n")
			}
			if _, e := io.Copy(out, tr); e != nil {
				out.Close()

				return fmt.Errorf("I/O: could not copy extracted files\n")

			}

			if e := out.Close(); e != nil {
				return fmt.Errorf("I/O: could not close %v\n", out)
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
			return fmt.Errorf("E: %s already exists: %w", link, e)
		} else if !os.IsNotExist(e) {
			return fmt.Errorf("I/O: %s does not exist: %w", link, e)
		}
		if e := os.Symlink(src, link); e != nil {
			return fmt.Errorf("E: link %q: %w", link, e)
		}
	}
	return nil

}
