package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	credMap, err := LoadCredentialsFromDockerConfig("~/.docker/config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	host := "index.docker.io"

	credentials, found := credMap[host]
	if !found {
		log.Fatalf("No creds for registry %s", host)
	}

	r, err := NewRegisty(host, credentials)
	if err != nil {
		log.Fatalln(err)
	}

	//resp, err := r.request("v2/_catalog")
	tags, err := r.GetTags("library/alpine")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(tags)

	before := time.Now()
	for _, tag := range tags {
		r.GetManifest("library/alpine", tag)
	}

	fmt.Println(time.Since(before))
}
