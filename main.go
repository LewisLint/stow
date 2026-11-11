package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	extract "stowy/install"
	"stowy/meta"
)

type Paths struct {
	Bin   string // compiled packages
	Lib   string // user-local libs
	Share string // data

}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "E: usage: stowy <(S)earch/(I)nstall/(L)ocal> ...")
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

		if aajdffhlkkas := dlfile(context.Background(), dburl, dbpath); aajdffhlkkas != nil {
			fmt.Fprintf(os.Stderr, "E: downloading repo db (%s): %v\n", repo, aajdffhlkkas)
			os.Exit(1)
		}

		matches, e := lsdb(dbpath, term)
		if e != nil {
			fmt.Fprintf(os.Stderr, "E: search db (%s): %v\n", repo, e)
			os.Exit(1)
		}
		for _, pkg := range matches {
			fmt.Printf("%s %s - %s\n\tarchive: %s\n", pkg.Name, pkg.Ver, pkg.Desk, pkg.Fname)
		}

	case "l":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "E: usage: stowy l <package.tar.zst>")
			os.Exit(2)
		}
		pkg, e := meta.Pkginfo(os.Args[2])
		if e != nil {
			fmt.Fprintf(os.Stderr, "E: %v\n", e)
			os.Exit(1)
		}

		fmt.Printf("Pkg name: %s\nPkg ver: %s\nPkg arch: %s\n", pkg.Name, pkg.Ver, pkg.Archbtw)
		fmt.Printf("Dependencies: \n")
		for _, dep := range pkg.Depends {
			fmt.Printf("\t%s\n", dep)
		}
		if e := extract.Expkg(pkg.Name, os.Args[2]); e != nil {
			fmt.Fprintf(os.Stderr, "E: extract error: %v\n", e)
			os.Exit(1)
		}

	case "d":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "E: usage: stowy  d <pkg>\n")
			os.Exit(2)
		}
		if e := extract.Remove(os.Args[2]); e != nil {
			fmt.Fprintf(os.Stderr, "E: remove error: %v\n", e)
			os.Exit(1)
		}

	case "i":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "E: usage: stowy i <repo> <term>")
			os.Exit(2)
		}

		repo := os.Args[2]
		term := os.Args[3]

		dburl := fmt.Sprintf("https://geo.mirror.pkgbuild.com/%s/os/x86_64/%s.db", repo, repo)
		dbpath := repodbpath(repo)

		if e := dlfile(context.Background(), dburl, dbpath); e != nil {
			fmt.Fprintf(os.Stderr, "E: download db (%s): %v\n", repo, e)
			os.Exit(1)
		}

		matches, e := lsdb(dbpath, term)
		if e != nil {
			fmt.Fprintf(os.Stderr, "E: search db (%s): %v\n", repo, e)
			os.Exit(1)
		}
		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "E: no pkg fond for %q in in %s\n", term, repo)
		}
		pkg := matches[0]
		archurl := fmt.Sprintf("https://geo.mirror.pkgbuild.com/%s/os/x86_64/%s", repo, pkg.Fname)
		archpath := filepath.Join(os.TempDir(), pkg.Fname)
		if e := dlfile(context.Background(), archurl, archpath); e != nil {
			fmt.Fprintf(os.Stderr, "E: download %s: %v", pkg.Name, e)
			os.Exit(1)
		}

		metapkg, e := meta.Pkginfo(archpath)
		if e != nil {
			fmt.Fprintf(os.Stderr, "E: parse %s pkg metadata: %v\n", pkg.Name, e)
			os.Exit(1)
		}
		if e := extract.Expkg(metapkg.Name, archpath); e != nil {
			fmt.Fprintf(os.Stderr, "E: install pkg (%s): %v\n", pkg.Name, e)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "E: unknown cmd: %q\nusage: stowy <(s)earch|(i)nstall> ...\n", os.Args[1])
		os.Exit(2)
	}

}

func cstr(shit string) bool {
	if shit == "" {
		return false
	}
	return true
}

func search(shit string) string {

	return "" // filler functions for different

	// future commands
}

func install(shit string) { // idk what to return
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
