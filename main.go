package main

import (
	"fmt"
	"log"
	"os"
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

	//resp, err := r.request("v2/_catalog")
	r.GetTags("library/alpine")

}
