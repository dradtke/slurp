package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

var (
	gopaths = strings.Split(os.Getenv("GOPATH"), ":")
	cwd     string

	timestamp = flag.Int("timestamp", 0, "Log timestamp: 1-6.")
	build     = flag.Bool("build", false, "build the current build as slurp-bin")
	install   = flag.Bool("install", false, "install current slurp.Go as slurp.PKG.")
	bare      = flag.Bool("bare", false, "Run/Install the slurp.go file without any other files.")
	slurpfile = flag.String("slurpfile", "slurp.go", "The file that includes the Slurp(*s.Build) function, use by -bare")

	keep = flag.Bool("keep", false, "keep the generated source under $GOPATH/src/slurp-run-*")
)

func init() {
	maxprocs := runtime.NumCPU()
	if maxprocs > 2 {
		runtime.GOMAXPROCS(maxprocs / 2)
	}
}

func main() {

	flag.Parse()

	if len(gopaths) == 0 || gopaths[0] == "" {
		log.Fatal("$GOPATH must be set.")
	}

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	path, err := generate()
	if err != nil {
		return err
	}

	//Don't forget to clean up.
	if !*keep {
		defer os.RemoveAll(path)
	}

	get := exec.Command("go", "get", "-tags=slurp", "-v")
	get.Dir = path
	get.Stdin = os.Stdin
	get.Stdout = os.Stdout
	get.Stderr = os.Stderr

	if *build || *install {
		err := get.Run()
		if err != nil {
			return err
		}
	}

	var args []string

	if *build {
		args = []string{"build", "-tags=slurp", "-o=slurp-bin", path}

	} else if *install {
		args = []string{"install", "-tags=slurp", path}

	} else {
		params := flag.Args()

		if len(params) > 0 && params[0] == "init" {
			err := get.Run()
			if err != nil {
				return err
			}
		}

		args = []string{"run", "-tags=slurp", filepath.Join(path, "main.go"), fmt.Sprintf("-timestamp=%d", *timestamp)}
		args = append(args, params...)
	}

	cmd := exec.Command("go", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if !*build && !*install {
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)

		go func() {
			for sig := range interrupts {
				cmd.Process.Signal(sig)
			}
		}()
	}
	err = cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

func generate() (string, error) {

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// find the correct gopath
	var gopathsrc string
	var pkgpath string
	for _, gopath := range gopaths {
		gopathsrcTest := filepath.Join(gopath, "src")
		// the target package import path.
		pkgpath, err = filepath.Rel(gopathsrcTest, cwd)
		if err != nil {
			return "", err
		}
		if base := filepath.Base(pkgpath); base == "." || base == ".." {
			continue // cwd is outside this gopath
		}
		gopathsrc = gopathsrcTest
	}

	if gopathsrc == "" {
		return "", errors.New("forbidden path. Your CWD must be under $GOPATH/src.")
	}

	//build our package path.
	path := filepath.Join(gopathsrc, "slurp", pkgpath)

	//Clean it up.
	os.RemoveAll(path)

	//log.Println("Creating temporary build path...", path)
	//Create the runner package directory.
	err = os.MkdirAll(path, 0700)
	if err != nil {
		return path, err
	}

	//TODO, copy [*.go !_test.go] files into tmp first,
	// this would allow slurp to work for broken packages
	// with "-bare" as the package files will be excluded.
	fset := token.NewFileSet() // positions are relative to fset

	var pkgs map[string]*ast.Package

	//log.Printf("Parsing %s...", pkgpath)

	if *bare {
		pkgs = make(map[string]*ast.Package)
		src, err := parser.ParseFile(fset, *slurpfile, nil, parser.PackageClauseOnly)
		if err != nil {
			return path, err
		}
		pkgs[src.Name.Name] = &ast.Package{
			Name:  src.Name.Name,
			Files: map[string]*ast.File{filepath.Join(cwd, *slurpfile): src},
		}
	} else {
		pkgs, err = parser.ParseDir(fset, cwd, nil, parser.PackageClauseOnly)
		if err != nil {
			return path, err
		}
	}

	if len(pkgs) > 1 {
		return path, errors.New("Error: Multiple packages detected.")
	}

	main, ok := pkgs["main"]

	if ok {
		//Create the target package directory.
		tmp := filepath.Join(path, "tmp")
		err = os.Mkdir(tmp, 0700)
		if err != nil {
			return path, err
		}

		for file, f := range main.Files {
			name, err := filepath.Rel(cwd, file)
			if err != nil {
				//Should never get error. But just incase.
				return path, err
			}
			dstfile, err := os.Create(filepath.Join(tmp, name))
			if err != nil {
				return path, err
			}
			defer dstfile.Close()
			srcfile, err := os.Open(file)
			if err != nil {
				return path, err
			}
			defer srcfile.Close()
			_, err = io.Copy(dstfile, srcfile)
			if err != nil {
				return path, err
			}

			pos := fset.Position(f.Name.NamePos)

			_, err = dstfile.Seek(int64(pos.Offset), 0)
			if err != nil {
				return path, err
			}

			_, err = dstfile.Write([]byte(`niam`))
			if err != nil {
				return path, err
			}
		}

		pkgpath = filepath.Join("slurp", pkgpath, "tmp")
	}

	//log.Println("Generating the runner...")
	file, err := os.Create(filepath.Join(path, "main.go"))
	if err != nil {
		return path, err
	}

	//tmp = filepath.Join(tmp, "tmp")

	err = runnerSrc.Execute(file, filepath.ToSlash(pkgpath))
	if err != nil {
		return path, err
	}

	err = file.Close()
	if err != nil {
		return path, err
	}

	return path, nil

}
