package tree

import "testing"

func TestGenerate(t *testing.T) {
	tree := Tree{
		KeyType:   "int64",
		ValueType: "apd.Decimal",
		TypeName:  "MyTree",
		Package:   "sample",
		Imports:   []string{"github.com/reddec/apd"},
	}
	t.Log(tree.Generate())
}
