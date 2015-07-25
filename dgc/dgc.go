package main

import (
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	"os"
	"sync"
	"time"
)

func collectAPIImages(images []docker.APIImages, client *docker.Client, ctx *cli.Context) {
	var imageSync sync.WaitGroup
	for _, image := range images {
		imageSync.Add(1)
		go func(image *docker.APIImages, client *docker.Client, grace time.Duration, quiet bool, force bool, noPrune bool) {
			defer imageSync.Done()
			imageDetail, _ := client.InspectImage(image.ID)
			handleImage(imageDetail, client, grace, quiet, force, noPrune)
		}(&image, client, ctx.Duration("grace"), ctx.Bool("quiet"), ctx.Bool("force"), ctx.Bool("no-prune"))
	}
	imageSync.Wait()
}

func handleImage(image *docker.Image, client *docker.Client, grace time.Duration, quiet bool, force bool, noPrune bool) {
	now := time.Now()
	options := docker.RemoveImageOptions{
		Force:   force,
		NoPrune: noPrune,
	}
	if now.Sub(image.Created) >= grace {
		client.RemoveImageExtended(image.ID, options)
	}
}

func collectAPIContainers(containers []docker.APIContainers, client *docker.Client, ctx *cli.Context) {
	var containerSync sync.WaitGroup
	for _, container := range containers {
		containerSync.Add(1)
		go func(container *docker.APIContainers, client *docker.Client, grace time.Duration, quiet bool, force bool, removeVolumes bool) {
			defer containerSync.Done()
			containerDetail, _ := client.InspectContainer(container.ID)
			handleContainer(containerDetail, client, grace, quiet, force, removeVolumes)
		}(&container, client, ctx.Duration("grace"), ctx.Bool("quiet"), ctx.Bool("force"), ctx.BoolT("remove-volumes"))
	}
	containerSync.Wait()
}

func handleContainer(container *docker.Container, client *docker.Client, grace time.Duration, quiet bool, force bool, removeVolumes bool) {
	now := time.Now()
	options := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: removeVolumes,
		Force:         force,
	}
	if now.Sub(container.Created) >= grace {
		client.RemoveContainer(options)
	}
}

func runDgc(ctx *cli.Context) {
	var dgcSync sync.WaitGroup
	client, _ := docker.NewClient(ctx.String("socket"))
	images, _ := client.ListImages(docker.ListImagesOptions{All: false})
	containers, _ := client.ListContainers(docker.ListContainersOptions{All: false})
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
