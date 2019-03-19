package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/reddec/memdata/generator/btree"
	"os"
)

func main() {
	config := btree.BTree{}
	_, err := flags.Parse(&config)
	if err != nil {
		os.Exit(1)
	}
	os.Stdout.WriteString(config.Generate())
}
