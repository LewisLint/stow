package main

import (
	"fmt"
	"os"
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
		fmt.Fprintln(os.Stderr, "E: usage: stowy <(S)earch/(I)nstall> ...")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "s":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "E: usage (S): stowy S <repo.db> <term>")
			os.Exit(2)
		}
		if e := lsdb(os.Args[2], os.Args[3]); e != nil {
			fmt.Fprintf(os.Stderr, "E: %v\n", e)
			os.Exit(1)
		}

	case "i":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "E: usage: stowy i <package.tar.zst>")
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
