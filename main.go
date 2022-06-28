package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/turboazot/helm-cache/pkg/services"
)

func main() {
	chartmuseumURL := os.Getenv("CHARTMUSEUM_URL")
	chartmuseumUsername := os.Getenv("CHARTMUSEUM_USERNAME")
	chartmuseumPassword := os.Getenv("CHARTMUSEUM_PASSWORD")

	if chartmuseumURL == "" || chartmuseumUsername == "" || chartmuseumPassword == "" {
		log.Fatal("CHARTMUSEUM_URL, CHARTMUSEUM_USERNAME, CHARTMUSEUM_PASSWORD environment variables are required")
		os.Exit(1)
	}

	c := services.NewCollector(chartmuseumURL, chartmuseumUsername, chartmuseumPassword)
	for {
		fmt.Println("Checking all helm secrets...")
		c.CheckAllSecrets()
		fmt.Println("Checking finished!")
		time.Sleep(time.Second * 10)
	}
}
