package embedded

import (
	"embed"
)

// global variable that holds the current embedded files
//
//go:embed files/*
var currentEmbedded embed.FS

func ReturnEmbedded() embed.FS {
	return currentEmbedded
}
