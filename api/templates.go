package api

import (
	"embed"
	"html/template"
)

//go:embed templates
var tmplFS embed.FS

var tmpls *template.Template

func init() {
	tmpls = template.Must(template.ParseFS(tmplFS, "templates/*.html"))
}
