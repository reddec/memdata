package memdata

type Model struct {
	Name         string
	Fields       map[string]string
	Indexed      string
	Ref          map[string]string
	HasMany      map[string]string `yaml:"many"`
	AutoSequence []string          `yaml:"sequence"`
	Key          string            `yaml:"key"` // helper: adds to auto seq, unique and indexed
	Project      *Project          `yaml:"-"`
}

type Project struct {
	Package       string `yaml:"package"`
	Name          string
	Synchronized  bool
	Imports       map[string]string
	Models        []*Model
	StorageRef    bool `yaml:"storage_ref"`
	Transactional bool
	IncludeModels []string `yaml:"include_models"`
}
