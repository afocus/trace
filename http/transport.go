package http

import (
	"errors"
	"fmt"
	"net/http"
	"tempotest/traceing"
)

type Transport struct {
	base http.RoundTripper
}

func NewTransport(baseTransport http.RoundTripper) *Transport {
	return &Transport{base: baseTransport}
}

func (tp *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	name := fmt.Sprintf("http %s", req.URL.Host)

	e := traceing.Start(
		req.Context(),
		name,
		traceing.Attribute("http.method", req.Method),
		traceing.Attribute("http.url", req.URL.String()),
		traceing.Attribute("http.request_content_length", req.ContentLength),
	)
	// 将信息注入到header里 使其可以传递到服务端，
	// 服务端根据header展开拿到taraceid等信息
	traceing.InjectHttpHeader(e.Context(), req.Header)

	resp, err := tp.base.RoundTrip(req)
	if err != nil {
		e.EndError(err)
	} else {
		switch resp.StatusCode / 100 {
		case 1, 2, 3:
			e.EndOK(
				traceing.Attribute("http.status_code", resp.StatusCode),
				traceing.Attribute("http.response_content_length", resp.ContentLength),
			)
		default:
			e.EndError(
				errors.New(http.StatusText(resp.StatusCode)),
				traceing.Attribute("http.status_code", resp.StatusCode),
				traceing.Attribute("http.response_content_length", resp.ContentLength),
			)
		}
	}
	return resp, err
}
