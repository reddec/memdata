package memdata

import (
	"github.com/dave/jennifer/jen"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

func ReadString(yamlStr string) (*Project, error) {
	var model Project
	err := yaml.Unmarshal([]byte(yamlStr), &model)
	if err != nil {
		return nil, err
	}
	for _, m := range model.Models {
		m.Project = &model
	}
	return &model, nil
}

func ReadModelFile(file string, project *Project) (*Model, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var model Model
	err = yaml.Unmarshal(data, &model)
	if err != nil {
		return nil, err
	}
	model.Project = project
	if model.Name == "" {
		name := filepath.Base(file)
		ext := filepath.Ext(name)
		name = name[:len(name)-len(ext)]
		model.Name = strings.Title(name)
	}
	return &model, nil
}

func ReadFile(file string) (*Project, error) {
	rootDir := filepath.Dir(file)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	project, err := ReadString(string(data))
	if err != nil {
		return nil, err
	}
	for _, include := range project.IncludeModels {
		if !filepath.IsAbs(include) {
			include = filepath.Join(rootDir, include)
		}
		model, err := ReadModelFile(include, project)
		if err != nil {
			return nil, err
		}
		project.Models = append(project.Models, model)
	}
	return project, nil
}

func (prj *Project) Model(name string) *Model {
	for _, m := range prj.Models {
		if m.Name == name {
			return m
		}
	}
	panic("model " + name + " not found in project")
}

func (md *Model) FieldType(name string) string {
	tp, ok := md.Fields[name]
	if !ok {
		panic("no field " + name + " in model " + md.Name)
	}
	return tp
}

var opsPat = regexp.MustCompile(`^[^\w]*`)

func (proj *Project) Qual(fieldType string) jen.Code {
	if strings.Contains(fieldType, ".") {
		ops := opsPat.FindString(fieldType)
		fieldType = fieldType[len(ops):]
		pts := strings.Split(fieldType, ".")
		pckg := proj.Imports[pts[0]]
		return jen.Op(ops).Qual(pckg, pts[1])
	}
	return jen.Id(fieldType)
}

func IsNumType(t string) bool {
	switch t {
	case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "float32", "float64", "byte", "rune":
		return true
	}
	return false
}

func ToLowerCamel(t string) string {
	return strings.ToLower(t[:1]) + t[1:]
}
