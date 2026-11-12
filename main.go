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
		if err := cp(p); err != nil {
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
	dir := filepath.Dir(exe)
	path := filepath.Join(dir, "stowy")

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

	//		if err := cp(path); err != nil {
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

		if err := installPackage(pkg, "", os.Args[2], newInstallState()); err != nil {
			fmt.Fprintf(os.Stderr, "E: install %s: %v\n", pkg.Name, err)
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
			dbpath, err := repoDatabase(candidateRepo)
			if err != nil {
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

		if err := installPackage(pkg, repo, "", newInstallState()); err != nil {
			fmt.Fprintf(os.Stderr, "E: install package (%s): %v\n", pkg.Name, err)
			os.Exit(1)
		}

		lS("installed %s", pkg.Name)
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
		dbpath, err := repoDatabase(repo)
		if err != nil {
			return meta.Pkg{}, "", err
		}
		matches, err := lsdb(dbpath, dev)
		if err != nil {
			return meta.Pkg{}, "", err
		}
		for _, match := range matches {
			if packageSatisfies(match, dev) {
				return match, repo, nil
			}
		}
	}
	return meta.Pkg{}, "", fmt.Errorf("could not resolve dependency %q", dev)
}

var loadedRepoDatabases = make(map[string]string)

func repoDatabase(repo string) (string, error) {
	if path, ok := loadedRepoDatabases[repo]; ok {
		return path, nil
	}
	url := fmt.Sprintf("https://geo.mirror.pkgbuild.com/%s/os/x86_64/%s.db", repo, repo)
	path := repodbpath(repo)
	ls("repo", "downloading %s repository database", repo)
	if err := dlfile(context.Background(), url, path); err != nil {
		return "", err
	}
	loadedRepoDatabases[repo] = path
	return path, nil
}

type installState struct {
	installing map[string]bool
	installed  map[string]bool
}

func newInstallState() *installState {
	return &installState{
		installing: make(map[string]bool),
		installed:  make(map[string]bool),
	}
}

func installPackage(pkg meta.Pkg, repo, archivePath string, state *installState) error {
	if state.installed[pkg.Name] {
		return nil
	}
	if state.installing[pkg.Name] {
		return fmt.Errorf("dependency cycle involving %q", pkg.Name)
	}
	state.installing[pkg.Name] = true
	defer delete(state.installing, pkg.Name)

	if archivePath == "" {
		if repo == "" {
			return fmt.Errorf("no repository specified for package %q", pkg.Name)
		}
		url := fmt.Sprintf("https://geo.mirror.pkgbuild.com/%s/os/x86_64/%s", repo, pkg.Fname)
		archivePath = filepath.Join(os.TempDir(), filepath.Base(pkg.Fname))
		ls("download", "fetching package archive from %s", url)
		if err := dlfile(context.Background(), url, archivePath); err != nil {
			return fmt.Errorf("download %s: %w", pkg.Name, err)
		}
	}

	ls("meta", "parsing %s package metadata", pkg.Name)
	packageInfo, err := meta.Pkginfo(archivePath)
	if err != nil {
		return fmt.Errorf("parse %s package metadata: %w", pkg.Name, err)
	}
	if pkg.Name != "" && packageInfo.Name != pkg.Name {
		return fmt.Errorf("archive package name %q does not match repository package %q", packageInfo.Name, pkg.Name)
	}

	fmt.Printf("Package %s depends on:\n", packageInfo.Name)
	for _, requirement := range packageInfo.Depends {
		fmt.Printf("\t%s\n", requirement)
		dependency, dependencyRepo, err := depends(requirement)
		if err != nil {
			return fmt.Errorf("resolve dependency %q of %s: %w", requirement, packageInfo.Name, err)
		}
		if err := installPackage(dependency, dependencyRepo, "", state); err != nil {
			return fmt.Errorf("install dependency %q of %s: %w", requirement, packageInfo.Name, err)
		}
	}

	ls("install", "installing %s %s", packageInfo.Name, packageInfo.Ver)
	if err := extract.Expkg(packageInfo.Name, archivePath); err != nil {
		return fmt.Errorf("extract package %s: %w", packageInfo.Name, err)
	}
	state.installed[pkg.Name] = true
	return nil
}
func cp(path string) error {
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
