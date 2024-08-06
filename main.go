package main

import (
	"github.com/oarkflow/nepse/app/models"
	"github.com/oarkflow/nepse/app/server"
	"github.com/oarkflow/nepse/config"
	"github.com/oarkflow/nepse/log"
	"github.com/oarkflow/nepse/nepse"
	"github.com/oarkflow/nepse/scrape"
)

func main() {
	go func() {
		nepse.InitCSVStock()
		scrape.Scrape()
	}()
	config.InitConfig()
	log.SetLogging()
	models.InitDB()
	server.Run()
}
