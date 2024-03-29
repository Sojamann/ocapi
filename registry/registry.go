package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

const maxParallelRequests = 5

// a map that stores the throtteling channel to use
// per host so that we don't create multiple registries
// and end up spamming the host again.
var throttleChanByHost sync.Map

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

type tagListResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type Manifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	Name          string `json:"name"`
	Tag           string `json:"tag"`
	Architecture  string `json:"architecture"`
	FsLayers      []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`
	History []struct {
		V1Compatibility string `json:"v1Compatibility"`
	} `json:"history"`
	// NOTE: signatures is ignored at the moment
}

type Registry struct {
	Host string
	auth authorizer
	// this cannel is used like n-locks. N is determined by
	// the buffer size and allows max n goroutines to
	// perform parallel requests. Others have to wait...
	throttleChan chan any
}

var ErrImageDoesNotExist = errors.New("image does not exist")
var ErrResourceDoesNotExist = errors.New("resource does not exist")
var ErrNotAllowedOrUnavailable = errors.New("your're either not allowed to access this resource or it does not exist")

func buildUrl(host, endpoint string) string {
	host = strings.TrimSuffix(host, "/")
	endpoint = strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("https://%s/%s", host, endpoint)
}

func NewRegisty(host string) (*Registry, error) {
	creds, found := credentialLookupTable[host]
	if !found {
		return nil, fmt.Errorf("no credentials for %s", host)
	}

	resp, err := http.Head(buildUrl(host, "v2/_catalog"))
	if err != nil {
		return nil, err
	}

	wwwAuth := resp.Header.Get("Www-authenticate")

	if resp.StatusCode != 401 || wwwAuth == "" {
		return nil, fmt.Errorf("expected auth challenge from registry, but got: %s", resp.Status)
	}

	// if another registry for the same host exists, use the same throttling channel
	throttleChan, _ := throttleChanByHost.LoadOrStore(host, make(chan any, maxParallelRequests))

	return &Registry{
		Host:         host,
		auth:         oauthAuthorizerFromChallenge(wwwAuth, creds),
		throttleChan: throttleChan.(chan any),
	}, nil
}

// makes the request and performs some common error checking
// and retry logic
// TODO: see error handling: https://github.com/opencontainers/distribution-spec/blob/main/spec.md
func (r *Registry) request(request *http.Request) (*http.Response, error) {
	request.Header.Set("Accept-Encoding", "*")

	// Send something into the channel. Either it blocks and we have to
	// wait until it is free or we can request right away and read from
	// it later to unblock. (chan = n locks)
	r.throttleChan <- nil
	log.Debug().Str("host", r.Host).Str("url", request.RequestURI)
	resp, err := http.DefaultClient.Do(request)
	<-r.throttleChan

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		resp.Body.Close()

		// retry one more time with fresh token
		r.auth.authorizeRequest(request)
		resp, err = http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			resp.Body.Close()
			// NOTE: in the Www-authenticate should be an error field that says so
			return nil, ErrNotAllowedOrUnavailable
		}
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
	catalogUrl := buildUrl(r.Host, "v2/_catalog")
	log.Debug().Str("host", r.Host).Msg("getting catalog")
	request, err := http.NewRequest("GET", catalogUrl, nil)
	if err != nil {
		return nil, err
	}

	if err = r.auth.authorizeRequest(request); err != nil {
		return nil, err
	}

	resp, err := r.request(request)
	if err != nil {
		if errors.Is(err, ErrNotAllowedOrUnavailable) {
			return nil, fmt.Errorf("you don't seem to have permission to request the image catalog of the registry %s", r.Host)
		}
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

	tagListUrl := buildUrl(r.Host, fmt.Sprintf("v2/%s/tags/list", imageName))
	log.Debug().Str("host", r.Host).Str("image", imageName).Msg("getting tags")
	request, err := http.NewRequest("GET", tagListUrl, nil)
	if err != nil {
		return nil, err
	}

	if err = r.auth.authorizeRepoPull(request, imageName); err != nil {
		return nil, err
	}

	resp, err := r.request(request)
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

	manifestUrl := buildUrl(r.Host, fmt.Sprintf("v2/%s/manifests/%s", imageName, tag))
	log.Debug().Str("host", r.Host).Str("image", imageName).Str("tag", tag).Msg("getting manifest")
	request, err := http.NewRequest("GET", manifestUrl, nil)
	if err != nil {
		return nil, err
	}

	if err = r.auth.authorizeRepoPull(request, imageName); err != nil {
		return nil, err
	}

	resp, err := r.request(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err = json.Unmarshal(content, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

func (r *Registry) Exists(imageName string, tag string) (bool, error) {
	imageName = strings.TrimSuffix(imageName, "/")
	imageName = strings.TrimPrefix(imageName, "/")

	manifestUrl := buildUrl(r.Host, fmt.Sprintf("v2/%s/manifests/%s", imageName, tag))
	request, err := http.NewRequest("HEAD", manifestUrl, nil)
	if err != nil {
		return false, err
	}

	if err = r.auth.authorizeRepoPull(request, imageName); err != nil {
		return false, err
	}

	resp, err := r.request(request)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}
