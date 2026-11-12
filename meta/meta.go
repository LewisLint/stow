package meta

// ignore this shit, i enslaved goimports
import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
)

type Pkg struct { // pretty fucking obvious what this mf does
	Depends  []string
	Provides []string
	Name     string // it stores data
	Fname    string
	Base     string
	Ver      string // shocking i know
	Desk     string
	Csize    int
	Isize    int
	Sha      string
	Pgp      string
	Url      string
	Lic      string
	Archbtw  string
	Bdate    string
}

func Read(s string) string { // read bytes from a
	// given file as a str

	txt, e := (os.ReadFile(s))
	if e != nil {
		panic(e)
	}
	dat := string(txt)
	// we'll just print the return later
	return dat
}

func Parsen(dat string) Pkg {
	var pkg Pkg
	field := ""

	for _, line := range strings.Split(dat, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "%") && strings.HasSuffix(line, "%") {
			field = line
			continue
		}
		if line == "" {
			continue
		}
		switch field {
		case "%NAME%":
			pkg.Name = line
		case "%VERSION%":
			pkg.Ver = line
		case "%DESC%":
			pkg.Desk = line
		case "%FILENAME%":
			pkg.Fname = line
		case "%DEPENDS%":
			pkg.Depends = append(pkg.Depends, line)
		case "%PROVIDES%":
			pkg.Provides = append(pkg.Provides, line)
		}
	}
	return pkg // return the output of something?
}

// idk wtf im doing yo.
// lol

func Parsepkg(dat string) Pkg { // parsed package info
	var pkg Pkg // variable  of the same type as the Pkg struct
	// allows us to edit the struct easily
	for _, l := range strings.Split(dat, "\n") { // seperaets by line
		key, val, ok := strings.Cut(l, "=") // splits info by '='
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch key {
		case "pkgname":
			pkg.Name = val
		case "pkgver":
			pkg.Ver = val
		case "pkgdesc":
			pkg.Desk = val
		case "arch":
			pkg.Archbtw = val
		case "depend":
			pkg.Depends = append(pkg.Depends, val)
		}
	}
	return pkg
}

func Pkginfo(path string) (Pkg, error) {
	archfile, e := os.Open(path)
	if e != nil {
		return Pkg{}, e
	}
	defer archfile.Close()

	zstdr, e := zstd.NewReader(archfile)
	if e != nil {
		return Pkg{}, e
	}
	defer zstdr.Close()

	tr := tar.NewReader(zstdr)
	for {
		header, e := tr.Next()
		if e == io.EOF {
			return Pkg{}, fmt.Errorf("%q: .PKGINFO not foud", path)
		}
		if e != nil {
			return Pkg{}, e
		}
		if header.Name != ".PKGINFO" {
			continue
		}
		dat, e := io.ReadAll(tr)
		if e != nil {
			return Pkg{}, e
		}
		return Parsepkg(string(dat)), nil

	}
}
