package admin

import (
	"embed"
	"io/fs"
)

//go:embed templates/*
var templatesFS embed.FS

// GetTemplateFS returns the filesystem for admin templates
func GetTemplateFS() (fs.FS, error) {
	return fs.Sub(templatesFS, "templates")
}
