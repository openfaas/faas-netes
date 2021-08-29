# Tools folder

The tools folder allows us a space to define and pin external dependencies. If they are go based tools we can create individual `mod` files that allow us to download or install these tools independent of the main package.

## Tools

1. `k8s.io/code-gen` this package needs to be _downloaded_ not installed. But we can not use the copy created in the `vendor` folder because vendor does not make a complete clone, it only keeps the Go files, and the code-gen project has several bash scripts that we need to reference.
The main project `Makefile` will attempt to keep the `code-generator.mod` file in sync with the `go.mod`. It should not need to be manually edited, but it does need to be committed.