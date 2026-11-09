package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <repository.db>\n", os.Args[0])
		os.Exit(2)
	}

	if e := lsdb(os.Args[1]); e != nil {
		fmt.Printf("%v\n", e)
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
