package model

import (
	"github.com/dave/jennifer/jen"
	"github.com/reddec/memdata"
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
	// prepare for transactional
	if proj.Transactional {
		proj.Synchronized = false
		proj.StorageRef = false
	}
	code := generateProjectInterfaces(proj).Line().Add(generateProjectStruct(proj)).Line().Add(generateProjectFuncs(proj))
	if proj.Transactional {
		code.Line().Add(generateDefaultTransactionalStorage(proj))
	} else {
		code.Line().Add(generateDefaultStorages(proj))
	}
	return code
}

func generateProjectInterfaces(proj *memdata.Project) *jen.Statement {
	indexed := make(map[string]bool)
	// project main interface - reader
	var code = jen.Type().Id(proj.Name + "Reader").InterfaceFunc(func(iface *jen.Group) {
		// index search and access
		for _, model := range proj.Models {
			indexName := model.Name + "By" + model.Indexed
			fnName := model.Name
			keyName := memdata.ToLowerCamel(model.Indexed)
			if indexed[indexName] {
				continue
			}
			indexed[indexName] = true
			iface.Id(fnName).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name)
		}
	}).Line().Line()
	// project main interface - writer
	code.Type().Id(proj.Name + "Writer").InterfaceFunc(func(iface *jen.Group) {
		for _, model := range proj.Models {
			// insert models (and assign sequences)
			iface.Id("Insert" + model.Name).Params(jen.Id("item").Op("*").Id(model.Name)).Op("*").Id(model.Name)
			// remove models (without following links)
			keyName := memdata.ToLowerCamel(model.Indexed)
			iface.Id("Remove" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)))
			// update model
			iface.Id("Update" + model.Name).Params(jen.Id("item").Op("*").Id(model.Name)).Op("*").Id(model.Name)
		}
	}).Line().Line()
	// read-writer
	code.Type().Id(proj.Name + "ReadWriter").InterfaceFunc(func(iface *jen.Group) {
		iface.Id(proj.Name + "Reader")
		iface.Id(proj.Name + "Writer")
	}).Line().Line()

	if proj.Transactional {
		// Generate transactional interfaces based on reader and writer
		code.Type().Id(proj.Name + "ReadWriterTx").InterfaceFunc(func(iface *jen.Group) {
			iface.Id(proj.Name + "Reader")
			iface.Id(proj.Name + "Writer")
			iface.Id("Commit").Params()
			iface.Id("Discard").Params()
		}).Line().Line()
		code.Type().Id(proj.Name + "ReaderTx").InterfaceFunc(func(iface *jen.Group) {
			iface.Id(proj.Name + "Reader")
			iface.Id("ReadUnlock").Params()
		}).Line().Line()
		// add action enum definitions
		code.Add(generateTransactionalDefines(proj))
	}

	// project main interface
	code.Type().Id(proj.Name).InterfaceFunc(func(iface *jen.Group) {
		if proj.Transactional {
			// generate access to read-view and write-lock transactions
			iface.Id("ReadLock").Params().Id(proj.Name + "ReaderTx")
			iface.Id("ReadWriteLock").Params().Id(proj.Name + "ReadWriterTx")
		} else {
			// plain read-write - alias to ReadWriter
			iface.Id(proj.Name + "ReadWriter")
		}
	}).Line().Line()
	if proj.Transactional {
		// transactional models storage should be only one
		code.Type().Id(proj.Name + "TxStorage").InterfaceFunc(func(iface *jen.Group) {
			for _, model := range proj.Models {
				keyName := memdata.ToLowerCamel(model.Indexed)
				// GetModel (id) -> value
				iface.Id("Get" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name)
				// IterateModel callback(id, value)
				iface.Id("Iterate" + model.Name).Params(jen.Id("iterator").Func().Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name)))
			}
			// transactional changes
			iface.Id("Apply").Params(jen.Id("batch").Index().Id(proj.Name + "LogEntity"))
		}).Line().Line()

	} else {
		// non-transactional models interfaces
		for _, model := range proj.Models {
			keyName := memdata.ToLowerCamel(model.Indexed)
			code = code.Type().Id(model.Name + "Storage").InterfaceFunc(func(iface *jen.Group) {
				// PutModel (id, value)
				iface.Id("Put"+model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name))
				// UpdateModel (id, newValue)
				iface.Id("Update"+model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name))
				// GetModel (id) -> value
				iface.Id("Get" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name)
				// DeleteModel (id)
				iface.Id("Delete" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)))
				// IterateModel callback(id, value)
				iface.Id("Iterate" + model.Name).Params(jen.Id("iterator").Func().Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name)))
			}).Line().Line()
		}
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
		// UpdateModel (id, value)
		code.Func().Params(jen.Id("storage").Op("*").Id(objName)).Id("Update"+model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name)).BlockFunc(func(putFunc *jen.Group) {
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

func generateDefaultTransactionalStorage(proj *memdata.Project) jen.Code {
	var code = jen.Type().Id("mem" + proj.Name + "MapStorage").StructFunc(func(store *jen.Group) {
		for _, model := range proj.Models {
			store.Id(model.Name).Map(jen.Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name)
		}
	})
	// define constructor
	code.Line().Func().Id("NewMap" + proj.Name + "Storage").Params().Id(proj.Name + "TxStorage").BlockFunc(func(init *jen.Group) {
		init.Return().Op("&").Id("mem" + proj.Name + "MapStorage").ValuesFunc(func(vals *jen.Group) {
			for _, model := range proj.Models {
				vals.Id(model.Name).Op(":").Make(jen.Map(jen.Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name))
			}
		})
	}).Line()

	objName := "mem" + proj.Name + "MapStorage"
	for _, model := range proj.Models {
		keyName := memdata.ToLowerCamel(model.Indexed)
		// GetModel (id) -> value
		code.Func().Params(jen.Id("storage").Op("*").Id(objName)).Id("Get" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name).BlockFunc(func(getFunc *jen.Group) {
			getFunc.Return().Id("storage").Dot(model.Name).Index(jen.Id(keyName))
		}).Line()
		// IterateModel callback(id, value)
		code.Func().Params(jen.Id("storage").Op("*").Id(objName)).Id("Iterate" + model.Name).Params(jen.Id("iterator").Func().Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name))).BlockFunc(func(iterFunc *jen.Group) {
			iterFunc.For(jen.List(jen.Id("key"), jen.Id("item")).Op(":=").Range().Id("storage").Dot(model.Name)).BlockFunc(func(rangeBlock *jen.Group) {
				rangeBlock.Id("iterator").Call(jen.Id("key"), jen.Id("item"))
			})
		}).Line()
	}
	code.Func().Params(jen.Id("storage").Op("*").Id(objName)).Id("Apply").Params(jen.Id("batch").Index().Id(proj.Name + "LogEntity")).BlockFunc(func(batchFunc *jen.Group) {
		batchFunc.For().List(jen.Id("_"), jen.Id("tx")).Op(":=").Range().Id("batch").BlockFunc(func(batchItem *jen.Group) {
			for _, model := range proj.Models {
				batchItem.If(jen.Id("tx").Dot(model.Name).Op("!=").Nil()).BlockFunc(func(modelChange *jen.Group) {
					modelChange.Switch(jen.Id("tx").Dot(model.Name).Dot("Action")).BlockFunc(func(action *jen.Group) {
						// put or update
						action.Case(jen.Id(proj.Name+"ActionInsert"), jen.Id(proj.Name+"ActionUpdate")).BlockFunc(func(operation *jen.Group) {
							operation.Id("storage").Dot(model.Name).Index(jen.Id("tx").Dot(model.Name).Dot(model.Indexed)).Op("=").Op("&").Id("tx").Dot(model.Name).Dot("Item")
						})
						// delete
						action.Case(jen.Id(proj.Name + "ActionDelete")).BlockFunc(func(operation *jen.Group) {
							operation.Delete(jen.Id("storage").Dot(model.Name), jen.Id("tx").Dot(model.Name).Dot(model.Indexed))
						})
					})

				})
			}
		})
	}).Line()
	return code
}

func generateProjectStruct(proj *memdata.Project) *jen.Statement {
	return jen.Type().Id("impl" + proj.Name).StructFunc(func(st *jen.Group) {
		// sequences fields
		for _, model := range proj.Models {
			for _, field := range model.AutoSequence {
				st.Id("sequence" + model.Name + field).Int64()
			}
		}
		if proj.Transactional {
			// single storage for everything
			st.Id("storage").Id(proj.Name + "TxStorage")
		} else {
			// indexes (stores)
			for _, model := range proj.Models {
				st.Id("index" + model.Name + "By" + model.Indexed).Id(model.Name + "Storage")
			}
		}
		// global lock if synchronized
		if proj.Synchronized {
			st.Id("_lock").Qual("sync", "RWMutex")
		}

		if proj.Transactional {
			// rw-lock
			st.Id("_tx").Qual("sync", "RWMutex")
			// changes
			st.Id("_log").Index().Id(proj.Name + "LogEntity")
		}
	})
}

func generateProjectFuncs(proj *memdata.Project) jen.Code {
	// constructor
	fs := jen.Func().Id("New" + proj.Name).ParamsFunc(func(paramsBlock *jen.Group) {
		// inject storages
		if proj.Transactional {
			// single transactional storage
			paramsBlock.Id("storage").Id(proj.Name + "TxStorage")
		} else {
			// regular single storage for each model
			for _, model := range proj.Models {
				paramsBlock.Id("storage" + model.Name + "By" + model.Indexed).Id(model.Name + "Storage")
			}
		}
	}).Id(proj.Name).BlockFunc(func(initFunc *jen.Group) {
		// restore sequences if needed

		for _, model := range proj.Models {
			keyName := memdata.ToLowerCamel(model.Indexed)
			storName := "storage" + model.Name + "By" + model.Indexed
			if proj.Transactional {
				storName = "storage"
			}
			for _, field := range model.AutoSequence {
				initFunc.Comment("restore auto-sequence for " + model.Name + "." + field)
				// iterate over all storage to get maximum of stored value
				varName := "max" + field + "Of" + model.Name
				initFunc.Var().Id(varName).Int64()
				initFunc.Id(storName).Dot("Iterate" + model.Name).Call(jen.Func().Params(jen.Id(keyName).Id(model.FieldType(model.Indexed)), jen.Id("item").Op("*").Id(model.Name)).BlockFunc(func(iterFunc *jen.Group) {
					iterFunc.If(jen.Id(keyName).Op(">").Id(varName)).Block(jen.Id(varName).Op("=").Id(keyName))
				}))
			}
		}
		// setup fields
		initFunc.ReturnFunc(func(rt *jen.Group) {
			rt.Op("&").Id("impl" + proj.Name).ValuesFunc(func(fv *jen.Group) {
				if proj.Transactional {
					fv.Id("storage").Op(":").Id("storage")
				} else {
					for _, model := range proj.Models {
						item := "index" + model.Name + "By" + model.Indexed
						fv.Id(item).Op(":").Id("storage" + model.Name + "By" + model.Indexed)
					}
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
	fs = fs.Func().Id("Default" + proj.Name).Params().Id(proj.Name).BlockFunc(func(initFunc *jen.Group) {
		initFunc.Return().Id("New" + proj.Name).CallFunc(func(callParams *jen.Group) {
			if proj.Transactional {
				callParams.Id("NewMap" + proj.Name + "Storage").Call()
			} else {
				for _, model := range proj.Models {
					callParams.Id("NewMap" + model.Name + "Storage").Call()
				}
			}
		})
	}).Line()
	if proj.Transactional {
		// generate access to read-view and write-lock transactions
		fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id("ReadLock").Params().Id(proj.Name + "ReaderTx").BlockFunc(func(txFunc *jen.Group) {
			txFunc.Id("project").Dot("_tx").Dot("RLock").Call()
			txFunc.Return(jen.Id("project"))
		}).Line()
		fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id("ReadWriteLock").Params().Id(proj.Name + "ReadWriterTx").BlockFunc(func(txFunc *jen.Group) {
			txFunc.Id("project").Dot("_tx").Dot("Lock").Call()
			txFunc.Return(jen.Id("project"))
		}).Line()
		// unlocks
		fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id("ReadUnlock").Params().BlockFunc(func(txFunc *jen.Group) {
			txFunc.Id("project").Dot("_tx").Dot("RUnlock").Call()
		}).Line()
		fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id("Commit").Params().BlockFunc(func(txFunc *jen.Group) {
			txFunc.Id("project").Dot("storage").Dot("Apply").Call(jen.Id("project").Dot("_log"))
			txFunc.Id("project").Dot("Discard").Call()
		}).Line()
		fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id("Discard").Params().BlockFunc(func(txFunc *jen.Group) {
			txFunc.If(jen.Id("project").Dot("_log").Op("!=").Nil()).BlockFunc(func(ifLogExists *jen.Group) {
				ifLogExists.Id("project").Dot("_log").Op("=").Id("project").Dot("_log").Index(jen.Empty(), jen.Lit(0))
			})
			txFunc.Id("project").Dot("_tx").Dot("Unlock").Call()
		}).Line()
	}
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
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id(fnName).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).Op("*").Id(model.Name).BlockFunc(func(indexFunc *jen.Group) {
			if proj.Transactional {
				indexFunc.Return().Id("project").Dot("storage").Dot("Get" + model.Name).Call(jen.Id(keyName))
			} else if proj.Synchronized {
				indexFunc.Id("project").Dot("_lock").Dot("RLock").Call()
				indexFunc.Defer().Id("project").Dot("_lock").Dot("RUnlock").Call()
				indexFunc.Return().Id("project").Dot("index" + indexName).Dot("Get" + model.Name).Call(jen.Id(keyName))
			} else {
				indexFunc.Return().Id("project").Dot("index" + indexName).Dot("Get" + model.Name).Call(jen.Id(keyName))
			}

		}).Line()
	}
	// sequence access methods
	for _, model := range proj.Models {
		for _, field := range model.AutoSequence {
			fName := "Next" + model.Name + field
			fs = fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id(fName).Params().Int64().BlockFunc(func(indexFunc *jen.Group) {
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
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id("Insert" + model.Name).Params(jen.Id("item").Op("*").Id(model.Name)).Op("*").Id(model.Name).BlockFunc(func(indexFunc *jen.Group) {

			for _, auto := range model.AutoSequence {
				indexFunc.Id("item").Dot(auto).Op("=").Id("project").Dot("Next" + model.Name + auto).Call()
			}
			indexFunc.Id("item").Dot("_project").Op("=").Id("project")
			if proj.Transactional {
				indexFunc.Id("project").Dot("_log").Op("=").Append(jen.Id("project").Dot("_log"), jen.Id(proj.Name+"LogEntity").ValuesFunc(func(logFn *jen.Group) {
					logFn.Id(model.Name).Op(":").Op("&").Id(model.Name + "LogEntity").ValuesFunc(func(modelLog *jen.Group) {
						modelLog.Id(model.Indexed).Op(":").Id("item").Dot(model.Indexed)
						modelLog.Id("Item").Op(":").Op("*").Id("item")
						modelLog.Id("Action").Op(":").Id(proj.Name + "ActionInsert")
					})
				}))
			} else if proj.Synchronized {
				indexFunc.Id("project").Dot("_lock").Dot("Lock").Call()
				indexFunc.Id("project").Dot("index"+indexName).Dot("Put"+model.Name).Call(jen.Id("item").Dot(model.Indexed), jen.Id("item"))
				indexFunc.Id("project").Dot("_lock").Dot("Unlock").Call()
			} else {
				indexFunc.Id("project").Dot("index"+indexName).Dot("Put"+model.Name).Call(jen.Id("item").Dot(model.Indexed), jen.Id("item"))
			}
			indexFunc.Return().Id("item")
		}).Line()
	}
	// update models (without assign sequences)
	for _, model := range proj.Models {
		indexName := model.Name + "By" + model.Indexed
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id("Update" + model.Name).Params(jen.Id("item").Op("*").Id(model.Name)).Op("*").Id(model.Name).BlockFunc(func(indexFunc *jen.Group) {
			if proj.Transactional {
				indexFunc.Id("project").Dot("_log").Op("=").Append(jen.Id("project").Dot("_log"), jen.Id(proj.Name+"LogEntity").ValuesFunc(func(logFn *jen.Group) {
					logFn.Id(model.Name).Op(":").Op("&").Id(model.Name + "LogEntity").ValuesFunc(func(modelLog *jen.Group) {
						modelLog.Id(model.Indexed).Op(":").Id("item").Dot(model.Indexed)
						modelLog.Id("Item").Op(":").Op("*").Id("item")
						modelLog.Id("Action").Op(":").Id(proj.Name + "ActionUpdate")
					})
				}))
			} else if proj.Synchronized {
				indexFunc.Id("project").Dot("_lock").Dot("Lock").Call()
				indexFunc.Id("project").Dot("index"+indexName).Dot("Update"+model.Name).Call(jen.Id("item").Dot(model.Indexed), jen.Id("item"))
				indexFunc.Id("project").Dot("_lock").Dot("Unlock").Call()
			} else {
				indexFunc.Id("project").Dot("index"+indexName).Dot("Update"+model.Name).Call(jen.Id("item").Dot(model.Indexed), jen.Id("item"))
			}
			indexFunc.Return().Id("item")
		}).Line()
	}
	// remove models (without following links)
	for _, model := range proj.Models {
		indexName := model.Name + "By" + model.Indexed
		keyName := memdata.ToLowerCamel(model.Indexed)
		fs = fs.Func().Parens(jen.Id("project").Op("*").Id("impl" + proj.Name)).Id("Remove" + model.Name).Params(jen.Id(keyName).Id(model.FieldType(model.Indexed))).BlockFunc(func(indexFunc *jen.Group) {
			if proj.Transactional {
				indexFunc.Id("project").Dot("_log").Op("=").Append(jen.Id("project").Dot("_log"), jen.Id(proj.Name+"LogEntity").ValuesFunc(func(logFn *jen.Group) {
					logFn.Id(model.Name).Op(":").Op("&").Id(model.Name + "LogEntity").ValuesFunc(func(modelLog *jen.Group) {
						modelLog.Id(model.Indexed).Op(":").Id(keyName)
						modelLog.Id("Action").Op(":").Id(proj.Name + "ActionDelete")
					})
				}))
			} else if proj.Synchronized {
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

func generateTransactionalDefines(proj *memdata.Project) jen.Code {
	var code = jen.Type().Id(proj.Name + "Action").Int().Line()
	code.Const().DefsFunc(func(defines *jen.Group) {
		defines.Id(proj.Name + "ActionUnknown").Id(proj.Name + "Action").Op("=").Lit(0)
		defines.Id(proj.Name + "ActionInsert").Id(proj.Name + "Action").Op("=").Lit(1)
		defines.Id(proj.Name + "ActionUpdate").Id(proj.Name + "Action").Op("=").Lit(2)
		defines.Id(proj.Name + "ActionDelete").Id(proj.Name + "Action").Op("=").Lit(3)
	}).Line()

	code.Type().Id(proj.Name + "LogEntity").StructFunc(func(group *jen.Group) {
		group.Comment("only of log entity should be filled")
		for _, model := range proj.Models {
			group.Id(model.Name).Op("*").Id(model.Name + "LogEntity")
		}
	}).Line()

	return code
}

func Generate(proj *memdata.Project) jen.Code {
	s := GenerateProject(proj)
	for _, md := range proj.Models {
		s = s.Line().Add(GenerateModel(md))
	}
	return s
}
