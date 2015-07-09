package main

import (
	"fmt"
	"os"
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
)

func runDgc(c *cli.Context) {
    client, _ := docker.NewClient(c.String("socket"))
    imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})
    for _, img := range imgs {
        fmt.Println("ID: ", img.ID)
        fmt.Println("RepoTags: ", img.RepoTags)
        fmt.Println("Created: ", img.Created)
        fmt.Println("Size: ", img.Size)
        fmt.Println("VirtualSize: ", img.VirtualSize)
        fmt.Println("ParentId: ", img.ParentID)
    }
}

func main() {
	dgc := cli.NewApp()
	dgc.Name = "dgc"
	dgc.Usage = "A minimal docker garbage collector"
	dgc.Version = "0.1.0"
	dgc.Author = "David J Felix <davidjfelix@davidjfelix.com>"
	dgc.Action = runDgc
	dgc.Flags = []cli.Flag {
		cli.StringFlag {
			Name: "grace, g",
			Value: "3600s",
			Usage: "the grace period for a container. Accepted compostable time units: [h, m, s, ms, ns us]",
			EnvVar: "GRACE_PERIOD_SECONDS,GRACE_PERIOD",
		},
		cli.StringFlag {
			Name: "socket, s",
			Value: "unix:///var/run/docker.sock",
			Usage: "the docker remote socket",
			EnvVar: "DOCKER_SOCKET",
		},
		cli.StringFlag {
			Name: "exclude, e",
			Value: "/etc/docker-gc-exclude",
			Usage: "the list of containers to exclude from garbage collection, as a file or directory",
			EnvVar: "EXCLUDE_FROM_GC",
		},
	}
	dgc.Run(os.Args)
}
