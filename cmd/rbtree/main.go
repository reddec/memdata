package main

import (
	"github.com/jessevdk/go-flags"
	"memdata/generator/tree"
	"os"
)

func main() {
	config := tree.Tree{}
	_, err := flags.Parse(&config)
	if err != nil {
		os.Exit(1)
	}
	os.Stdout.WriteString(config.Generate())
}
