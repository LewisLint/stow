package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func lsdb(dbpath string) error {
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
		entry, e := arch.Next()
		if errors.Is(e, io.EOF) {
			return nil /// this means that there are no more entries
		}
		if e != nil {
			return e // because we're not using pussy ass python
		}

		if strings.HasSuffix(entry.Name, "/desc") {
			fmt.Println(entry.Name)
		}
	}

}
