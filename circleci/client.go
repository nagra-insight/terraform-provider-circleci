package circleci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

const (
	defaultBaseURL = "https://circleci.com/api/v1.1"
	envvarEndpoint = "envvar"
)

var (
	envVarNameRegexp = regexp.MustCompile("^[[:alpha:]]+[[:word:]]*$")
)

// EnvironmentVariable inside a CircleCI project
type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ClientOpt
type ClientOpt func(*Client) error

// WithBaseURL sets the base URL of the client
func WithBaseURL(baseURL string) ClientOpt {
	return func(c *Client) error {
		c.baseURL = baseURL
		return nil
	}
}

// ValidateEnvironmentVariableName validates the name of the variable is allowed
// by CircleCI
func ValidateEnvironmentVariableName(name string) bool {
	return envVarNameRegexp.MatchString(name)
}

// Client for the CircleCI API
type Client struct {
	baseURL      string
	token        string
	vcsType      string
	organization string
	httpClient   *http.Client
}

// NewClient creates a new CircleCI API client
func NewClient(token, vcsType, organization string, opts ...ClientOpt) (*Client, error) {
	client := &Client{
		baseURL:      defaultBaseURL,
		token:        token,
		vcsType:      vcsType,
		organization: organization,
		httpClient:   http.DefaultClient,
	}

	// Applies all the optional options
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func (c *Client) buildApiURL(projectName, endpoint string) string {
	return fmt.Sprintf("%s/project/%s/%s/%s/%s", c.baseURL, c.vcsType, c.organization, projectName, endpoint)
}

// AddEnvironmentVariable creates a new environment variable.
// https://circleci.com/docs/api/#add-environment-variables
func (c *Client) AddEnvironmentVariable(projectName, envName, envValue string) error {
	endpointURL := c.buildApiURL(projectName, envvarEndpoint)

	e := EnvironmentVariable{
		Name:  envName,
		Value: envValue,
	}

	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(e); err != nil {
		// TODO(matteo): proper error handling
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpointURL, b)
	if err != nil {
		// TODO(matteo): proper error handling
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(c.token, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// TODO(matteo): proper error handling
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("client: create wrong status code %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) EnvironmentVariableExists(projectName, envName string) (bool, error) {
	endpointURL := fmt.Sprintf("%s/%s", c.buildApiURL(projectName, envvarEndpoint), envName)

	req, err := http.NewRequest(http.MethodHead, endpointURL, nil)
	if err != nil {
		// TODO(matteo): proper error handling
		return false, err
	}

	req.SetBasicAuth(c.token, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// TODO(matteo): proper error handling
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}

		return false, fmt.Errorf("circleci: wrong status code %d getting environment variable", resp.StatusCode)
	}

	return true, nil
}

// GetEnvironmentVariable returns the value of an environment variable of a
// project given its name.
// https://circleci.com/docs/api/#get-single-environment-variable
func (c *Client) GetEnvironmentVariable(projectName, envName string) (*EnvironmentVariable, error) {
	endpointURL := fmt.Sprintf("%s/%s", c.buildApiURL(projectName, envvarEndpoint), envName)

	req, err := http.NewRequest(http.MethodGet, endpointURL, nil)
	if err != nil {
		// TODO(matteo): proper error handling
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.token, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// TODO(matteo): proper error handling
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("circleci: wrong status code %d getting environment variable", resp.StatusCode)
	}

	e := new(EnvironmentVariable)
	if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
		// TODO(matteo): proper error handling
		return nil, err
	}

	return e, nil
}

// DeleteEnvironmentVariable deletes an environment variable from a project
// given its name.
// https://circleci.com/docs/api/#delete-environment-variables
func (c *Client) DeleteEnvironmentVariable(projectName, envName string) error {
	endpointURL := fmt.Sprintf("%s/%s", c.buildApiURL(projectName, envvarEndpoint), envName)

	req, err := http.NewRequest(http.MethodDelete, endpointURL, nil)
	if err != nil {
		// TODO(matteo): proper error handling
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.token, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// TODO(matteo): proper error handling
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("circleci: wrong status code %d deleting environment variable", resp.StatusCode)
	}

	return nil
}
