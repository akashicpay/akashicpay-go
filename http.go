package akashicpay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
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

	err = checkResponseForErrors(response)

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

// Checks for HTTP errors, handles and formats them
func checkResponseForErrors(response *http.Response) error {
	if response.StatusCode < 400 {
		return nil
	}
	// Check if response has a JSON body we can parse for an error message
	isJson := slices.ContainsFunc(response.Header.Values("Content-Type"), func(e string) bool { m, err := regexp.MatchString("application/json", e); return m && err == nil })

	if isJson {
		var httpError map[string]any
		body, err := io.ReadAll(response.Body)

		if err != nil {
			return err
		}

		// Unmarshal freaks out if body is empty
		if len(body) > 0 {
			err = json.Unmarshal(body, &httpError)
			if err == nil {
				return fmt.Errorf("%v: %v", httpError["error"], httpError["message"])
			}
		}
	}

	// If it's not JSON or we couldn't find any JSON body, just return status code
	return fmt.Errorf("HTTP Error: %v", response.StatusCode)

}
