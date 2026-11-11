package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"stowy/meta"
	"strings"
)

func lsdb(dbpath, query string) error {
	db, e := os.Open(dbpath)
	if e != nil {
		return e
	} // error handling
	defer db.Close() // close if shit happens
	gz, e := gzip.NewReader(db)
	if e != nil {
		return e
	}
	defer gz.Close()

	arch := tar.NewReader(gz)

	for {
		entry, err := arch.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if !strings.HasSuffix(entry.Name, "/desc") {
			continue
		}

		data, err := io.ReadAll(arch)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %d bytes\n", entry.Name, len(data))

		pkg := meta.Parsen(string(data))
		need := strings.ToLower(query)
		if !strings.Contains(strings.ToLower(pkg.Name), need) &&
			!strings.Contains(strings.ToLower(pkg.Desk), need) {
			continue
		}
		fmt.Printf("%s %s - %s\n", pkg.Name, pkg.Ver, pkg.Desk)
	}
}
