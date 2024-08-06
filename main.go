package main

import (
	"github.com/jumpei00/gostocktrade/app/models"
	"github.com/jumpei00/gostocktrade/app/server"
	"github.com/jumpei00/gostocktrade/config"
	"github.com/jumpei00/gostocktrade/log"
	"github.com/jumpei00/gostocktrade/nepse"
	"github.com/jumpei00/gostocktrade/scrape"
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
