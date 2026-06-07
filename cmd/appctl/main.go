package main

import "github.com/bulaya-ute/appctl/internal/cli"

// version is set at build time via -ldflags="-X main.version=vX.Y.Z"
var version = "dev"

func main() {
	cli.Execute(version)
}
