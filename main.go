package main

import (
	"encoding/json"
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
			Name:    "download-clip",
			Aliases: []string{},
			Usage:   "download a specific clip",
			Action:  DownloadClip,
			Flags: []cli.Flag{
				cli.Int64Flag{
					Name:  "id",
					Usage: "clip id",
				},
				cli.StringFlag{
					Name:  "filename",
					Usage: "filename to save clip to",
				},
			},
		},
		{
			Name:    "download-clips",
			Aliases: []string{},
			Usage:   "download all clips",
			Action:  DownloadClips,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "directory",
					Usage: "directory to save clips to",
				},
			},
		},
		{
			Name:    "download-video",
			Aliases: []string{},
			Usage:   "download video by creating and deleting clips",
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
		{
			Name:    "list-clips",
			Aliases: []string{},
			Usage:   "list video clips",
			Action:  ListClips,
			Flags:   []cli.Flag{},
		},
	}
	app.Run(os.Args)
}

func DownloadClips(c *cli.Context) {
	nest.Load()
	nest.Login()
	nest.Save()

	directory := c.String("directory")
	if directory == "" {
		log.Fatal("directory is required")
	}

	log.Info("listing clips")
	clips, err := nest.ListClips()
	if err != nil {
		panic(err)
	}

	for _, clip := range clips {
		filename := fmt.Sprintf("%s/%s", directory, clip.Filename)
		log.WithFields(log.Fields{
			"filename": filename,
			"title":    clip.Title,
		}).Info("saving clip")
		clip.Save(filename)
	}

}
func ListClips(c *cli.Context) {
	nest.Load()
	nest.Login()
	nest.Save()

	clips, err := nest.ListClips()
	if err != nil {
		panic(err)
	}

	out, err := json.MarshalIndent(clips, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", out)
}

func DownloadClip(c *cli.Context) {
	id := c.Int("id")
	if id == 0 {
		log.Fatal("id is required")
	}

	filename := c.String("filename")
	if filename == "" {
		filename = fmt.Sprintf("videos/%d.mp4", id)
	}

	nest.Load()
	nest.Login()
	nest.Save()

	clips, err := nest.ListClips()
	if err != nil {
		panic(err)
	}
	for _, clip := range clips {
		if clip.ID == id {
			clip.Save(filename)
		}
	}
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
