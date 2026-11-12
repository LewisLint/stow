package extract

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Remove(pkg string) error {
	if !validPackageName(pkg) {
		return fmt.Errorf("E: invalid package name %q", pkg)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("E: could not find home directory: %w", err)
	}

	path := manifestPath(home, pkg)
	current, err := readManifest(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("E: package %q is not installed or has no install manifest", pkg)
		}
		return fmt.Errorf("E: read package manifest %q: %w", pkg, err)
	}
	if current.Package != pkg {
		return fmt.Errorf("E: package manifest %q contains mismatched package name %q", pkg, current.Package)
	}

	root := filepath.Join(home, ".local", "stowy", "root")
	info, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("E: inspect install root %q: %w", root, err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("E: install root %q is not a real directory", root)
	}

	otherPaths, err := pathsOwnedByOtherPackages(home, pkg)
	if err != nil {
		return fmt.Errorf("E: inspect installed package manifests: %w", err)
	}

	for _, file := range current.Files {
		if _, shared := otherPaths[file]; shared {
			continue
		}
		if filepath.Dir(file) == filepath.Join("usr", "bin") {
			link := filepath.Join(home, ".local", "bin", filepath.Base(file))
			info, err := os.Lstat(link)
			if err == nil {
				remove := false
				if info.Mode()&os.ModeSymlink != 0 {
					target, err := os.Readlink(link)
					if err != nil {
						return fmt.Errorf("E: inspect command link %q: %w", link, err)
					}
					remove = target == filepath.Join(root, file)
				} else if info.Mode().IsRegular() {
					launcher := commandLauncher(filepath.Base(file))
					if info.Size() == int64(len(launcher)) {
						data, err := os.ReadFile(link)
						if err != nil {
							return fmt.Errorf("E: inspect command launcher %q: %w", link, err)
						}
						remove = string(data) == launcher
					}
				}
				if remove {
					if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
						return fmt.Errorf("E: remove command link %q: %w", link, err)
					}
				}
			} else if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("E: inspect command link %q: %w", link, err)
			}
		}
	}

	files := append([]string(nil), current.Files...)
	sort.Slice(files, func(i, j int) bool {
		return strings.Count(files[i], string(filepath.Separator)) >
			strings.Count(files[j], string(filepath.Separator))
	})
	for _, file := range files {
		if _, shared := otherPaths[file]; shared {
			continue
		}
		if err := removeManagedPath(root, file); err != nil {
			return fmt.Errorf("E: remove package file %q: %w", file, err)
		}
	}

	dirs := make(map[string]struct{})
	for _, dir := range current.Dirs {
		dirs[dir] = struct{}{}
	}
	for _, file := range current.Files {
		dir := filepath.Dir(file)
		for dir != "." {
			dirs[dir] = struct{}{}
			dir = filepath.Dir(dir)
		}
	}
	dirList := make([]string, 0, len(dirs))
	for dir := range dirs {
		dirList = append(dirList, dir)
	}
	sort.Slice(dirList, func(i, j int) bool {
		return strings.Count(dirList[i], string(filepath.Separator)) >
			strings.Count(dirList[j], string(filepath.Separator))
	})
	for _, dir := range dirList {
		if _, shared := otherPaths[dir]; shared {
			continue
		}
		if err := removeManagedDir(root, dir); err != nil {
			if errors.Is(err, os.ErrNotExist) || os.IsExist(err) {
				continue
			}
			return fmt.Errorf("E: remove package directory %q: %w", dir, err)
		}
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("E: remove package manifest %q: %w", pkg, err)
	}
	return nil
}

func readManifest(path string) (packageManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return packageManifest{}, err
	}
	var manifest packageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return packageManifest{}, err
	}
	if !validPackageName(manifest.Package) {
		return packageManifest{}, fmt.Errorf("invalid package name %q", manifest.Package)
	}
	for _, path := range append(append([]string(nil), manifest.Files...), manifest.Dirs...) {
		if path == "" || path == "." || !filepath.IsLocal(path) || filepath.Clean(path) != path {
			return packageManifest{}, fmt.Errorf("unsafe path %q", path)
		}
	}
	return manifest, nil
}

func pathsOwnedByOtherPackages(home, excluding string) (map[string]struct{}, error) {
	owned := make(map[string]struct{})
	dir := filepath.Join(home, ".local", "stowy", "installed")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return owned, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		manifest, err := readManifest(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", entry.Name(), err)
		}
		if manifest.Package == excluding {
			continue
		}
		for _, file := range manifest.Files {
			owned[file] = struct{}{}
		}
		for _, dir := range manifest.Dirs {
			owned[dir] = struct{}{}
		}
	}
	return owned, nil
}

func removeManagedDir(root, rel string) error {
	if rel == "" || rel == "." || !filepath.IsLocal(rel) || filepath.Clean(rel) != rel {
		return fmt.Errorf("unsafe managed directory %q", rel)
	}
	path := filepath.Join(root, rel)
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	return removeManagedPath(root, rel)
}

func removeManagedPath(root, rel string) error {
	if rel == "" || rel == "." || !filepath.IsLocal(rel) || filepath.Clean(rel) != rel {
		return fmt.Errorf("unsafe managed path %q", rel)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	current := root
	for _, part := range parts[:len(parts)-1] {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("parent %q is not a real directory", current)
		}
	}
	return os.Remove(filepath.Join(root, rel))
}
