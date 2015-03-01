//Package http provides helpful http functions.
package http

import (
	"fmt"
	"mime"
	"net/http"
	"path"

	"github.com/omeid/slurp"
)

func name(url string, response *http.Response) string {

	_, params, err := mime.ParseMediaType(response.Header.Get("Content-Disposition"))

	name, ok := params["filename"]
	if !ok || err != nil {
		name = path.Base(url)
	}

	return name
}

func Get(url string) (slurp.File, error) {

	file := slurp.File{Cwd: "", Dir: ""}

	resp, err := http.Get(url)
	if err != nil {
		return file, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		return file, fmt.Errorf("%s (%s)", resp.Status, url)
	}

	file.Reader = resp.Body

	name := name(url, resp)

	file.FileInfo.SetName(name)
	file.FileInfo.SetSize(resp.ContentLength)

	return file, nil
}
