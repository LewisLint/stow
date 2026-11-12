package extract

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestRemoveDeletesInstalledPackageFilesAndLinks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	archive := filepath.Join(t.TempDir(), "demo.pkg.tar.zst")
	writeTestPackage(t, archive, map[string]string{
		"usr/bin/demo":       "executable",
		"usr/share/demo.txt": "package data",
	})

	if err := Expkg("demo", archive); err != nil {
		t.Fatal(err)
	}

	root := filepath.Join(home, ".local", "stowy", "root")
	command := filepath.Join(home, ".local", "bin", "demo")
	if target, err := os.Readlink(command); err != nil || target != filepath.Join(root, "usr/bin/demo") {
		t.Fatalf("command link = %q, %v", target, err)
	}

	if err := Remove("demo"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(root, "usr/bin/demo"),
		filepath.Join(root, "usr/share/demo.txt"),
		command,
		manifestPath(home, "demo"),
	} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Errorf("expected %q to be removed; lstat error = %v", path, err)
		}
	}
}

func TestRemoveCleansPackageWhenCommandExposureFails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	archive := filepath.Join(t.TempDir(), "demo.pkg.tar.zst")
	writeTestPackage(t, archive, map[string]string{"usr/bin/demo": "executable"})
	command := filepath.Join(home, ".local", "bin", "demo")
	if err := os.MkdirAll(filepath.Dir(command), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(command, []byte("existing command"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Expkg("demo", archive); err == nil {
		t.Fatal("Expkg succeeded despite an existing command")
	}
	if err := Remove("demo"); err != nil {
		t.Fatalf("Remove could not clean the partially installed package: %v", err)
	}
	if data, err := os.ReadFile(command); err != nil || string(data) != "existing command" {
		t.Fatalf("existing command changed: %q, %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(home, ".local", "stowy", "root", "usr/bin/demo")); !os.IsNotExist(err) {
		t.Fatalf("partially installed package file was not removed: %v", err)
	}
}

func TestRemoveKeepsFilesOwnedByAnotherPackage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	archive := filepath.Join(t.TempDir(), "demo.pkg.tar.zst")
	writeTestPackage(t, archive, map[string]string{"usr/share/shared.txt": "shared"})

	if err := Expkg("first", archive); err != nil {
		t.Fatal(err)
	}
	if err := Expkg("second", archive); err != nil {
		t.Fatal(err)
	}
	if err := Remove("first"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".local", "stowy", "root", "usr/share/shared.txt")); err != nil {
		t.Fatalf("shared file was removed: %v", err)
	}
}

func TestRemoveKeepsNonEmptyDirectories(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	archive := filepath.Join(t.TempDir(), "demo.pkg.tar.zst")
	writeTestPackage(t, archive, map[string]string{"usr/share/demo.txt": "package data"})
	if err := Expkg("demo", archive); err != nil {
		t.Fatal(err)
	}

	personalFile := filepath.Join(home, ".local", "stowy", "root", "usr/share/personal.txt")
	if err := os.WriteFile(personalFile, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := Remove("demo"); err != nil {
		t.Fatalf("Remove failed with a non-empty package directory: %v", err)
	}
	if data, err := os.ReadFile(personalFile); err != nil || string(data) != "keep me" {
		t.Fatalf("personal file changed: %q, %v", data, err)
	}
	if _, err := os.Stat(manifestPath(home, "demo")); !os.IsNotExist(err) {
		t.Fatalf("package manifest was not removed: %v", err)
	}
}

func TestRemoveRejectsInvalidPackageName(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../outside", "nested/name"} {
		if err := Remove(name); err == nil {
			t.Errorf("Remove(%q) unexpectedly succeeded", name)
		}
	}
}

func TestRemoveDoesNotFollowReplacedParentSymlink(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	archive := filepath.Join(t.TempDir(), "demo.pkg.tar.zst")
	writeTestPackage(t, archive, map[string]string{"usr/share/demo.txt": "package data"})
	if err := Expkg("demo", archive); err != nil {
		t.Fatal(err)
	}

	root := filepath.Join(home, ".local", "stowy", "root")
	share := filepath.Join(root, "usr/share")
	if err := os.Remove(filepath.Join(share, "demo.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(share); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "demo.txt")
	if err := os.WriteFile(outsideFile, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, share); err != nil {
		t.Fatal(err)
	}

	if err := Remove("demo"); err == nil {
		t.Fatal("Remove succeeded through a replaced parent symlink")
	}
	if data, err := os.ReadFile(outsideFile); err != nil || string(data) != "keep me" {
		t.Fatalf("outside file changed: %q, %v", data, err)
	}
}

func writeTestPackage(t *testing.T, path string, files map[string]string) {
	t.Helper()
	var compressed bytes.Buffer
	zw, err := zstd.NewWriter(&compressed)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(zw)
	for name, contents := range files {
		mode := int64(0o644)
		if name == "usr/bin/demo" {
			mode = 0o755
		}
		if err := tw.WriteHeader(&tar.Header{
			Name:     name,
			Mode:     mode,
			Size:     int64(len(contents)),
			Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, compressed.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
}
