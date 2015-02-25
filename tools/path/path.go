package path

import (
	"strings"

	"github.com/omeid/slurp"
)

func ReplaceExt(f slurp.File, Old string, New string) (slurp.File, error) {

	path := strings.TrimSuffix(f.Path, Old) + New

	s, err := f.Stat()
	if err != nil {
		return f, err
	}

	stat := slurp.FileInfoFrom(s)
	stat.SetName(path)
	f.Path = path
	f.SetStat(stat)

	return f, nil
}
