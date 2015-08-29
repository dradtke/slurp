package glob

import (
	"os"
	"path/filepath"
	"strings"
)

func Dir(glob string) string {
	glob = filepath.Dir(glob)
	for strings.IndexAny(glob, "*?[") > 0 {
		glob = filepath.Dir(glob)
	}
	return glob
}

func Base(glob string) string {
	base, _ := filepath.Rel(Dir(glob), glob)
	return base
}

func Match(pattern, name string) (bool, error) {

	negative := pattern != "" && pattern[0] == '!'
	if negative {
		pattern = pattern[1:]
	}

	m, err := filepath.Match(pattern, name)
	return m != negative, err
}

type MatchPair struct {
	Glob string
	Name string
}

type pattern struct {
	Glob     string
	Negative bool
}

func Excluded(patterns []pattern, name string) bool {

	for _, pattern := range patterns {
		if !pattern.Negative {
			continue
		}
		if m, _ := filepath.Match(pattern.Glob, name); m {
			return true
		}
	}

	return false
}

// If the glob contains "/**/" (backslashes on Windows), then a filepath.Walk()
// is performed starting at the directory determined by the path before it,
// and each file name is checked against the path after it.
// Otherwise, a standard filepath.Glob() is used.
func doGlob(glob string) ([]string, error) {
	const recurse = string(filepath.Separator)+"**"+string(filepath.Separator)
	if index := strings.Index(glob, recurse); index != -1 {
		var (
			g       = glob[index+4:]
			results = make([]string, 0)
		)
		if err := filepath.Walk(glob[:index], func(path string, info os.FileInfo, _ error) error {
			m, err := Match(g, info.Name())
			if err != nil {
				return err
			}
			if m {
				results = append(results, path)
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return results, nil
	}
	// Otherwise, standard globbing.
	return filepath.Glob(glob)
}

func Glob(globs ...string) (<-chan MatchPair, error) {

	//defer close(out)

	patterns := []pattern{}

	for _, glob := range globs {
		negative, err := Match(glob, "")
		if err != nil {
			return nil, err
		}

		if negative {
			glob = glob[1:]
		}

		patterns = append(patterns, pattern{glob, negative})
	}

	matches := make(chan MatchPair)
	go func() {

		seen := make(map[string]struct{})

		defer close(matches)
		for i, pattern := range patterns {

			if pattern.Negative {
				continue
			}
			//Patterns already checked and fs errors are ignored. so no error handling here.
			files, _ := doGlob(pattern.Glob)

			for _, file := range files {
				if _, seen := seen[file]; seen || Excluded(patterns[i:], file) {
					continue
				}

				seen[file] = struct{}{}
				matches <- MatchPair{pattern.Glob, file}
			}
		}
	}()

	return matches, nil
}
