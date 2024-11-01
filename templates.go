package main

import "embed"

//go:embed "templates/*"
var TemplatesFileSystem embed.FS
