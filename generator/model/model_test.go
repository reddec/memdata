package model

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	"github.com/reddec/memdata"
	"os"
	"testing"
)

func TestGenerateModel(t *testing.T) {
	project, err := memdata.ReadFile("sample.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	bts := &bytes.Buffer{}
	err = jen.NewFile(project.Package).Add(Generate(project)).Render(bts)
	if err != nil {
		t.Error(err)
		return
	}
	os.Stderr.Write(bts.Bytes())
	os.Stderr.Sync()
}
