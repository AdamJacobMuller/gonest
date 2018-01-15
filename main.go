package main

import (
	"github.com/AdamJacobMuller/gonest/gonest"

	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	n := gonest.Nest{}
	n.Load()
	n.Login()
	n.Save()

	start, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		panic(err)
	}

	clip, err := n.CreateClip(os.Args[1], time.Unix(start, 0), 3600)
	if err != nil {
		panic(err)
	}

	clip.Save(fmt.Sprintf("videos/%d.mp4", start))
	clip.Delete()
}
