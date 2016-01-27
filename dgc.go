package main

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"sync"
	"time"
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

func collectAPIImages(images []docker.APIImages, client *docker.Client, ctx *cli.Context, excludes []string) {
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

			imageDetail, err := client.InspectImage(image.ID)
			if err != nil {
				log.Printf("Error. Failed to inspect image: %s\n", image.ID)
				return
			}
			now := time.Now()
			if now.Sub(imageDetail.Created) < grace {
				return
			}

			// Delete image
			log.Printf("Deleting image: %s\n", imageDetail.ID)

			if err := client.RemoveImageExtended(imageDetail.ID, options); err == nil {
				log.Printf("Deleted image: %s\n", imageDetail.ID)
				if !quiet {
					fmt.Printf("Deleted image: %s\n", imageDetail.ID)
				}
			} else {
				log.Printf("Error. Failed to delete image: %s\n", imageDetail.ID)
				return
			}
		}(image)
	}

	imageSync.Wait()
}

func collectAPIContainers(containers []docker.APIContainers, client *docker.Client, ctx *cli.Context, excludes []string) {
	var containerSync sync.WaitGroup
	grace := ctx.Duration("grace")
	quiet := ctx.Bool("quiet")

	for _, container := range containers {
		containerSync.Add(1)
		go func(container docker.APIContainers) {
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
			log.Printf("Inspecting container: %s\n", container.ID)

			containerDetail, err := client.InspectContainer(container.ID)
			if err != nil {
				log.Printf("Error. Failed to inspect container: %s\n", container.ID)
				return
			}
			now := time.Now()
			if now.Sub(containerDetail.Created) < grace {
				return
			}

			// Delete container
			options := docker.RemoveContainerOptions{
				ID:            containerDetail.ID,
				RemoveVolumes: ctx.Bool("remove-volumes"),
				Force:         ctx.Bool("force"),
			}

			log.Printf("Deleting container: %s\n", containerDetail.ID)

			if err := client.RemoveContainer(options); err == nil {
				log.Printf("Deleted container: %s\n", containerDetail.ID)
				if !quiet {
					fmt.Printf("Deleted container: %s\n", containerDetail.ID)
				}
			} else {
				log.Printf("Error. Failed to delete container: %s\n", containerDetail.ID)
				return
			}
		}(container)
	}

	containerSync.Wait()
}

func runDgc(ctx *cli.Context) {
	var dgcSync sync.WaitGroup
	var excludes []string

	client, err := docker.NewClient(ctx.String("socket"))
	if err != nil {
		log.Fatal("Error. Failed to create a docker client to: %s", ctx.String("socket"))
	}

	log.Println("Getting a List of images...")
	images, err := client.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		log.Fatal("Error. Failed to retrieve images from the docker host.")
	}

	log.Println("Getting a list of containers...")
	containers, err := client.ListContainers(docker.ListContainersOptions{All: true})
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
