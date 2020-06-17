package handler

import (
	"github.com/miky4u2/RAserver/server/config"
	"html/template"
	"path/filepath"
)

var tpl *template.Template

func parseTemplates() {
	if tpl == nil {
		tpl = template.Must(template.ParseGlob(filepath.Join(config.AppBasePath, "templates", "*.gohtml")))
	}
}
