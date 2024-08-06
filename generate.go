package main

import (
	"github.com/oarkflow/nepse/nepse"
	"github.com/oarkflow/nepse/utils"
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
