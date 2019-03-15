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