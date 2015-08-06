package main

import (
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	"fmt"
	"os"
	"sync"
	"time"
)

func collectAPIImages(images []docker.APIImages, client *docker.Client, ctx *cli.Context) {
	var imageSync sync.WaitGroup
	grace := ctx.Duration("grace")
	quiet := ctx.Bool("quiet")
	options := docker.RemoveImageOptions{
		Force:   ctx.Bool("force"),
		NoPrune: ctx.Bool("no-prune"),
	}
	for _, image := range images {
		imageSync.Add(1)
		go func(image docker.APIImages) {
			defer imageSync.Done()
			imageDetail, _ := client.InspectImage(image.ID)
			now := time.Now()
			if now.Sub(imageDetail.Created) >= grace {
				if err := client.RemoveImageExtended(imageDetail.ID, options); err == nil {
					if !quiet {
						fmt.Printf("Deleting image: %s.\n", imageDetail.ID)
					}
				}
			}
		}(image)
	}
	imageSync.Wait()
}

func collectAPIContainers(containers []docker.APIContainers, client *docker.Client, ctx *cli.Context) {
	var containerSync sync.WaitGroup
	grace := ctx.Duration("grace")
	quiet := ctx.Bool("quiet")
	for _, container := range containers {
		containerSync.Add(1)
		go func(container docker.APIContainers){
			defer containerSync.Done()
			containerDetail, _ := client.InspectContainer(container.ID)
			now := time.Now()
			if now.Sub(containerDetail.Created) >= grace {
				options := docker.RemoveContainerOptions{
					ID:            containerDetail.ID,
					RemoveVolumes: ctx.Bool("remove-volumes"),
					Force:         ctx.Bool("force"),
				}
				if err := client.RemoveContainer(options); err == nil {
					if !quiet {
						fmt.Printf("Deleting container: %s.\n", containerDetail.ID)
					}
				}
			}
		}(container)
	}
	containerSync.Wait()
}

func runDgc(ctx *cli.Context) {
	var dgcSync sync.WaitGroup
	client, _ := docker.NewClient(ctx.String("socket"))
	images, _ := client.ListImages(docker.ListImagesOptions{All: true})
	containers, _ := client.ListContainers(docker.ListContainersOptions{All: true})
	dgcSync.Add(2)
	go func(containers []docker.APIContainers, client *docker.Client, ctx *cli.Context) {
		defer dgcSync.Done()
		collectAPIContainers(containers, client, ctx)
	}(containers, client, ctx)
	go func(images []docker.APIImages, client *docker.Client, ctx *cli.Context) {
		defer dgcSync.Done()
		collectAPIImages(images, client, ctx)
	}(images, client, ctx)
	dgcSync.Wait()
}

func main() {
	dgc := cli.NewApp()
	dgc.EnableBashCompletion = true
	dgc.Name = "dgc"
	dgc.Usage = "A minimal docker garbage collector"
	dgc.Version = "0.1.0"
	dgc.Author = "David J Felix <davidjfelix@davidjfelix.com>"
	dgc.Action = runDgc
	dgc.Flags = []cli.Flag{
		cli.DurationFlag{
			Name:   "grace, g",
			Value:  time.Duration(3600) * time.Second,
			Usage:  "the grace period for a container. Accepted compostable time units: [h, m, s, ms, ns us]",
			EnvVar: "GRACE_PERIOD_SECONDS,GRACE_PERIOD",
		},
		cli.StringFlag{
			Name:   "socket, s",
			Value:  "unix:///var/run/docker.sock",
			Usage:  "the docker remote socket",
			EnvVar: "DOCKER_SOCKET",
		},
		cli.StringFlag{
			Name:   "exclude, e",
			Value:  "/etc/docker-gc-exclude",
			Usage:  "the list of containers to exclude from garbage collection, as a file or directory",
			EnvVar: "EXCLUDE_FROM_GC",
		},
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: "don't print name of garbage-collected containers",
		},
		cli.BoolTFlag{
			Name:  "remove-volumes, r",
			Usage: "remove volumes with the container",
		},
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "force images and containers to stop and be collected",
		},
		cli.BoolFlag{
			Name:  "no-prune, n",
			Usage: "don't prune parent images to a GC'd image",
		},
	}
	dgc.Run(os.Args)
}
