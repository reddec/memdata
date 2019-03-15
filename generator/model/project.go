package model

import (
	"github.com/dave/jennifer/jen"
	"memdata"
)

func GenerateProject(proj *memdata.Project) jen.Code {
	return jen.Empty().Add(generateProjectStruct(proj)).Line().Add(generateProjectFuncs(proj))
}

func generateProjectStruct(proj *memdata.Project) jen.Code {
	return jen.Type().Id(proj.Name).StructFunc(func(st *jen.Group) {
		// sequences fields
		for _, model := range proj.Models {
			for _, field := range model.AutoSequence {
				st.Id("sequence" + model.Name + field).Int64()
			}
		}
		// indexes (also stores)
		for _, model := range proj.Models {
			st.Id("index" + model.Name + "By" + model.Indexed).Map(jen.Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name)
		}
		// global lock if synchronized
		if proj.Synchronized {
			st.Id("_lock").Qual("sync", "RWMutex")
		}

	}).Line()
}

func generateProjectFuncs(proj *memdata.Project) jen.Code {
	fs := jen.Empty()
	// constructor
	fs = fs.Func().Id("New" + proj.Name).Params().Op("*").Id(proj.Name).BlockFunc(func(initFunc *jen.Group) {
		initFunc.ReturnFunc(func(rt *jen.Group) {
			rt.Op("&").Id(proj.Name).ValuesFunc(func(fv *jen.Group) {
				for _, model := range proj.Models {
					item := "index" + model.Name + "By" + model.Indexed
					fv.Id(item).Op(":").Make(jen.Map(jen.Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name))
				}
			})
		})
	}).Line()
	indexed := make(map[string]bool)
	// index search and access
	for _, model := range proj.Models {
		indexName := model.Name + "By" + model.Indexed
		fnName := model.Name
		if indexed[indexName] {
			continue
		}
		indexed[indexName] = true
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id(proj.Name)).Id(fnName).Params(jen.Id("key").Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name).BlockFunc(func(indexFunc *jen.Group) {
			if proj.Synchronized {
				indexFunc.Id("project").Dot("_lock").Dot("RLock").Call()
				indexFunc.Defer().Id("project").Dot("_lock").Dot("RUnlock").Call()
			}
			indexFunc.Return().Id("project").Dot("index" + indexName).Index(jen.Id("key"))
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
				indexFunc.Id("project").Dot("index" + indexName).Index(jen.Id("item").Dot(model.Indexed)).Op("=").Id("item")
				indexFunc.Id("project").Dot("_lock").Dot("Unlock").Call()
			} else {
				indexFunc.Id("project").Dot("index" + indexName).Index(jen.Id("item").Dot(model.Indexed)).Op("=").Id("item")
			}
			indexFunc.Return().Id("item")
		}).Line()
	}
	// remove models (without following links)
	for _, model := range proj.Models {
		indexName := model.Name + "By" + model.Indexed
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id(proj.Name)).Id("Remove" + model.Name).Params(jen.Id("key").Id(model.FieldType(model.Indexed))).BlockFunc(func(indexFunc *jen.Group) {
			if proj.Synchronized {
				indexFunc.Id("project").Dot("_lock").Dot("Lock").Call()
				indexFunc.Delete(jen.Id("project").Dot("index"+indexName), jen.Id("key"))
				indexFunc.Id("project").Dot("_lock").Dot("Unlock").Call()
			} else {
				indexFunc.Delete(jen.Id("project").Dot("index"+indexName), jen.Id("key"))
			}
		}).Line()
	}
	return fs
}

func Generate(proj *memdata.Project) jen.Code {
	s := jen.Empty().Add(GenerateProject(proj))
	for _, md := range proj.Models {
		s = s.Line().Add(GenerateModel(md))
	}
	return s
}
