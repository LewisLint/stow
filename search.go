package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"stowy/meta"
	"strings"
)

func lsdb(dbpath, query string) ([]meta.Pkg, error) {
	db, e := os.Open(dbpath)
	if e != nil {
		return nil, e
	} // error handling
	defer db.Close() // close if shit happens
	gz, e := gzip.NewReader(db)
	if e != nil {
		return nil, e
	}
	defer gz.Close()

	arch := tar.NewReader(gz)
	var matches []meta.Pkg
	need := strings.ToLower(query)

	for {
		entry, err := arch.Next()
		if errors.Is(err, io.EOF) {
			return matches, nil
		}
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(entry.Name, "/desc") {
			continue
		}

		data, err := io.ReadAll(arch)
		if err != nil {
			return nil, err
		}
		pkg := meta.Parsen(string(data))
		if !strings.Contains(strings.ToLower(pkg.Name), need) &&
			!strings.Contains(strings.ToLower(pkg.Desk), need) {
			continue
		}
		matches = append(matches, pkg)
	}
}
