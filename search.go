package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"stowy/meta"
	"strings"
	"unicode"
)

func splitDependency(dependency string) (name, operator, version string) {
	for _, candidate := range []string{">=", "<=", "=", ">", "<"} {
		if index := strings.Index(dependency, candidate); index >= 0 {
			return strings.TrimSpace(dependency[:index]), candidate, strings.TrimSpace(dependency[index+len(candidate):])
		}
	}
	return strings.TrimSpace(dependency), "", ""
}

func packageSatisfies(pkg meta.Pkg, dependency string) bool {
	name, operator, version := splitDependency(dependency)
	if strings.EqualFold(pkg.Name, name) && versionSatisfies(pkg.Ver, operator, version) {
		return true
	}
	for _, provided := range pkg.Provides {
		providedName, _, providedVersion := splitDependency(provided)
		if !strings.EqualFold(providedName, name) {
			continue
		}
		if operator == "" {
			return true
		}
		if providedVersion != "" && versionSatisfies(providedVersion, operator, version) {
			return true
		}
	}
	return false
}

func versionSatisfies(candidate, operator, required string) bool {
	if operator == "" {
		return true
	}
	comparison := compareVersions(candidate, required)
	switch operator {
	case "=":
		return comparison == 0
	case ">":
		return comparison > 0
	case ">=":
		return comparison >= 0
	case "<":
		return comparison < 0
	case "<=":
		return comparison <= 0
	default:
		return false
	}
}

func compareVersions(left, right string) int {
	leftEpoch, leftVersion := splitEpoch(left)
	rightEpoch, rightVersion := splitEpoch(right)
	if comparison := compareNumeric(leftEpoch, rightEpoch); comparison != 0 {
		return comparison
	}

	leftParts := versionParts(leftVersion)
	rightParts := versionParts(rightVersion)
	for i := 0; i < len(leftParts) && i < len(rightParts); i++ {
		leftNumber := isNumeric(leftParts[i])
		rightNumber := isNumeric(rightParts[i])
		switch {
		case leftNumber && rightNumber:
			if comparison := compareNumeric(leftParts[i], rightParts[i]); comparison != 0 {
				return comparison
			}
		case leftNumber:
			return 1
		case rightNumber:
			return -1
		default:
			if comparison := strings.Compare(leftParts[i], rightParts[i]); comparison != 0 {
				return comparison
			}
		}
	}
	switch {
	case len(leftParts) > len(rightParts):
		return 1
	case len(leftParts) < len(rightParts):
		return -1
	default:
		return 0
	}
}

func splitEpoch(version string) (string, string) {
	epoch, rest, found := strings.Cut(version, ":")
	if !found {
		return "0", version
	}
	return epoch, rest
}

func versionParts(version string) []string {
	return strings.FieldsFunc(version, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func compareNumeric(left, right string) int {
	left = strings.TrimLeft(left, "0")
	right = strings.TrimLeft(right, "0")
	if len(left) > len(right) {
		return 1
	}
	if len(left) < len(right) {
		return -1
	}
	return strings.Compare(left, right)
}

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
	need := strings.ToLower(func() string {
		name, _, _ := splitDependency(query)
		return name
	}())

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
			!strings.Contains(strings.ToLower(pkg.Desk), need) &&
			!packageSatisfies(pkg, query) {
			continue
		}
		matches = append(matches, pkg)
	}
}
