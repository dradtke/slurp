package path

import (
	"strings"

	"github.com/omeid/slurp"
)

func ReplaceExt(f slurp.File, Old string, New string) (slurp.File, error) {

	path := strings.TrimSuffix(f.Path, Old) + New
	f.Path = path
	f.FileInfo.SetName(path)

	return f, nil
}
