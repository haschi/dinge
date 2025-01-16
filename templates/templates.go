package templates

import "embed"

//go:embed common/* layout/* pages/*
var TemplatesFileSystem embed.FS
