package main

import (
	"fmt"
	"os"
	"os/exec"
	"github.com/codegangsta/cli"
)

func runDgc(c *cli.Context) {
	fmt.println("Hello Test")
}

func main() {
	app := cli.NewApp()
	dgc.Name = "dgc"
	dgc.Usage = "A minimal docker garbage collector"
	dgc.Version = "0.1.0"
	dgc.Author = "David J Felix <davidjfelix@davidjfelix.com>"
	dgc.Action = runDgc
	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name: "grace, g",
			Value: "3600",
			Usage: "the grace period for a container, defualt time unit is seconds",
			EnvVar: "GRACE_PERIOD_SECONDS,GRACE_PERIOD",
		},
		cli.StringFlag {
			Name: "time-unit, t",
			Value: "s",
			Usage: "the time unit used for the grace period",
			EnvVar: "GRACE_PERIOD_TIME_UNIT,TIME_UNIT",
		},
		cli.StringFlag {
			Name: "docker, d",
			Value: "docker",
			Usage: "the docker executable",
			EnvVar: "DOCKER",
		},
		cli.StringFlag {
			Name: "exclude, e",
			Value: "/etc/docker-gc-exclude",
			Usage: "the directory of the list of containers to exclude from garbage collection",
			EnvVar: "EXCLUDE_FROM_GC",
		}
	}
	app.Run(os.Args)
}
