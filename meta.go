package main

// ignore this shit, i enslaved goimports
import (
	"os"
	"strings"
)

type Pkg struct { // pretty fucking obvious what this mf does
	Name    string // it stores data
	Fname   string
	Base    string
	Ver     string // shocking i know
	Desk    string
	Csize   int
	Isize   int
	Sha     string
	Pgp     string
	Url     string
	Lic     string
	Archbtw string
	Bdate   string
}

func read(s string) string { // read bytes from a
	// given file as a str

	txt, e := (os.ReadFile(s))
	if e != nil {
		panic(e)
	}
	dat := string(txt)
	// we'll just print the return later
	return dat
}

func parse(dat string) Pkg {
	var pkg Pkg
	field := ""

	for _, l := range strings.Split(dat, "\n") { // seperates lines by return ('\n')
		if strings.HasPrefix(l, "%") && // this checks if the line seperated above
			strings.HasSuffix(l, "%") { // starts and ends with '%'
			field = l
			continue
		}
		switch field {
		case "%NAME%":
			pkg.Name = l
		case "%VERSION%":
			pkg.Ver = l
		case "%DESC%":
			pkg.Desk = l
		}
		field = ""

	}
	return pkg // return the output of something?
}

// idk wtf im doing yo.
// lol
