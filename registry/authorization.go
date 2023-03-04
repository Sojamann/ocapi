package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type tokenResponse struct {
	Token    string `json:"token"`
	Scope    string `json:"scope"`
	Validity int    `json:"expires_in"`
	Issued   string `json:"issued_at"`
}

type token struct {
	token      string
	scope      string
	validUntil time.Time
}

type authorizer interface {
	// authorizes this request by making it
	// and authenticating at the desired endpoint
	authorizeRequest(*http.Request) error
	// authorizes this request by optaining a registy
	// pull token. This uses the caching mechanism
	authorizeRepoPull(*http.Request, string) error
}

// TODO: make a on demand oauth authorizer
//			which asks the user for password
type oAuthAuthorizer struct {
	authEndpoint   string
	service        string
	credentials    credentials
	repoPullTokens map[string]*token
	mutex          sync.Mutex
}

func oauthAuthorizerFromChallenge(authenticate string, creds credentials) *oAuthAuthorizer {
	realm, service, _ := extractOAuthSettings(authenticate)
	return &oAuthAuthorizer{
		authEndpoint:   realm,
		service:        service,
		credentials:    creds,
		repoPullTokens: make(map[string]*token),
	}
}

func (o *oAuthAuthorizer) authorizeRequest(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// the request does not need authorization
	if resp.StatusCode == 200 {
		return nil
	}

	resp.Body.Close()

	wwwAuth := resp.Header.Get("Www-authenticate")

	if resp.StatusCode != 401 || wwwAuth == "" {
		return fmt.Errorf("expected authentication request but got response status: %s", resp.Status)
	}

	realm, service, scope := extractOAuthSettings(wwwAuth)
	token, err := optainToken(realm, service, scope, &o.credentials)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.token))

	return nil
}

func (o *oAuthAuthorizer) authorizeRepoPull(req *http.Request, repo string) error {
	repo = strings.TrimPrefix(repo, "/")
	repo = strings.TrimSuffix(repo, "/")

	o.mutex.Lock()
	token, found := o.repoPullTokens[repo]
	o.mutex.Unlock()

	if found {
		if time.Now().Before(token.validUntil) {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.token))
			return nil
		}
	}

	scope := "repository:" + repo + ":pull"
	token, err := optainToken(o.authEndpoint, o.service, scope, &o.credentials)
	if err != nil {
		return err
	}

	o.mutex.Lock()
	o.repoPullTokens[repo] = token
	o.mutex.Unlock()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.token))
	return nil
}

func optainToken(realm, service, scope string, creds *credentials) (*token, error) {
	// https://stackoverflow.com/questions/56193110/how-can-i-use-docker-registry-http-api-v2-to-obtain-a-list-of-all-repositories-i/68654659#68654659
	// https://docs.docker.com/registry/spec/auth/token/

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

	tokenGeneratedAt := time.Now().Add(-time.Second)
	if tResp.Issued != "" {
		tokenGeneratedAt, err = time.Parse(time.RFC3339, tResp.Issued)
		if err != nil {
			return nil, errors.New("got bad token 'issued at' timestamp")
		}
	}
	validFor := time.Second * time.Duration(tResp.Validity)
	validUntil := tokenGeneratedAt.Add(validFor).Add(-time.Second)
	return &token{
		token:      tResp.Token,
		scope:      tResp.Scope,
		validUntil: validUntil,
	}, nil
}

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
