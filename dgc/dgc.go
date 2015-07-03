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
	}
	app.Run(os.Args)
}
