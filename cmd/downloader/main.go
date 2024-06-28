package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joesiltberg/gophercon2024finalexercise/downloader"
)

func main() {
	var destination = flag.String("destination", "", "The path to the destination file")
	var url = flag.String("url", "", "URL of the resource to download")
	var procs = flag.Int("procs", 4, "Number of parallel processes for the download")
	var chunkSize = flag.Int64("chunk-size", 1024*1024, "Size to download per HTTP request")

	flag.Parse()

	if *destination == "" {
		fmt.Fprintf(os.Stderr, "Please specify a destination file.\n")
		flag.Usage()
		os.Exit(1)
	}

	if *url == "" {
		fmt.Fprintf(os.Stderr, "Please specify a URL.\n")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	err := downloader.ParallelDownload(ctx, *destination, *url, *procs, *chunkSize)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to download %s : %s", *url, err.Error())
		os.Exit(2)
	}
}
