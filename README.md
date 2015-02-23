> Heads up! This is pre-release software.  

[![GoDoc](https://godoc.org/github.com/omeid/slurp?status.svg)](https://godoc.org/github.com/omeid/slurp)

# Slurp 
Building with Go, easier than a slurp.

Slurp is a [Gulp.js](http://gulpjs.com/) inspired build toolkit designed with idiomatic Go [Pipelines](http://blog.golang.org/pipelines) and following principles: 

- Convention over configuration
- Explicit is better than implicit.
- Do one thing. Do it well.
- ...


> ### Why?
  The tale of Gulp, Go, Go Templates, CSS, CoffeeScript, and minifiaction and building assets should go here.



Slurp is made of two integral parts:

### 1. The Toolkit 

The slurp toolkit provides a task harness that you can register tasks and dependencies, you can then run these tasks with slurp runner.

A task is any function that accepts a pointer to `slurp.C` (Slurp Context) and returns an error.  
The Context provides logging functions. _it may be extended in the future_.

```go
b.Task("example-task", []string{"list", "of", "dependency", "tasks"},

  func(c *slurp.C) error {
    c.Info("Hello from example-task!")
  },

)
```

Following the Convention Over Configuration paradigm, slurps provides you with a collection of nimble tools to instrument a pipeline.

A pipeline is created by a source _stage_ and typically piped to subsequent _transformation_ stages and a final _destination_ stage.

Currently Slurp provides two source stages `slurp/stages/fs` and `slurp/stages/web` that provide access to file-system and http source respectively.

```go
b.Task("example-task-with-pipeline", nil , func(c *slurp.C) error {
    //Read .tpl files from frontend/template.
    return fs.Src(c, "frontend/template/*.tpl").Pipe(
      //Compile them.
      template.HTML(c, TemplateData),
      //Write the result to disk.
      fs.Dest(c, "./public"),
    ).Wait() //Wait for all to finish.
})
```

or the same code shorter

```go
b.Task("example-task-with-pipeline", nil , func(c *slurp.C) error {
    //Read .tpl files from frontend/template.
    return fs.Src(c, "frontend/template/*.tpl").Then(
      //Compile them.
      template.HTML(c, TemplateData),
      //Write the result to disk.
      fs.Dest(c, "./public"),
      )
})
```

and another example,

```go
// Download deps.
b.Task("deps", nil, func(c *slurp.C) error {
    return web.Get(c,
      "https://github.com/twbs/bootstrap/archive/v3.3.2.zip",
      "https://github.com/FortAwesome/Font-Awesome/archive/v4.3.0.zip",
    ).Then(
      archive.Unzip(c),
      fs.Dest(c, "./frontend/libs/"),
    )

})
```

Currently the following _stages_ are provided with Slurp:
> No 3rd party dependency, just standard library.  

- [archive](https://godoc.org/github.com/omeid/slurp/stages/archive/)
- [fs](https://godoc.org/github.com/omeid/slurp/stages/fs/)
- [passthrough](https://godoc.org/github.com/omeid/slurp/stages/passthrough/)
- [template](https://godoc.org/github.com/omeid/slurp/stages/template/)
- [web](https://godoc.org/github.com/omeid/slurp/stages/web/)


You can find more at [slurp-contrib](https://github.com/slurp-contrib). gin, gcss, ace, watch, resources (embed), and livereload to name a few.


### 2. The Runner (cmd/slurp)

This is a cli tool that runs and help you compile your builders. It is go get’able and you can install with:

```bash
 $ go get -u -v github.com/omeid/slurp/cmd/slurp  # get it.
```

Slurp uses the Slurp build tag. That is, it passes `-tags=slurp` to go tooling when building or running your project,
this allows decoupling of build and project code. This means you can use Go tools just like you're used to, even if your
project has a slurp file.

Somewhat similar to `go test` Slurp expects a `Slurp(*slurp.Build)` function from your project, this is typically put in a file with the `// +build slurp` build tag.

```sh
github.com/omeid/slurp/examples (master) $ cat example.go
```
```go
package main //Anything, even main.

import (
  "errors"
  "time"

  "github.com/omeid/slurp"
)

func Slurp(b *slurp.Build) {
  b.Task("turtle", nil, func(c *slurp.C) error {
    c.Info("Hello!")
    c.Warn("I will take at least 3 seconds.")
    time.Sleep(3 * time.Second)
    c.Info("Well, here is a line.")
    return errors.New("I died.")
  })

  b.Task("rabbit", nil, func(c *slurp.C) error {
    c.Info("Hello, I am the the fast one.")
    for i := 0; i < 4; i++ {
      c.Infof("This is the %d line of my work.", i)
      time.Sleep(500 * time.Millisecond)
    }
    return nil
  })

  b.Task("default", []string{"turtle", "rabbit"}, func(c *slurp.C) error {
    //This task is run when slurp is called with any task parameter.
    c.Info("Default task is running.")
    return nil
  })
}
```

```sh
github.com/omeid/slurp/examples (master) $ slurp
09:22:49 [INFO] Running: default 
09:22:49 [INFO] Starting. 
09:22:49 [INFO] Waiting for turtle 
09:22:49 [INFO] turtle: Starting. 
09:22:49 [INFO] Waiting for rabbit 
09:22:49 [INFO] rabbit: Starting. 
09:22:49 [INFO] rabbit: Hello, I am the the fast one. 
09:22:49 [INFO] rabbit: This is the 0 line of my work. 
09:22:49 [INFO] turtle: Hello! 
09:22:49 [WARN] turtle: I will take at least 3 seconds. 
09:22:49 [INFO] rabbit: This is the 1 line of my work. 
09:22:50 [INFO] rabbit: This is the 2 line of my work. 
09:22:50 [INFO] rabbit: This is the 3 line of my work. 
09:22:51 [INFO] rabbit: Done. 
09:22:52 [INFO] turtle: Well, here is a line. 
09:22:52 [ERR!] turtle: I died. 
09:22:52 [ERR!] Task Canacled. Reason: Failed Dependency (turtle).

github.com/omeid/slurp/examples (master) $ slurp turtle
09:27:19 [INFO] Running: turtle 
09:27:19 [INFO] Starting. 
09:27:19 [INFO] Hello! 
09:27:19 [WARN] I will take at least 3 seconds. 
09:27:22 [INFO] Well, here is a line. 
09:27:22 [ERR!] I died. 

github.com/omeid/slurp/examples (master) $ slurp rabbit
09:27:25 [INFO] Running: rabbit 
09:27:25 [INFO] Starting. 
09:27:25 [INFO] Hello, I am the the fast one. 
09:27:25 [INFO] This is the 0 line of my work. 
09:27:26 [INFO] This is the 1 line of my work. 
09:27:26 [INFO] This is the 2 line of my work. 
09:27:27 [INFO] This is the 3 line of my work. 
09:27:27 [INFO] Done. 
```

### Contributing

Please see [Contributing](CONTRIBUTING.md)


### Examples
 - The obligatory [Todo App (Slurp)](https://github.com/omeid/slurp-todo)


### LICENSE
  [MIT](LICENSE).
