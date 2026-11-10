package meta

import "testing"

func TestParsepkg(t *testing.T) {
	pkg := Parsepkg(Read("../testdata/go-tools.PKGINFO"))
	t.Logf("%+v", pkg)
}
