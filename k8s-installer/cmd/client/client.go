package main

import (
	"k8s-installer/cmd/client/app"
	"log"
	"math/rand"
	"time"
)

func main() {

	rand.Seed(time.Now().UnixNano())
	app.LoadConfigFileIfFound()
	command := app.NewServerCommand()
	if err := command.Execute(); err != nil {
		log.Fatal(err)
	}
}
