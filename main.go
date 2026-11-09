package main

import (
	"fmt"
	"os"
)

func main() {
	//	uidb := list("uninstalled")
	//	idb := list("installed")
	if len(os.Args) < 3 {
		fmt.Printf("You must call stow with an argument.")
		return
	}
	var o []string
	arg := os.Args[1]
	cmd := arg
	switch cmd {
	case "s":
		search(os.Args[2])
	case "i":
		install(os.Args[2])
	case "l":
		fmt.Printf("(L)ist ")
		// Make sure they actually passed the u or i argument
		if len(os.Args) < 3 {
			fmt.Println("Usage: ./stow l [i|u]")
			return
		}

		// Change these from [3] to [2]
		if os.Args[2] == "u" {
			o = list("uninstalled")
			fmt.Printf("(U)ninstalled packages:\n")
		} else if os.Args[2] == "i" {
			o = list("installed")
			fmt.Printf("(I)nstalled packages:\n")
		} else {
			fmt.Printf("'l' does not take argument '%s'\n", os.Args[2])
		}
		for i := range o {
			// Also make sure to change your %f to %s here so it prints names correctly!
			fmt.Printf("%s\n", o[i])
		}

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
