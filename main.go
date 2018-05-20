package main

import (
	"fmt"
	"os"
	"time"

	"github.com/AdamJacobMuller/gonest/gonest"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var nest gonest.Nest

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:    "download-video",
			Aliases: []string{},
			Usage:   "download a series of video clips",
			Action:  DownloadVideo,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "id",
					Usage: "camera id",
				},
				cli.Int64Flag{
					Name: "start",
				},
				cli.Int64Flag{
					Name: "end",
				},
			},
		},
	}
	app.Run(os.Args)
}

func DownloadVideo(c *cli.Context) {
	start := c.Int64("start")
	if start == 0 {
		log.Fatal("start is required")
	}

	end := c.Int64("end")
	if end == 0 {
		log.Fatal("end is required")
	}

	id := c.String("id")
	if id == "" {
		log.Fatal("id is required")
	}

	nest.Load()
	nest.Login()
	nest.Save()

	for i := start; i < end; i += 3600 {
		clip, err := nest.CreateClip(id, time.Unix(i, 0), 3600)
		if err != nil {
			panic(err)
		}

		clip.Save(fmt.Sprintf("videos/%s-%d-%d.mp4", id, i, i+3600))
		clip.Delete()
	}
}
