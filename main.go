package main

import (
	"fmt"
	"log"
	"os"
	"strings"
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

	r := Registry{host: host, credentials: credentials}
	//resp, err := r.request("v2/_catalog")
	tags, err := r.GetTags("library/alpine")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(tags)

	fmt.Println(strings.Repeat("-", 40))

	manifest, err := r.GetManifest("library/alpine", "latest")
	fmt.Println(manifest, err)

}
