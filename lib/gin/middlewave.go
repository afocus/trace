package gin

import (
	"errors"
	"net/http"

	"github.com/afocus/trace"

	"github.com/gin-gonic/gin"
)

func Middlewave() func(*gin.Context) {
	return func(c *gin.Context) {

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		if raw != "" {
			path = path + "?" + raw
		}

		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()

		e := trace.Start(
			trace.ExtractHttpHeader(savedCtx, c.Request.Header),
			c.FullPath(),
			trace.Attribute("http.url", path),
			trace.Attribute("http.method", c.Request.Method),
			trace.Attribute("http.user_agent", c.Request.UserAgent()),
		)

		c.Request = c.Request.WithContext(e.Context())

		c.Next()

		status := c.Writer.Status()
		size := c.Writer.Size()

		if len(c.Errors) > 0 {
			e.Log().Error().Str("error", c.Errors.String()).Msg("gin.errors")
		}

		switch status / 100 {
		case 1, 2, 3:
			e.End(
				trace.Attribute("http.status_code", status),
				trace.Attribute("http.response_content_length", size),
			)
		default:
			e.EndError(
				errors.New(http.StatusText(status)),
				trace.Attribute("http.status_code", status),
				trace.Attribute("http.response_content_length", size),
			)
		}

	}
}
