package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type tokenResponse struct {
	Token    string    `json:"token"`
	Scope    string    `json:"scope"`
	Validity int       `json:"expires_in"`
	Issued   time.Time `json:"issued_at"`
}

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

type tagListResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type manifestResponse struct {
	SchemaVersion int    `json:"schemaVersion"`
	Name          string `json:"name"`
	Tag           string `json:"tag"`
	Architechture string `json:"architecture"`
	FsLayers      []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`
	History []struct {
		V1Compatibility string `json:"v1Compatibility"`
	} `json:"history"`
	// NOTE: signatures is ignored at the moment
}

type Credentials struct {
	username string
	password string
}

type Registry struct {
	host        string
	credentials Credentials
}

type Manifest struct {
	Name   string
	Tag    string
	Layers []string
}

var ErrImageDoesNotExist = errors.New("image does not exist")
var ErrResourceDoesNotExist = errors.New("resource does not exist")
var ErrNotAllowedOrUnavailable = errors.New("your're either not allowed to access this resource or it does not exist")

func extractOAuthSettings(s string) (string, string, string) {
	// Bearer realm="...",service="...",scope=""

	var realm, service, scope string

	_, rest, _ := strings.Cut(s, " ")

	for _, item := range strings.Split(rest, ",") {
		// item == x=".."
		key, value, _ := strings.Cut(item, "=")
		value = value[1 : len(value)-1]

		switch key {
		case "realm":
			realm = value
		case "service":
			service = value
		case "scope":
			if scope == "" {
				scope = value
			} else {
				scope += "," + value
			}

		}
	}

	return realm, service, scope
}

func optainToken(authenticate string, creds *Credentials) (*tokenResponse, error) {
	// https://stackoverflow.com/questions/56193110/how-can-i-use-docker-registry-http-api-v2-to-obtain-a-list-of-all-repositories-i/68654659#68654659
	// https://docs.docker.com/registry/spec/auth/token/

	realm, service, scope := extractOAuthSettings(authenticate)

	authUrl, err := url.Parse(realm)
	if err != nil {
		return nil, err
	}

	values := make(url.Values)
	values.Add("service", service)
	values.Add("grant_type", "password")
	values.Add("client_id", "dockerengine")
	values.Add("scope", scope)
	values.Add("username", creds.username)
	values.Add("password", creds.password)

	authUrl.RawQuery = values.Encode()

	// to access private docker-hub repos this is required
	authUrl.User = url.UserPassword(creds.username, creds.password)

	resp, err := http.Get(authUrl.String())

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("could not authenticate with password")
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tResp tokenResponse
	err = json.Unmarshal(content, &tResp)
	if err != nil {
		return nil, errors.New("could not deserialize token response")
	}

	return &tResp, nil
}

func (r *Registry) request(endpoint string) (*http.Response, error) {
	host := strings.TrimSuffix(r.host, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	targetUrl := fmt.Sprintf("https://%s/%s", host, endpoint)
	// Head instead of Get because we have to optain a token anyways
	// on https registries a login is required. Meaning we expect a 401
	// type of response
	resp, err := http.Head(targetUrl)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	if resp.StatusCode != 401 || resp.Header.Get("Www-authenticate") == "" {
		return nil, fmt.Errorf("expected authentication request but got response status: %s", resp.Status)
	}

	token, err := optainToken(resp.Header.Get("Www-authenticate"), &r.credentials)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	request.Header.Set("Accept-Encoding", "*")

	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		resp.Body.Close()
		// NOTE: in the Www-authenticate should be an error field that says so
		return nil, ErrNotAllowedOrUnavailable
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		return nil, ErrResourceDoesNotExist
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("unknown error: %s", resp.Status)
	}

	return resp, nil
}

func (r *Registry) GetCatalog() ([]string, error) {
	resp, err := r.request("v2/_catalog")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cResp catalogResponse
	if err = json.Unmarshal(content, &cResp); err != nil {
		return nil, err
	}

	return cResp.Repositories, nil
}

func (r *Registry) GetTags(imageName string) ([]string, error) {
	imageName = strings.TrimSuffix(imageName, "/")
	imageName = strings.TrimPrefix(imageName, "/")
	resp, err := r.request(fmt.Sprintf("v2/%s/tags/list", imageName))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, ErrImageDoesNotExist
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tagsResp tagListResponse
	if err = json.Unmarshal(content, &tagsResp); err != nil {
		return nil, err
	}

	return tagsResp.Tags, nil
}

func (r *Registry) GetManifest(imageName string, tag string) (*Manifest, error) {
	imageName = strings.TrimSuffix(imageName, "/")
	imageName = strings.TrimPrefix(imageName, "/")
	resp, err := r.request(fmt.Sprintf("v2/%s/manifests/%s", imageName, tag))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mResp manifestResponse
	if err = json.Unmarshal(content, &mResp); err != nil {
		return nil, err
	}

	layers := make([]string, len(mResp.FsLayers))
	for _, item := range mResp.FsLayers {
		layers = append(layers, item.BlobSum)
	}
	return &Manifest{
		Name:   mResp.Name,
		Tag:    mResp.Tag,
		Layers: layers,
	}, nil
}
