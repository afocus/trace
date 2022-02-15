package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var DefaultClient = &http.Client{Transport: NewTransport(http.DefaultTransport)}

func Get(c context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return DefaultClient.Do(req)
}

func Post(c context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c, "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return DefaultClient.Do(req)
}

func PostForm(c context.Context, url string, data url.Values) (*http.Response, error) {
	return Post(c, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func Head(c context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c, "HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return DefaultClient.Do(req)
}

func PostJSON(c context.Context, url string, send, result interface{}) (int, error) {
	data, err := json.Marshal(send)
	if err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	buf.Write(data)
	response, err := Post(c, url, "application/json", &buf)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(result)
	return response.StatusCode, err
}

func GetJSON(c context.Context, url string, result interface{}) (int, error) {
	response, err := Get(c, url)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(result)
	return response.StatusCode, err
}
