package meta

import (
	"reflect"
	"testing"
)

func TestParsepkg(t *testing.T) {
	pkg := Parsepkg(Read("../testdata/go-tools.PKGINFO"))
	t.Logf("%+v", pkg)
}

func TestParsen(t *testing.T) {
	pkg := Parsen(Read("../testdata/sample.desc"))
	want := "stowy-demo-1.0-1-any.pkg.tar.zst"
	if pkg.Fname != want {
		t.Fatalf("Fname = %q, want %q", pkg.Fname, want)
	}
}

func TestParsenReadsRepeatedDependencyAndProviderFields(t *testing.T) {
	pkg := Parsen(`%NAME%
tree-sitter

%VERSION%
0.26.3-1

%DEPENDS%
glibc
libgcc

%PROVIDES%
libtree-sitter.so=0.26-64
tree-sitter-runtime
`)

	if pkg.Name != "tree-sitter" {
		t.Fatalf("package name = %q", pkg.Name)
	}
	if !reflect.DeepEqual(pkg.Depends, []string{"glibc", "libgcc"}) {
		t.Fatalf("dependencies = %v", pkg.Depends)
	}
	if !reflect.DeepEqual(pkg.Provides, []string{"libtree-sitter.so=0.26-64", "tree-sitter-runtime"}) {
		t.Fatalf("providers = %v", pkg.Provides)
	}
}
