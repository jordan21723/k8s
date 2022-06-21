package main

import (
	"k8s-installer/cmd/server/app"
	"log"
)

func main() {
	//app.LoadConfigFileIfFound()
	command := app.NewServerCommand()
	if err := command.Execute(); err != nil {
		log.Fatal(err)
	}
}
