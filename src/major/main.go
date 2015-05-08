package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = ""
	app.Action = RunMajor
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			Value:  "localhost",
			Usage:  "Hostname of MongoDB",
			EnvVar: "MONGO_PORT_27017_TCP_ADDR",
		},
		cli.StringFlag{
			Name:  "database",
			Value: "major",
			Usage: "Database",
		},
		cli.IntFlag{
			Name:   "port",
			Value:  27017,
			Usage:  "Port of MongoDB",
			EnvVar: "MONGO_PORT_27017_TCP_PORT",
		},
	}

	app.Run(os.Args)
}
