package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Credentials struct {
	username string
	password string
}

type dockerConfig struct {
	Auth map[string]struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Auth     string `json:"auth"`
	} `json:"auths"`
}

func expandUser(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(home, path[2:])
	}

	return path
}

func LoadCredentialsFromDockerConfig(path string) (map[string]Credentials, error) {
	path = filepath.Clean(path)
	path = expandUser(path)
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("got invalid path. Reason: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not load docker config. Reason: %v", err)
	}

	var df dockerConfig
	if err := json.Unmarshal(content, &df); err != nil {
		return nil, errors.New("docker config seems to have a unknown format")
	}

	hostToCredMapping := make(map[string]Credentials)

	for k, v := range df.Auth {
		host := k

		// some entries look like https://some.host/endpoint
		if reg, err := url.Parse(k); err == nil && reg.Host != "" {
			host = reg.Host
		}

		if v.Username != "" && v.Password != "" {
			hostToCredMapping[host] = Credentials{
				username: v.Username,
				password: v.Password,
			}
			continue
		}

		if v.Auth != "" {

			data, err := base64.StdEncoding.DecodeString(v.Auth)
			if err != nil {
				return nil, fmt.Errorf("Invalid login in auth of %s: %v", k, err)
			}
			username, password, _ := strings.Cut(string(data), ":")

			hostToCredMapping[host] = Credentials{
				username: username,
				password: password,
			}
			continue
		}

		// TODO: use debug logging
		fmt.Println("Unusable auth entry in docker config")
	}

	return hostToCredMapping, nil
}
