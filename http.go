package akashicpay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const (
	// TODO: Get version from somewhere
	version = "1.0.0"
	client  = "go-sdk"
)

var DefaultClient = &http.Client{}

// TODO: Handle bad response statuses (400s etc.)
func Get[T any](url string) (T, error) {
	var result T

	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return result, err
	}

	setHeaders(request)

	response, err := DefaultClient.Do(request)

	if err != nil {
		return result, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Send a POST request. data should be a struct with json tags
func Post[T any](url string, data any) (T, error) {
	var result T

	jsonData, err := json.Marshal(data)

	if err != nil {
		return result, err
	}

	request, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	request.ContentLength = int64(len(jsonData))

	if err != nil {
		return result, err
	}

	setHeaders(request)

	response, err := DefaultClient.Do(request)

	if err != nil {
		return result, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	if err != nil {
		return result, err
	}

	// Unmarshal freaks out if body is empty
	if len(body) > 0 {
		err = json.Unmarshal(body, &result)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func setHeaders(request *http.Request) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Ap-Version", version)
	request.Header.Set("Ap-Client", client)
}
