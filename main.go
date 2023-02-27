package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	credMap, err := LoadCredentialsFromDockerConfig("~/.docker/config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	host := "docker.lp.rsint.net"

	credentials, found := credMap[host]
	if !found {
		log.Fatalf("No creds for registry %s", host)
	}

	r, err := NewRegisty(host, credentials)
	if err != nil {
		log.Fatalln(err)
	}

	image := "comvidence/com-on-k8s/preview/ci/int"
	//resp, err := r.request("v2/_catalog")
	//tags, err := r.GetTags("library/alpine")
	tags, err := r.GetTags(image)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(tags)

	before := time.Now()

	var wg sync.WaitGroup
	wg.Add(len(tags))
	for _, tag := range tags {
		go func(tag string) {
			r.GetManifest(image, tag)
			fmt.Println(image, tag)
			wg.Done()
		}(tag)
	}
	wg.Wait()
	fmt.Println(time.Since(before))
}
