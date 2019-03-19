# Go code-generator for structures

Install everything by go get:

`go get -u github.com/reddec/memdata/cmd/...`

## RB-Tree


`go get -u github.com/reddec/memdata/cmd/rbtree`

Adapted version from https://github.com/emirpasic/gods#redblacktree for code generating.

* Optimization for number-based key
* Supports `.Cmp` method for complex objects
* Exposed `Lookup` method for manipulating value on-place



### CLI

    Usage:
      rbtree [OPTIONS]
    
    Application Options:
          --package= package name in generated file (default: tree) [$PACKAGE]
          --import=  imports included to the generated file [$IMPORTS]
          --type=    tree type name (default: tree) [$TYPENAME]
          --key=     key typename (default: int64) [$KEY]
          --value=   value typename [$VALUE]
          --cmp      user Cmp method to compare keys
    
    Help Options:
      -h, --help     Show this help message
      
## Memory typed storage

`go get -u github.com/reddec/memdata/cmd/memdata`

Generates interfaces and default implementation to store user-defined objects. Allows to specify any storage.

* Default in-memory implementation
* Vendor independent storage drivers
* Statically typed, no `interface{}` and runtime casts
* Supports synchronized and non-synchronized access
* Separated read/write locks (multiple readers, one writer) in transactional mode

See examples in `generator/model/example` folder for yaml definition.

### YAML

YAML files are required to define user models. There are two types of configuration files:

1. project file ('main' yaml)
2. model files (optional, only if you prefer separate model definition from project configuration)


**project** yaml (see generator/model/example/sample.yaml)

*   **name** - name of main generated structure
*   **package** - result package name
*   **synchronized** (boolean, default false) - generate global locks or not
*   **imports** (map of alias->import path) - use specified imports for lookup of custom types
*   **models** (list of model definition) - list of user models
*   **include_models** (list of string) - list of files of model definition relative to the current file
*   **storage_ref** (bool, default false) - add storage reference to the generated models
*   **transactional** (boolean, default false) - copy changes and apply as batch on commit

**model** yaml / definition

* **name** - name of model/structure
* **fields** (map of string->string) - map of fields and types
    - if field names starts from `$` (dollar sign) it's mean it is reference to another model
    (aka alias for `ref` field)
    - if field ends with `...` (three dots) it's mean it is multiple reference to another model
    (aka alias for `many` field)
 * **indexed** (required, string) - field name that unique identified model, can be omitted if `key` specified (see below)
 * **ref** (map, string->string) - field name and name of another model as reference (many-to-one)
 * **many** (map, string->string) - field name and name of another model as multiple reference (many-to-many); 
 think about it as array of ref to another models
 * **sequence** (list of string) - name fields that acts as sequences with automatic increment after insertion (field should be int64 and defined in `fields`)
 * **key** (string) - name primary key in model. Automatically defines `indexed` and `sequence` (if key is number)
 
 ### CLI
 
 
     Usage:
       memdata [OPTIONS] file
     
     Help Options:
       -h, --help  Show this help message
     
     Arguments:
       file:       path to project YAML file
