package extract

import (
	"archive/tar"
	"bytes"
	"os"
	"os/exec"
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
	if data, err := os.ReadFile(command); err != nil || string(data) != commandLauncher("demo") {
		t.Fatalf("command launcher = %q, %v", data, err)
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

func TestCommandLauncherUsesUserLocalLibraries(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LD_LIBRARY_PATH", "/existing")
	archive := filepath.Join(t.TempDir(), "demo.pkg.tar.zst")
	writeTestPackage(t, archive, map[string]string{
		"usr/bin/demo": "#!/bin/sh\nprintf '%s\\n' \"$LD_LIBRARY_PATH\" \"$1\"\n",
	})
	commandPath := filepath.Join(home, ".local", "bin", "demo")
	if err := os.MkdirAll(filepath.Dir(commandPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, ".local", "stowy", "root", "usr/bin/demo"), commandPath); err != nil {
		t.Fatal(err)
	}
	if err := Expkg("demo", archive); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(commandPath, "hello")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run command launcher: %v: %s", err, output)
	}
	expected := filepath.Join(home, ".local", "stowy", "root", "usr", "lib") + ":/existing\nhello\n"
	if string(output) != expected {
		t.Fatalf("command output = %q, want %q", output, expected)
	}
}

func TestExpkgCanBeRepeatedWithTheSameSymlink(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	archive := filepath.Join(t.TempDir(), "neovim.pkg.tar.zst")
	writeSymlinkTestPackage(t, archive, "usr/share/nvim/runtime/parser", "../../../lib/tree_sitter")

	for range 2 {
		if err := Expkg("neovim", archive); err != nil {
			t.Fatalf("Expkg failed: %v", err)
		}
	}

	link := filepath.Join(home, ".local", "stowy", "root", "usr/share/nvim/runtime/parser")
	if target, err := os.Readlink(link); err != nil || target != "../../../lib/tree_sitter" {
		t.Fatalf("symlink target = %q, %v", target, err)
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

func writeSymlinkTestPackage(t *testing.T, path, name, target string) {
	t.Helper()
	var compressed bytes.Buffer
	zw, err := zstd.NewWriter(&compressed)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(zw)
	if err := tw.WriteHeader(&tar.Header{
		Name:     name,
		Linkname: target,
		Mode:     0o777,
		Typeflag: tar.TypeSymlink,
	}); err != nil {
		t.Fatal(err)
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
