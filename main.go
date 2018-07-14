package main

import (
	"log"
	"strings"

	db "core.db"
	"core.globals"
	"core.server"
	vm "core.vm"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/robfig/cron"
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
	globals.CronKernel = cron.New()

	go (func() {
		cronRunner(globals.CronKernel)
	})()

	go (func() {
		for {
			select {
			case <-dbh.CronReload:
				cronRunner(globals.CronKernel)
			}
		}
	})()

	log.Println("[Server]", "Initializing the server ...")
	server.Serve(*globals.FlagHTTPAddr)
}

func cronRunner(cronKernel *cron.Cron) {
	cronKernel.Stop()
	for _, c := range globals.DBHandler.CronsGet() {
		cronKernel.AddFunc(c.Interval, func() {
			log.Println("Executing: ", c.Job)
			log.Println(vm.New(vm.VM{
				AllowedHosts: strings.Split(*globals.FlagAllowedHosts, ","),
				MaxExecTime:  *globals.FlagMaxExecTime,
				Request:      nil,
			}).Exec("(" + c.Job + ")" + "()"))
		})
	}
	cronKernel.Start()
}
