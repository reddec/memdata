package model

import (
	"github.com/dave/jennifer/jen"
	"gtihub.com/reddec/memdata"
	"strings"
)

func GenerateProject(proj *memdata.Project) *jen.Statement {
	// prepare project
	// replace key to sequence (if it's a number) and indexed
	for _, model := range proj.Models {
		if model.Key != "" {
			model.Indexed = model.Key
			if memdata.IsNumType(model.FieldType(model.Key)) {
				model.AutoSequence = append(model.AutoSequence, model.Key)
			}
			model.Key = ""
		}
	}
	// replace fields with $ as ref
	for _, model := range proj.Models {
		cp := make(map[string]string)
		if model.Ref == nil {
			model.Ref = make(map[string]string)
		}
		for field, typeName := range model.Fields {
			if strings.HasPrefix(typeName, "$") {
				model.Ref[field] = typeName[1:]
			} else {
				cp[field] = typeName
			}
		}
		model.Fields = cp
	}
	// replace fields with ... suffix as has_many
	for _, model := range proj.Models {
		cp := make(map[string]string)
		if model.HasMany == nil {
			model.HasMany = make(map[string]string)
		}
		for field, typeName := range model.Fields {
			if strings.HasSuffix(typeName, "...") {
				model.HasMany[field] = typeName[:len(typeName)-3]
			} else {
				cp[field] = typeName
			}
		}
		model.Fields = cp
	}
	return generateProjectInterfaces(proj).Line().Add(generateProjectStruct(proj)).Line().Add(generateProjectFuncs(proj)).Line().Add(generateDefaultStorages(proj))
}

func generateProjectInterfaces(proj *memdata.Project) *jen.Statement {
	var code = jen.Line()
	for _, model := range proj.Models {
		keyName := memdata.ToLowerCamel(model.Indexed)
		code = code.Type().Id(model.Name + "Storage").InterfaceFunc(func(iface *jen.Group) {
			// PutModel (id, value)
			iface.Id("Put"+model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name))
			// GetModel (id) -> value
			iface.Id("Get" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name)
			// DeleteModel (id)
			iface.Id("Delete" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)))
			// IterateModel callback(id, value)
			iface.Id("Iterate" + model.Name).Params(jen.Id("iterator").Func().Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name)))
		}).Line().Line()
	}
	return code
}

func generateDefaultStorages(proj *memdata.Project) *jen.Statement {
	var code = jen.Line()
	for _, model := range proj.Models {
		keyName := memdata.ToLowerCamel(model.Indexed)
		// default on maps
		objName := "map" + model.Name + "Storage"
		// define struct { data map[id]*Value }
		code = code.Type().Id(objName).Struct(jen.Id("data").Map(jen.Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name)).Line()
		// define constructor
		code.Func().Id("NewMap" + model.Name + "Storage").Params().Id(model.Name + "Storage").BlockFunc(func(init *jen.Group) {
			init.Return().Op("&").Id(objName).Values(jen.Id("data").Op(":").Make(jen.Map(jen.Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name)))
		}).Line()
		// define methods

		// PutModel (id, value)
		code.Func().Params(jen.Id("storage").Op("*").Id(objName)).Id("Put"+model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name)).BlockFunc(func(putFunc *jen.Group) {
			putFunc.Id("storage").Dot("data").Index(jen.Id(keyName)).Op("=").Id("item")
		}).Line()
		// GetModel (id) -> value
		code.Func().Params(jen.Id("storage").Op("*").Id(objName)).Id("Get" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name).BlockFunc(func(getFunc *jen.Group) {
			getFunc.Return().Id("storage").Dot("data").Index(jen.Id(keyName))
		}).Line()
		// DeleteModel (id)
		code.Func().Params(jen.Id("storage").Op("*").Id(objName)).Id("Delete" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).BlockFunc(func(delFunc *jen.Group) {
			delFunc.Delete(jen.Id("storage").Dot("data"), jen.Id(keyName))
		}).Line()
		// IterateModel callback(id, value)
		code.Func().Params(jen.Id("storage").Op("*").Id(objName)).Id("Iterate" + model.Name).Params(jen.Id("iterator").Func().Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name))).BlockFunc(func(iterFunc *jen.Group) {
			iterFunc.For(jen.List(jen.Id("key"), jen.Id("item")).Op(":=").Range().Id("storage").Dot("data")).BlockFunc(func(rangeBlock *jen.Group) {
				rangeBlock.Id("iterator").Call(jen.Id("key"), jen.Id("item"))
			})
		}).Line()
	}
	return code
}

func generateProjectStruct(proj *memdata.Project) *jen.Statement {
	return jen.Type().Id(proj.Name).StructFunc(func(st *jen.Group) {
		// sequences fields
		for _, model := range proj.Models {
			for _, field := range model.AutoSequence {
				st.Id("sequence" + model.Name + field).Int64()
			}
		}
		// indexes (stores)
		for _, model := range proj.Models {
			st.Id("index" + model.Name + "By" + model.Indexed).Id(model.Name + "Storage")
		}
		// global lock if synchronized
		if proj.Synchronized {
			st.Id("_lock").Qual("sync", "RWMutex")
		}

	})
}

func generateProjectFuncs(proj *memdata.Project) jen.Code {
	// constructor
	fs := jen.Func().Id("New" + proj.Name).ParamsFunc(func(paramsBlock *jen.Group) {
		// inject storages
		for _, model := range proj.Models {
			paramsBlock.Id("storage" + model.Name + "By" + model.Indexed).Id(model.Name + "Storage")
		}
	}).Op("*").Id(proj.Name).BlockFunc(func(initFunc *jen.Group) {
		// restore sequences if needed
		for _, model := range proj.Models {
			keyName := memdata.ToLowerCamel(model.Indexed)
			for _, field := range model.AutoSequence {
				initFunc.Comment("restore auto-sequence for " + model.Name + "." + field)
				// iterate over all storage to get maximum of stored value
				varName := "max" + field + "Of" + model.Name
				initFunc.Var().Id(varName).Int64()
				initFunc.Id("storage" + model.Name + "By" + model.Indexed).Dot("Iterate" + model.Name).Call(jen.Func().Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name)).BlockFunc(func(iterFunc *jen.Group) {
					iterFunc.If(jen.Id(keyName).Op(">").Id(varName)).Block(jen.Id(varName).Op("=").Id(keyName))
				}))
			}
		}
		// setup fields
		initFunc.ReturnFunc(func(rt *jen.Group) {
			rt.Op("&").Id(proj.Name).ValuesFunc(func(fv *jen.Group) {
				for _, model := range proj.Models {
					item := "index" + model.Name + "By" + model.Indexed
					fv.Id(item).Op(":").Id("storage" + model.Name + "By" + model.Indexed)
				}
				// setup sequences from restored values
				for _, model := range proj.Models {
					for _, field := range model.AutoSequence {
						fv.Id("sequence" + model.Name + field).Op(":").Id("max" + field + "Of" + model.Name)
					}
				}
			})
		})
	}).Line()
	// default constructor (based on map)
	fs = fs.Func().Id("Default" + proj.Name).Params().Op("*").Id(proj.Name).BlockFunc(func(initFunc *jen.Group) {
		initFunc.Return().Id("New" + proj.Name).CallFunc(func(callParams *jen.Group) {
			for _, model := range proj.Models {
				callParams.Id("NewMap" + model.Name + "Storage").Call()
			}
		})
	}).Line()

	indexed := make(map[string]bool)
	// index search and access
	for _, model := range proj.Models {
		indexName := model.Name + "By" + model.Indexed
		fnName := model.Name
		keyName := memdata.ToLowerCamel(model.Indexed)
		if indexed[indexName] {
			continue
		}
		indexed[indexName] = true
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id(proj.Name)).Id(fnName).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name).BlockFunc(func(indexFunc *jen.Group) {
			if proj.Synchronized {
				indexFunc.Id("project").Dot("_lock").Dot("RLock").Call()
				indexFunc.Defer().Id("project").Dot("_lock").Dot("RUnlock").Call()
			}
			indexFunc.Return().Id("project").Dot("index" + indexName).Dot("Get" + model.Name).Call(jen.Id(keyName))
		}).Line()
	}
	// sequence access methods
	for _, model := range proj.Models {
		for _, field := range model.AutoSequence {
			fName := "Next" + model.Name + field
			fs = fs.Func().Parens(jen.Id("project").Op("*").Id(proj.Name)).Id(fName).Params().Int64().BlockFunc(func(indexFunc *jen.Group) {
				if proj.Synchronized {
					indexFunc.Return().Qual("atomic", "AddInt64").Call(jen.Id("project").Dot("sequence"+model.Name+field), jen.Lit(1))
				} else {
					indexFunc.Id("project").Dot("sequence" + model.Name + field).Op("++")
					indexFunc.Return().Id("project").Dot("sequence" + model.Name + field)
				}
			}).Line()
		}
	}
	// insert models (and assign sequences)
	for _, model := range proj.Models {
		indexName := model.Name + "By" + model.Indexed
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id(proj.Name)).Id("Insert" + model.Name).Params(jen.Id("item").Op("*").Id(model.Name)).Op("*").Id(model.Name).BlockFunc(func(indexFunc *jen.Group) {

			for _, auto := range model.AutoSequence {
				indexFunc.Id("item").Dot(auto).Op("=").Id("project").Dot("Next" + model.Name + auto).Call()
			}
			indexFunc.Id("item").Dot("_project").Op("=").Id("project")
			if proj.Synchronized {
				indexFunc.Id("project").Dot("_lock").Dot("Lock").Call()
				indexFunc.Id("project").Dot("index"+indexName).Dot("Put"+model.Name).Call(jen.Id("item").Dot(model.Indexed), jen.Id("item"))
				indexFunc.Id("project").Dot("_lock").Dot("Unlock").Call()
			} else {
				indexFunc.Id("project").Dot("index"+indexName).Dot("Put"+model.Name).Call(jen.Id("item").Dot(model.Indexed), jen.Id("item"))
			}
			indexFunc.Return().Id("item")
		}).Line()
	}
	// remove models (without following links)
	for _, model := range proj.Models {
		indexName := model.Name + "By" + model.Indexed
		keyName := memdata.ToLowerCamel(model.Indexed)
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id(proj.Name)).Id("Remove" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).BlockFunc(func(indexFunc *jen.Group) {
			if proj.Synchronized {
				indexFunc.Id("project").Dot("_lock").Dot("Lock").Call()
				indexFunc.Id("project").Dot("index" + indexName).Dot("Delete" + model.Name).Call(jen.Id(keyName))
				indexFunc.Id("project").Dot("_lock").Dot("Unlock").Call()
			} else {
				indexFunc.Id("project").Dot("index" + indexName).Dot("Delete" + model.Name).Call(jen.Id(keyName))
			}
		}).Line()
	}
	return fs
}

func Generate(proj *memdata.Project) jen.Code {
	s := GenerateProject(proj)
	for _, md := range proj.Models {
		s = s.Line().Add(GenerateModel(md))
	}
	return s
}
