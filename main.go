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
	if len(os.Args) != 2 {
		fmt.Printf("usage: stowy <foo.tar.zst>\n")
		os.Exit(2)
	}

	pkg, e := meta.Pkginfo(os.Args[1])
	if e != nil {
		fmt.Printf("E: %v\n", e)
		os.Exit(1)
	}

	fmt.Printf("Pkg Name: %s\nPkg Version: %s\nPkg Architecture: %s\n", pkg.Name, pkg.Ver, pkg.Archbtw)
	fmt.Printf("Dependancies:\n")
	for i := range pkg.Depends {
		fmt.Printf("	%s\n", pkg.Depends[i])
	}
	if e := extract.Expkg(pkg.Name, os.Args[1]); e != nil {
		fmt.Printf("%v", e)
		os.Exit(1)
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
