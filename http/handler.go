package http

import (
	"net/http"
	"tempotest/traceing"
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

	e := traceing.Start(
		req.Context(),
		req.URL.Path,
		traceing.Attribute("http.method", req.Method),
		traceing.Attribute("http.url", path),
		traceing.Attribute("http.request_content_length", req.ContentLength),
	)

	w := &responseWriter{w: rw}

	h.hander.ServeHTTP(w, req.WithContext(e.Context()))
	e.End(
		traceing.Attribute("http.status_code", w.statusCode),
		traceing.Attribute("http.response_content_length", w.size),
	)
}
