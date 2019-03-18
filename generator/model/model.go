package model

import (
	"github.com/dave/jennifer/jen"
	"gtihub.com/reddec/memdata"
)

func GenerateModel(model *memdata.Model) *jen.Statement {
	return generateModelStruct(model).Add(generateModelFuncs(model))
}

func generateModelStruct(model *memdata.Model) *jen.Statement {
	return jen.Type().Id(model.Name).StructFunc(func(st *jen.Group) {
		for name, fieldType := range model.Fields {
			st.Id(name).Add(model.Project.Qual(fieldType))
		}
		// to-many array links
		for fieldName, modelRef := range model.HasMany {
			targetModel := model.Project.Model(modelRef)
			refType := targetModel.FieldType(targetModel.Indexed)
			keysSlice := fieldName + targetModel.Indexed
			st.Id(keysSlice).Index().Id(refType)
		}
		// to-one links
		for fieldName, modelRef := range model.Ref {
			targetModel := model.Project.Model(modelRef)
			refType := targetModel.FieldType(targetModel.Indexed)
			fieldName := fieldName + targetModel.Indexed
			st.Id(fieldName).Id(refType)
		}

		st.Id("_project").Op("*").Id(model.Project.Name).Tag(map[string]string{"msgp": "-"})
	}).Line()
}

func generateModelFuncs(model *memdata.Model) *jen.Statement {
	fns := jen.Line()
	// add references access
	for refName, ref := range model.Ref {
		targetModel := model.Project.Model(ref)
		fnName := ref
		fns = fns.Func().Parens(jen.Id("model").Op("*").Id(model.Name)).Id(refName).Call().Op("*").Id(ref).BlockFunc(func(refFun *jen.Group) {
			refFun.Return(jen.Id("model").Dot("_project").Dot(fnName)).Call(jen.Id("model").Dot(refName + targetModel.Indexed))
		}).Line()
	}
	// add access to the project
	fns = fns.Func().Parens(jen.Id("model").Op("*").Id(model.Name)).Id(model.Project.Name).Params().Op("*").Id(model.Project.Name).BlockFunc(func(projRef *jen.Group) {
		projRef.Return().Id("model").Dot("_project")
	}).Line()
	// add access by one-to-many
	for fieldName, targetModel := range model.HasMany {
		targetModel := model.Project.Model(targetModel)
		keysSlice := fieldName + targetModel.Indexed
		indexName := targetModel.Name
		fns = fns.Func().Parens(jen.Id("model").Op("*").Id(model.Name)).Id(fieldName).Params().Index().Op("*").Id(targetModel.Name).BlockFunc(func(manyRef *jen.Group) {
			manyRef.Var().Id("items").Op("=").Make(jen.Index().Op("*").Id(targetModel.Name), jen.Len(jen.Id("model").Dot(keysSlice)))
			manyRef.For(jen.List(jen.Id("i"), jen.Id("key")).Op(":=").Range().Id("model").Dot(keysSlice)).BlockFunc(func(iter *jen.Group) {
				iter.Id("items").Index(jen.Id("i")).Op("=").Id("model").Dot("_project").Dot(indexName).Call(jen.Id("key"))
			})
			manyRef.Return().Id("items")
		}).Line()
	}
	return fns
}
