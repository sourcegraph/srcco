package main

import (
	"log"

	"github.com/samertm/srclib-docco"
)

func main() {
	if err := srclib_docco.Main(); err != nil {
		log.Fatal(err)
	}
}
