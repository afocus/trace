package gin

import (
	"errors"
	"net/http"

	"github.com/afocus/trace"

	"github.com/gin-gonic/gin"
)

func Middlewave() func(*gin.Context) {
	return func(c *gin.Context) {

		var (
			path     = c.Request.URL.Path
			raw      = c.Request.URL.RawQuery
			savedCtx = c.Request.Context()
		)

		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()

		if raw != "" {
			path = path + "?" + raw
		}

		e, ctx := trace.Start(
			trace.ExtractHttpHeader(savedCtx, c.Request.Header),
			c.FullPath(),
		)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		var (
			status = c.Writer.Status()
			size   = c.Writer.Size()
		)

		e.SetAttributes(
			trace.Attribute("http.url", path),
			trace.Attribute("http.method", c.Request.Method),
			trace.Attribute("http.user_agent", c.Request.UserAgent()),
			trace.Attribute("http.status_code", status),
			trace.Attribute("http.response_content_length", size),
			trace.Attribute("http.clientip", c.ClientIP()),
		)

		if len(c.Errors) > 0 {
			e.SetAttributes(
				trace.Attribute("gin.errors", c.Errors.String()),
			)
		}

		switch status / 100 {
		case 1, 2, 3:
			e.End()
		default:
			e.End(errors.New(http.StatusText(status)))
		}

	}
}
