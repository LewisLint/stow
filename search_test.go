package main

import (
	"testing"

	"stowy/meta"
)

func TestPackageSatisfiesDependencies(t *testing.T) {
	pkg := meta.Pkg{
		Name:     "tree-sitter",
		Ver:      "0.26.3-1",
		Provides: []string{"libtree-sitter.so=0.26-64"},
	}

	for _, requirement := range []string{
		"tree-sitter",
		"libtree-sitter.so=0.26-64",
		"tree-sitter>=0.26.3",
	} {
		if !packageSatisfies(pkg, requirement) {
			t.Errorf("%q was not satisfied", requirement)
		}
	}
	for _, requirement := range []string{
		"tree-sitter>=0.27",
		"libtree-sitter.so=0.25-64",
		"libvterm",
	} {
		if packageSatisfies(pkg, requirement) {
			t.Errorf("%q was unexpectedly satisfied", requirement)
		}
	}
}
