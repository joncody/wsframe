package main

import (
	"github.com/joncody/wsframe"
)

func main() {
	app := wsframe.NewApp("./config.json")
	app.Start()
}
