package slurp

import (
  "text/template"
)

// The text template for the build help topic.
// slrup uses text/template to render templates. You can
// render custom help text by setting this variable.
var BuildHelpTemplate = `SLURP:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} [global options]  {task [flags]...}

VERSION:
   {{.Version}}{{if or .Author .Email}}

AUTHOR:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}

TASKS:
   {{range .Tasks }}{{printf "%-15s %s" .Name .Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

// The text template for the task help topic.
// slurp uses text/template to render templates. You can
// render custom help text by setting this variable.
var TaskHelpTemplate = `TASK:
   {{.Name}} - {{.Usage}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .Deps}}

DEPENDENCIES:
   {{ range .Deps }}{{ . }}
   {{ end }}{{ end }}{{ if .Flags }}

OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{ end }}
`

var HelpTemplate *template.Template

func init() {
  HelpTemplate = template.Must(template.New("build").Parse(BuildHelpTemplate))
  HelpTemplate = template.Must(HelpTemplate.New("task").Parse(TaskHelpTemplate))
}
