package netu

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// CommonClient Common client.
var CommonClient = &http.Client{Timeout: 10 * time.Second}

// MakeUrl Make URL from base url and params.
func MakeUrl(baseURL string, params map[string]string) (string, error) {
	myurl, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	parameters := url.Values{}
	for k, v := range params {
		parameters.Add(k, v)
	}
	myurl.RawQuery = parameters.Encode()
	return myurl.String(), nil
}

// MakeRequest Make request from base url, params and headers.
func MakeRequest(baseURL string, params map[string]string, headers map[string]string) ([]byte, error) {

	myurl, err0 := url.Parse(baseURL)
	if err0 != nil {
		return nil, err0
	}

	parameters := url.Values{}
	for k, v := range params {
		parameters.Add(k, v)
	}
	myurl.RawQuery = parameters.Encode()

	baseURL = myurl.String()

	req, err1 := http.NewRequest("GET", baseURL, nil)
	if err1 != nil {
		return nil, err1
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err2 := CommonClient.Do(req)
	if err2 != nil {
		return nil, err2
	}

	defer resp.Body.Close()
	responseData, err3 := ioutil.ReadAll(resp.Body)
	if err3 != nil {
		return nil, err3
	}

	return responseData, nil
}
