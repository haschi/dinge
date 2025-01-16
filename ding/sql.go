package ding

import (
	_ "embed"
)

//go:embed create.sql
var CreateScript string

//go:embed fixture.sql
var FixtureScript string
