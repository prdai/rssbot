package utils

import (
	"log/slog"

	"github.com/noirbizarre/gonja"
)

func LoadTemplate(path string) string {
	tpl := gonja.Must(gonja.FromFile(path))
	out, err := tpl.Execute(gonja.Context{})
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	return out
}
