package main

import (
	"log"

	"github.com/samertm/srclib-docco"
)

// Some comment

func main() {
	if err := srclib_docco.Main(); err != nil {
		log.Fatal(err)
	}
}
