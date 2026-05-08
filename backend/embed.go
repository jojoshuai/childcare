//go:build embed

package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/dist/*
var frontendEmbed embed.FS

// FrontendFS serves the embedded frontend dist files.
var FrontendFS http.FileSystem = mustFrontendFS()

func mustFrontendFS() http.FileSystem {
	sub, err := fs.Sub(frontendEmbed, "web/dist")
	if err != nil {
		panic("frontend embed: " + err.Error())
	}
	return http.FS(sub)
}
