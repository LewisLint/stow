package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	extract "stowy/install"
	"stowy/meta"
	"strings"
)

type Paths struct {
	Bin   string // compiled packages
	Lib   string // user-local libs
	Share string // data
}

const (
	colorReset = "\033[0m"
	colorCyan  = "\033[36m"
	colorGreen = "\033[32m"
	colorWarn  = "\033[33m"
	colorBold  = "\033[1m"
)

func main() {
	// this motherfucker gets to be the first line because it  pissed  me off

	home, _ := os.UserHomeDir()
	paths := Paths{
		Bin:   filepath.Join(home, ".local", "stowy", "usr", "bin"),
		Lib:   filepath.Join(home, ".local", "stowy", "usr", "lib"),
		Share: filepath.Join(home, ".local", "stowy", "usr", "share"),
	}

	for _, p := range []string{paths.Bin, paths.Lib, paths.Share} {
		if err := checkPaths(p); err != nil {
			fmt.Fprintf(os.Stderr, "E: %v\n", err)
			os.Exit(1)
		}
	}

	// symlink Stowy* to ~/.local/bin

	exe, err := os.Executable()
	if err != nil {
		panic(err) // im too fucking lazy to make a error message for this
	}
	usrbin := filepath.Join(home, ".local", "bin", "stowy")
	fmt.Printf("%v\n", usrbin)
	dir := filepath.Dir(exe)
	path := filepath.Join(dir, "stowy")
	fmt.Printf("%v\n\n", path)

	_, e := os.Lstat(usrbin)
	if e == nil {
	} else if os.IsNotExist(e) {
		fmt.Printf("Symlink %v > %v does not exist\nCreating symlink...\n", path, usrbin)
		err := os.Symlink(path, usrbin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "E: failed to create symlink: %v\n", err)
			fmt.Printf("\nContinuing anyways...")
		}
	} else {
		fmt.Fprintf(os.Stderr, "E: an error occured when symlinking stowy to ~/.local/bin/\ncontinuing anyways...")
	}

	//		if err := checkPaths(path); err != nil {
	//		fmt.Fprintf(os.Stderr, "E: could not start because Stowy is a bitch about ~/.local/stowy/usr/bin/: %v\n", err)
	//		os.Exit(1)
	//	}
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "E: usage: stowy <search|install|local|uninstall> ...")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "s":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "E: usage (S): stowy S <repo> <term>")
			os.Exit(2)
		}
		repo := os.Args[2]
		term := os.Args[3]
		dburl := fmt.Sprintf("https://geo.mirror.pkgbuild.com/%s/os/x86_64/%s.db", repo, repo)
		dbpath := repodbpath(repo)

		ls("repo", "downloading %s database", repo)
		if err := dlfile(context.Background(), dburl, dbpath); err != nil {
			fmt.Fprintf(os.Stderr, "E: downloading repo db (%s): %v\n", repo, err)
			os.Exit(1)
		}

		matches, err := lsdb(dbpath, term)
		if err != nil {
			fmt.Fprintf(os.Stderr, "E: search db (%s): %v\n", repo, err)
			os.Exit(1)
		}
		if len(matches) == 0 {
			lw("no package matches for %q in %s", term, repo)
			os.Exit(0)
		}
		pplist(matches)

	case "l":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "E: usage: stowy l <package.tar.zst>")
			os.Exit(2)
		}
		ls("archive", "reading local package metadata from %s", os.Args[2])
		pkg, err := meta.Pkginfo(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "E: %v\n", err)
			os.Exit(1)
		}

		ppkginf(pkg)
		ls("install", "installing %s %s", pkg.Name, pkg.Ver)
		if err := extract.Expkg(pkg.Name, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "E: extract error: %v\n", err)
			os.Exit(1)
		}
		lS("installed %s", pkg.Name)

	case "d", "delete", "remove", "uninstall":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "E: usage: stowy uninstall <package>")
			os.Exit(2)
		}
		if err := extract.Remove(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "E: remove error: %v\n", err)
			os.Exit(1)
		}
		lS("removed %s", os.Args[2])
	case "i":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "E: usage: stowy i <package>")
			os.Exit(2)
		}

		term := os.Args[2]
		repos := []string{"core", "extra", "multilib"}

		var pkg meta.Pkg
		repo := ""

		for _, candidateRepo := range repos {
			dburl := fmt.Sprintf(
				"https://geo.mirror.pkgbuild.com/%s/os/x86_64/%s.db",
				candidateRepo,
				candidateRepo,
			)
			dbpath := repodbpath(candidateRepo)

			ls("repo", "downloading %s repository database", candidateRepo)
			if err := dlfile(context.Background(), dburl, dbpath); err != nil {
				fmt.Fprintf(os.Stderr, "E: download database (%s): %v\n", candidateRepo, err)
				os.Exit(1)
			}

			matches, err := lsdb(dbpath, term)
			if err != nil {
				fmt.Fprintf(os.Stderr, "E: search database (%s): %v\n", candidateRepo, err)
				os.Exit(1)
			}

			for _, match := range matches {
				if strings.EqualFold(term, match.Name) {
					pkg = match
					repo = candidateRepo
					break
				}
			}

			if repo != "" {
				break
			}
		}

		if repo == "" {
			fmt.Fprintf(os.Stderr, "E: package %q not found in %s\n", term, strings.Join(repos, ", "))
			os.Exit(1)
		}

		ls("select", "selected %s %s from %s", pkg.Name, pkg.Ver, repo)

		archurl := fmt.Sprintf(
			"https://geo.mirror.pkgbuild.com/%s/os/x86_64/%s",
			repo,
			pkg.Fname,
		)
		archpath := filepath.Join(os.TempDir(), pkg.Fname)

		ls("download", "fetching package archive from %s", archurl)
		if err := dlfile(context.Background(), archurl, archpath); err != nil {
			fmt.Fprintf(os.Stderr, "E: download %s: %v\n", pkg.Name, err)
			os.Exit(1)
		}

		ls("meta", "parsing package metadata")
		metapkg, err := meta.Pkginfo(archpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "E: parse %s package metadata: %v\n", pkg.Name, err)
			os.Exit(1)
		}

		ls("install", "installing %s %s", metapkg.Name, metapkg.Ver)
		if err := extract.Expkg(metapkg.Name, archpath); err != nil {
			fmt.Fprintf(os.Stderr, "E: install package (%s): %v\n", pkg.Name, err)
			os.Exit(1)
		}

		lS("installed %s", metapkg.Name)
	default:
		fmt.Fprintf(os.Stderr, "E: unknown cmd: %q\nusage: stowy <search|install|local|uninstall> ...\n", os.Args[1])
		os.Exit(2)
	}
}

func ls(label, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s%s%s %s\n", colorBold, colorCyan, label, colorReset, msg)
}

func lS(format string, args ...interface{}) {
	fmt.Printf("%s%s✔%s %s\n", colorBold, colorGreen, colorReset, fmt.Sprintf(format, args...))
}

func lw(format string, args ...interface{}) {
	fmt.Printf("%s%s!%s %s\n", colorBold, colorWarn, colorReset, fmt.Sprintf(format, args...))
}

func ppkginf(pkg meta.Pkg) {
	fmt.Printf("\n%sPackage%s %s %s\n", colorBold, colorReset, pkg.Name, pkg.Ver)
	fmt.Printf("  arch: %s\n", pkg.Archbtw)
	if len(pkg.Depends) > 0 {
		fmt.Printf("  depends:\n")
		for _, dep := range pkg.Depends {
			fmt.Printf("    - %s\n", dep)
		}
	} else {
		fmt.Printf("  depends: none\n")
	}
	fmt.Println()
}

func pplist(matches []meta.Pkg) {
	fmt.Printf("\n%sResults%s\n", colorBold, colorReset)
	for _, pkg := range matches {
		fmt.Printf("  %s %s - %s\n", pkg.Name, pkg.Ver, pkg.Desk)
		fmt.Printf("    archive: %s\n", pkg.Fname)
	}
	fmt.Println()
}

func list(dir string) []string {
	// returns the installed packages
	o, e := os.ReadDir(dir)
	if e != nil {
		panic(e)
	}
	var obj []string
	for i := range o {
		obj = append(obj, o[i].Name())
	}

	return obj
}

func depends(dev string) (meta.Pkg, string, error) {
	repos := []string{"core", "extra", "multilib"}

	for _, repo := range repos {
		dburl := fmt.Sprintf("https://geo.mirror.pkgbuild.com/%s/os/x86_64/%s.db", repo, repo)
		dbpath := repodbpath(repo)

		if e := dlfile(context.Background(), dburl, dbpath); e != nil {
			return meta.Pkg{}, "", e
		}
		matches, e := lsdb(dbpath, dev)
		if e != nil {
			return meta.Pkg{}, "", e
		}
		for _, match := range matches {
			if !strings.EqualFold(dev, match.Name) {
				continue
			}

		}
	}
	return meta.Pkg{}, "", fmt.Errorf("could not resolve dependency %q", dev)
}
func checkPaths(path string) error {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s exists but isn't a directory", path)
	}

	return nil
}
