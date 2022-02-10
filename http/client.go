package http

import (
	"context"
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
