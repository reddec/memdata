package btree

import (
	"bytes"
	"github.com/reddec/memdata"
	"text/template"
)

type BTree struct {
	Package    string   `yaml:"package" long:"package" description:"package name in generated file" default:"tree" env:"PACKAGE"`
	Imports    []string `yaml:"imports" long:"import"  description:"imports included to the generated file" env:"IMPORTS" env-delim:","`
	TypeName   string   `yaml:"type"    long:"type"    description:"tree type name" env:"TYPENAME" default:"BTree"`
	KeyType    string   `yaml:"key"     long:"key"     description:"key typename" env:"KEY" default:"int64"`
	ValueType  string   `yaml:"value"   long:"value"   description:"value typename" env:"VALUE" required:"yes"`
	Comparator bool     `yaml:"cmp"     long:"cmp"     description:"user Cmp method to compare keys"`
}

func (t *BTree) IsKeyNum() bool {
	return memdata.IsNumType(t.KeyType)
}

//go:generate go-bindata -pkg btree template.gotemplate

func (tree *BTree) Generate() string {
	t := template.Must(template.New("").Parse(string(MustAsset("template.gotemplate"))))
	buf := &bytes.Buffer{}
	err := t.Execute(buf, tree)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
