package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/afocus/trace"
)

type Transport struct {
	base http.RoundTripper
}

func NewTransport(baseTransport http.RoundTripper) *Transport {
	return &Transport{base: baseTransport}
}

func (tp *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	name := fmt.Sprintf("http %s", req.URL.Host)

	e, ctx := trace.Start(
		req.Context(),
		name,
		trace.Attribute("http.method", req.Method),
		trace.Attribute("http.url", req.URL.String()),
		trace.Attribute("http.request_content_length", req.ContentLength),
	)
	// 将信息注入到header里 使其可以传递到服务端，
	// 服务端根据header展开拿到taraceid等信息
	trace.InjectHttpHeader(ctx, req.Header)

	resp, err := tp.base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		e.End(err)
	} else {
		e.SetAttributes(
			trace.Attribute("http.status_code", resp.StatusCode),
			trace.Attribute("http.response_content_length", resp.ContentLength),
		)
		switch resp.StatusCode / 100 {
		case 1, 2, 3:
			e.End()
		default:
			e.End(errors.New(http.StatusText(resp.StatusCode)))
		}
	}
	return resp, err
}
