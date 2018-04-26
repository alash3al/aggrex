package main

import (
	"log"

	db "core.db"
	"core.globals"
	"core.server"
	_ "github.com/robertkrimen/otto/underscore"
)

func main() {
	globals.PopulateGlobals()

	log.Println("[AdminToken]", *globals.FlagAdminToken)

	log.Println("[DB]", "Initializing the database ...")
	dbh, err := db.Open(*globals.FlagIndexName)
	if err != nil {
		log.Fatal("[DB]", err.Error())
	}
	globals.DBHandler = dbh

	log.Println("[Server]", "Initializing the server ...")
	server.Serve(*globals.FlagHTTPAddr)
}
