package main

import (
	"github.com/dave/jennifer/jen"
	"github.com/jessevdk/go-flags"
	"gtihub.com/reddec/memdata"
	"memdata/generator/model"
	"os"
)

var config struct {
	Args struct {
		File string `positional-arg-name:"file" description:"path to project YAML file"`
	} `positional-args:"yes" required:"yes"`
}

func main() {
	_, err := flags.Parse(&config)
	if err != nil {
		os.Exit(1)
	}
	project, err := memdata.ReadFile(config.Args.File)
	if err != nil {
		panic(err)
		return
	}
	modelGen := model.Generate(project)
	fs := jen.NewFile(project.Package)
	fs.Add(modelGen)
	err = fs.Render(os.Stdout)
	if err != nil {
		panic(err)
		return
	}
}
