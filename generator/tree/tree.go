package tree

import (
	"bytes"
	"text/template"
)

type Tree struct {
	Package    string   `yaml:"package" long:"package" description:"package name in generated file" default:"tree" env:"PACKAGE"`
	Imports    []string `yaml:"imports" long:"import"  description:"imports included to the generated file" env:"IMPORTS" env-delim:","`
	TypeName   string   `yaml:"type"    long:"type"    description:"tree type name" env:"TYPENAME" default:"tree"`
	KeyType    string   `yaml:"key"     long:"key"     description:"key typename" env:"KEY" default:"int64"`
	ValueType  string   `yaml:"value"   long:"value"   description:"value typename" env:"VALUE" required:"yes"`
	Comparator bool     `yaml:"cmp"     long:"cmp"     description:"user Cmp method to compare keys"`
}

func (t *Tree) IsKeyNum() bool {
	switch t.KeyType {
	case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "float32", "float64", "byte", "rune":
		return true
	}
	return false
}

//go:generate go-bindata -pkg tree template.gotemplate

func (tree *Tree) Generate() string {
	t := template.Must(template.New("").Parse(string(MustAsset("template.gotemplate"))))
	buf := &bytes.Buffer{}
	err := t.Execute(buf, tree)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
