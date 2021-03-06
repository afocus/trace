package http

import (
	"net/http"
	"strings"

	"github.com/afocus/trace"
)

type Handler struct {
	hander http.Handler
}

type responseWriter struct {
	statusCode int
	size       int
	w          http.ResponseWriter
}

func (w *responseWriter) Write(data []byte) (int, error) {
	n, err := w.w.Write(data)
	if n > 0 {
		w.size += n
	}
	return n, err
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.w.WriteHeader(statusCode)
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	path := req.URL.Path
	raw := req.URL.RawQuery

	if raw != "" {
		path = path + "?" + raw
	}

	e, ctx := trace.Start(trace.ExtractHttpHeader(req.Context(), req.Header), req.URL.Path)

	w := &responseWriter{w: rw}
	h.hander.ServeHTTP(w, req.WithContext(ctx))

	clientip := req.Header.Get("X-Forwarded-For")
	if clientip == "" {
		clientip = strings.Split(req.RemoteAddr, ":")[0]
	}

	e.SetAttributes(
		trace.Attribute("http.method", req.Method),
		trace.Attribute("http.url", path),
		trace.Attribute("http.request_content_length", req.ContentLength),
		trace.Attribute("http.status_code", w.statusCode),
		trace.Attribute("http.response_content_length", w.size),
		trace.Attribute("http.clientip", clientip),
	)
	e.End()
}
