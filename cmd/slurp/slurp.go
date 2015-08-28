package main

import (
	"errors"
	"flag"
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
	gopaths = strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator))
	cwd     string

	flags     = flag.NewFlagSet("slurp", flag.ContinueOnError)
	build     = flags.Bool("build", false, "build the current build as slurp-bin")
	install   = flags.Bool("install", false, "install current slurp.Go as slurp.PKG.")
	bare      = flags.Bool("bare", false, "Run/Install the slurp.go file without any other files.")
	slurpfile = flags.String("slurpfile", "slurp.go", "The file that includes the Slurp(*s.Build) function, use by -bare")
	keep      = flags.Bool("keep", false, "keep the generated source under $GOPATH/src/slurp/IMPORT/PATH")
)

func init() {
	maxprocs := runtime.NumCPU()
	if maxprocs > 2 {
		runtime.GOMAXPROCS(maxprocs / 2)
	}
	flags.Usage = func() {}
}

func main() {

	flags.Parse(os.Args[1:])

	if len(gopaths) == 0 || gopaths[0] == "" {
		log.Fatal("$GOPATH must be set.")
	}

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func runnerpath(path string) string {
	return filepath.Join(path, "slurp-"+filepath.Base(path))
}

func run() error {
	path, pkgpath, err := generate()
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

	path = runnerpath(path)

	if *build {
		args = []string{"build", "-tags=slurp", "-o=slurp-bin", runnerpath(pkgpath)}
	} else if *install {
		args = []string{"install", "-tags=slurp", runnerpath(pkgpath)}
	} else {
		params := os.Args[1:]

		if len(params) > 0 && params[0] == "init" {
			err := get.Run()
			if err != nil {
				return err
			}
		}

		args = []string{"run", "-tags=slurp", filepath.Join(path, "main.go")}
		args = append(args, params...)
	}

	log.Println(args)

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "GO15VENDOREXPERIMENT=1")
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

func generate() (string, string, error) {

	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	// find the correct gopath
	var gopathsrc string
	var pkgpath string
	for _, gopath := range gopaths {
		gopathsrcTest := filepath.Join(gopath, "src")
		// the target package import path.
		pkgpath, err = filepath.Rel(gopathsrcTest, cwd)
		if err != nil {
			return "", "", err
		}
		if base := filepath.Base(pkgpath); base == "." || base == ".." {
			continue // cwd is outside this gopath
		}
		gopathsrc = gopathsrcTest
		break
	}

	if gopathsrc == "" {
		return "", pkgpath, errors.New("forbidden path. Your CWD must be under $GOPATH/src.")
	}

	//build our package path.
	path := filepath.Join(gopathsrc, "slurp", pkgpath)

	//Clean it up.
	os.RemoveAll(path)

	//log.Println("Creating temporary build path...", path)
	//Create the runner package directory.
	err = os.MkdirAll(path, 0700)
	if err != nil {
		return path, pkgpath, err
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
			return path, pkgpath, err
		}
		pkgs[src.Name.Name] = &ast.Package{
			Name:  src.Name.Name,
			Files: map[string]*ast.File{filepath.Join(cwd, *slurpfile): src},
		}
	} else {
		pkgs, err = parser.ParseDir(fset, cwd, nil, parser.PackageClauseOnly)
		if err != nil {
			return path, pkgpath, err
		}
	}

	if len(pkgs) > 1 {
		return path, pkgpath, errors.New("Error: Multiple packages detected.")
	}

	main, ok := pkgs["main"]

	if ok {

		for file, f := range main.Files {
			name, err := filepath.Rel(cwd, file)
			if err != nil {
				//Should never get error. But just incase.
				return path, pkgpath, err
			}
			srcfile, dstfile, err := copyFile(file, filepath.Join(path, name))
			if err != nil {
				return path, pkgpath, err
			}
			defer dstfile.Close()
			defer srcfile.Close()

			pos := fset.Position(f.Name.NamePos)

			_, err = dstfile.Seek(int64(pos.Offset), 0)
			if err != nil {
				return path, pkgpath, err
			}

			_, err = dstfile.Write([]byte(`niam`))
			if err != nil {
				return path, pkgpath, err
			}
		}

		pkgpath = filepath.Join("slurp", pkgpath)
	}

	if info, err := os.Stat("vendor"); err == nil && info.IsDir() {
		if err := filepath.Walk("vendor", func(p string, info os.FileInfo, _ error) error {
			if info.IsDir() {
				return os.Mkdir(filepath.Join(path, p), 0700)
			}
			if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			srcfile, dstfile, err := copyFile(p, filepath.Join(path, p))
			if err != nil {
				return err
			}
			srcfile.Close()
			dstfile.Close()
			return nil
		}); err != nil {
			log.Fatal(err)
		}
	}

	//log.Println("Generating the runner...")
	runner := runnerpath(path)
	err = os.Mkdir(runner, 0700)
	if err != nil {
		return path, pkgpath, err
	}

	file, err := os.Create(filepath.Join(runner, "main.go"))
	if err != nil {
		return path, pkgpath, err
	}

	//tmp = filepath.Join(tmp, "tmp")

	err = runnerSrc.Execute(file, filepath.ToSlash(pkgpath))
	if err != nil {
		return path, pkgpath, err
	}

	err = file.Close()
	return path, pkgpath, err
}

// copyFile is a helper method for copying a file to another location.
// If err is non-nil, then the returned files must be closed by the caller.
func copyFile(src, dest string) (srcfile *os.File, dstfile *os.File, err error) {
	df, err := os.Create(dest)
	if err != nil {
		return nil, nil, err
	}

	sf, err := os.Open(src)
	if err != nil {
		df.Close()
		return nil, nil, err
	}

	_, err = io.Copy(df, sf)
	return sf, df, err
}
