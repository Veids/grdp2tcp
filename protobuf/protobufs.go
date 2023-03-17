package protobufs

import (
	"embed"
)

var (

	// FS - Embedded FS access to proto files
	//go:embed common/* client/*
	FS embed.FS
)
