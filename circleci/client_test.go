package circleci

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testingRoundTripper func(r *http.Request) (*http.Response, error)

func (t testingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return t(r)
}

func TestClient_buildApiURL(t *testing.T) {
	testCases := []struct {
		vcsType      string
		organization string
		projectName  string
		endpoint     string
		expected     string
	}{
		{
			vcsType:      "github",
			organization: "circleci",
			projectName:  "project1",
			endpoint:     "test1",
			expected:     "https://circleci.com/api/v1.1/project/github/circleci/project1/test1",
		},
		{
			vcsType:      "bitbucket",
			organization: "circleci",
			projectName:  "project2",
			endpoint:     "test2",
			expected:     "https://circleci.com/api/v1.1/project/bitbucket/circleci/project2/test2",
		},
	}

	for _, tt := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tt.projectName, tt.endpoint), func(t *testing.T) {
			client, err := NewClient("something", tt.vcsType, tt.organization)
			assert.NoError(t, err)
			actual := client.buildApiURL(tt.projectName, tt.endpoint)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func TestClient_AddEnvironmentVariableOK(t *testing.T) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())

	httpClient := &http.Client{
		Transport: testingRoundTripper(func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "application/json; charset=utf-8", r.Header.Get("Content-Type"))

			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, token, username)
			assert.Equal(t, "", password)

			bodyContent, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "{\"name\":\"key\",\"value\":\"value\"}\n", string(bodyContent))

			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			}, nil
		}),
	}

	client := Client{
		token:        token,
		vcsType:      "github",
		organization: "foo",
		httpClient:   httpClient,
	}

	err := client.AddEnvironmentVariable("bar", "key", "value")
	assert.NoError(t, err)
}

func TestClient_GetEnvironmentVariableOK(t *testing.T) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())

	httpClient := &http.Client{
		Transport: testingRoundTripper(func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, token, username)
			assert.Equal(t, "", password)

			assert.Nil(t, r.Body)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"name\":\"key\",\"value\":\"value\"}\n"))),
			}, nil
		}),
	}

	client := Client{
		token:        token,
		vcsType:      "github",
		organization: "foo",
		httpClient:   httpClient,
	}

	envVar, err := client.GetEnvironmentVariable("bar", "key")
	assert.NoError(t, err)
	assert.Equal(t, &EnvironmentVariable{
		Name:  "key",
		Value: "value",
	}, envVar)
}

func TestClient_GetEnvironmentVariableWrongStatus(t *testing.T) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())

	httpClient := &http.Client{
		Transport: testingRoundTripper(func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, token, username)
			assert.Equal(t, "", password)

			assert.Nil(t, r.Body)

			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(nil),
			}, nil
		}),
	}

	client := Client{
		token:        token,
		vcsType:      "github",
		organization: "foo",
		httpClient:   httpClient,
	}

	envVar, err := client.GetEnvironmentVariable("bar", "key")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "circleci: wrong status code 400 getting environment variable")
	assert.Nil(t, envVar)
}

func TestClient_DeleteEnvironmentVariableOK(t *testing.T) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())

	httpClient := &http.Client{
		Transport: testingRoundTripper(func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, token, username)
			assert.Equal(t, "", password)

			assert.Nil(t, r.Body)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"message\":\"ok\"}\n"))),
			}, nil
		}),
	}

	client := Client{
		token:        token,
		vcsType:      "github",
		organization: "foo",
		httpClient:   httpClient,
	}

	err := client.DeleteEnvironmentVariable("bar", "key")
	assert.NoError(t, err)
}

func TestClient_DeleteEnvironmentVariableWrongStatus(t *testing.T) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())

	httpClient := &http.Client{
		Transport: testingRoundTripper(func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, token, username)
			assert.Equal(t, "", password)

			assert.Nil(t, r.Body)

			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(nil),
			}, nil
		}),
	}

	client := Client{
		token:        token,
		vcsType:      "github",
		organization: "foo",
		httpClient:   httpClient,
	}

	err := client.DeleteEnvironmentVariable("bar", "key")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "circleci: wrong status code 404 deleting environment variable")
}
