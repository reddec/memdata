package memdata

type Model struct {
	Name         string
	Fields       map[string]string
	Unique       []string
	Indexed      string
	Ref          map[string]string
	HasMany      map[string]string `yaml:"has_many"`
	AutoSequence []string          `yaml:"auto_sequence"`
	Project      *Project          `yaml:"-"`
}

type Project struct {
	Package       string `yaml:"package"`
	Name          string
	Synchronized  bool
	Imports       map[string]string
	Models        []*Model
	IncludeModels []string `yaml:"include_models"`
}
