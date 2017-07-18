package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli"
	dockerTypes "github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"log"
	"os"
	"sync"
	"time"
	"context"
)

func readExcludes(fileName string) []string {
	var excludeNames []string
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Error opening input file:", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		excludeNames = append(excludeNames, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading exclude file:", scanner.Err())
	}
	return excludeNames
}

func collectAPIImages(images []dockerTypes.ImageSummary, client *dockerClient.Client, ctx *cli.Context, excludes []string) {
	var imageSync sync.WaitGroup
	grace := ctx.Duration("grace")
	quiet := ctx.Bool("quiet")
	options := dockerTypes.ImageRemoveOptions{
		Force:         ctx.Bool("force"),
		PruneChildren: ctx.Bool("no-prune"),
	}

	for _, image := range images {
		imageSync.Add(1)
		go func(image dockerTypes.ImageSummary) {
			defer imageSync.Done()

			// Check if the image id or tag is on excludes list
			for _, excludeName := range excludes {
				if image.ID == excludeName {
					return
				}
				for _, tag := range image.RepoTags {
					if tag == excludeName {
						return
					}
				}
			}

			// End if the image is still in the grace period
			log.Printf("Inspecting image: %s\n", image.ID)

			now := time.Now()
			if now.Sub(time.Unix(image.Created, 0)) < grace {
				return
			}

			// Delete image
			log.Printf("Deleting image: %s\n", image.ID)

			if _, err := client.ImageRemove(context.Background(), image.ID, options); err == nil {
				log.Printf("Deleted image: %s\n", image.ID)
				if !quiet {
					fmt.Printf("Deleted image: %s\n", image.ID)
				}
			} else {
				log.Printf("Error. Failed to delete image: %s\n", image.ID)
				return
			}
		}(image)
	}

	imageSync.Wait()
}

func collectAPIContainers(containers []dockerTypes.Container, client *dockerClient.Client, ctx *cli.Context, excludes []string) {
	var containerSync sync.WaitGroup
	grace := ctx.Duration("grace")
	quiet := ctx.Bool("quiet")

	for _, container := range containers {
		containerSync.Add(1)
		go func(container dockerTypes.Container) {
			defer containerSync.Done()

			// Check if the container id or tag is on excludes list
			for _, excludeName := range excludes {
				if container.ID == excludeName {
					return
				}
				if container.Image == excludeName {
					return
				}
				for _, name := range container.Names {
					if name == excludeName {
						return
					}
				}
			}

			// End if the container is still in the grace period
			now := time.Now()
			if now.Sub(time.Unix(container.Created, 0)) < grace {
				return
			}

			// Delete container
			options := dockerTypes.ContainerRemoveOptions{
				RemoveVolumes: ctx.Bool("remove-volumes"),
				Force:         ctx.Bool("force"),
			}

			log.Printf("Deleting container: %s\n", container.ID)

			if err := client.ContainerRemove(context.Background(), container.ID, options); err == nil {
				log.Printf("Deleted container: %s\n", container.ID)
				if !quiet {
					fmt.Printf("Deleted container: %s\n", container.ID)
				}
			} else {
				log.Printf("Error. Failed to delete container: %s\n", container.ID)
				return
			}
		}(container)
	}

	containerSync.Wait()
}

func runDgc(ctx *cli.Context) {
	var dgcSync sync.WaitGroup
	var excludes []string

	// TODO: change this to use socket
	client, err := dockerClient.NewEnvClient()
	if err != nil {
		log.Fatalf("Error. Failed to create a docker client to: %s", ctx.String("socket"))
	}

	log.Println("Getting a List of images...")
	images, err := client.ImageList(context.Background(), dockerTypes.ImageListOptions{All: true})
	if err != nil {
		log.Fatal("Error. Failed to retrieve images from the docker host.")
	}

	log.Println("Getting a list of containers...")
	containers, err := client.ContainerList(context.Background(), dockerTypes.ContainerListOptions{All: true})
	if err != nil {
		log.Fatal("Error. Failed to retrieve containers from the docker host.")
	}
	if ctx.String("exclude") != "" {
		excludes = readExcludes(ctx.String("exclude"))
	}

	dgcSync.Add(2)
	log.Println("Performing garbage collection...")
	go func() {
		defer dgcSync.Done()
		collectAPIContainers(containers, client, ctx, excludes)
	}()
	go func() {
		defer dgcSync.Done()
		collectAPIImages(images, client, ctx, excludes)
	}()
	dgcSync.Wait()
	log.Println("Finished garbage collection!")
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
			Value:  "",
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
