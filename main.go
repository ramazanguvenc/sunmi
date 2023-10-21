package main

import (
	"flag"
	"fmt"
	"logging"
	"os"
	"videorequest"
)

func main() {
	var URL string
	var destination string
	var debug bool

	flag.StringVar(&URL, "url", "", "URL of the video")
	flag.StringVar(&destination, "destination", "video.mp4", "Destination file path")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")

	flag.Parse()

	if debug {
		logging.EnableDebugMode()
	}

	if URL == "" {
		fmt.Println("Please provide a valid URL using the -url flag.")
		os.Exit(1)
	}

	videorequest.GetVideo(URL, destination)

}
