package main

import (
	"github.com/jumpei00/gostocktrade/nepse"
	"github.com/jumpei00/gostocktrade/utils"
)

func mai1n() {
	data, err := nepse.LoadAllCsvFilesToMap("./data/date")
	if err != nil {
		panic(err)
	}
	err = utils.GenerateGoFile("./nepse/nepse_archive.go", "nepse", "archivedData", data)
	if err != nil {
		panic(err)
	}
}
