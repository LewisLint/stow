package meta

import "testing"

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
