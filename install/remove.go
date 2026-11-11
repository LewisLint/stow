package extract

import (
	"fmt"
	"os"
	"path/filepath"
)

func Remove(pkg string) error {
	home, e := os.UserHomeDir()
	if e != nil {
		return fmt.Errorf("E: user is homeless\nE: %v\n", e)
	}
	return rmat(home, pkg)
}

func rmat(home, pkg string) error {
	if pkg == "" ||
		pkg == "." ||
		!filepath.IsLocal(pkg) ||
		filepath.Base(pkg) != pkg {
		return fmt.Errorf("E: invalid package name %q\nSTDERR: %v\n", pkg, os.Stderr)
	}
	pkgdir := filepath.Join(home, ".local", "stowy", pkg)

	// err code
	inf, e := os.Lstat(pkgdir)
	if e != nil {
		return fmt.Errorf("E: inspect pgkdir; %q: %w\n", pkgdir, e)
	}
	if !inf.IsDir() {
		return fmt.Errorf("E: pkg path %q is not dir\n", pkgdir)
	}
	return fmt.Errorf("E: not implimented :/\n")

}
