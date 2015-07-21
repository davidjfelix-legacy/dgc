package main

import (
	"fmt"
	"errors"
	"os"
	"time"
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
)

func collectAPIImages(images []APIImages, client *Client, ctx *cli.Context) (error) {
}

func handleImage(image *Image, client *Client, grace *time.Duration, bool output) (error) {
}

func collectAPIContainers(containers []APIContainers, client *Client, ctx *cli.Context) (error) {
}

func handleContainer(container *Container, client *Client, grace *time.Duration, bool output) (error) {
}

func runDgc(ctx *cli.Context) {
	client, _ := docker.NewClient(ctx.String("socket"))
	images, _ := client.ListImages(docker.ListImagesOptions{All: false})
	containers, _ := client.ListContainers(docker.ListContainersOptions{All: false})
	collectAPIContainers(containers, &client, ctx)
	collectAPIImages(containers, &client, ctx)
}

func main() {
	dgc := cli.NewApp()
	dgc.EnableBashCompletion = true
	dgc.Name = "dgc"
	dgc.Usage = "A minimal docker garbage collector"
	dgc.Version = "0.1.0"
	dgc.Author = "David J Felix <davidjfelix@davidjfelix.com>"
	dgc.Action = runDgc
	dgc.Flags = []cli.Flag {
		cli.DurationFlag {
			Name: "grace, g",
			Value: time.Duration(3600)*time.Second,
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
		cli.BoolFlag {
			Name: "verbose, v",
			Usage: "print name of garbage-collected containers",
			EnvVar: "VERBOSE",
		},
	}
	dgc.Run(os.Args)
}
