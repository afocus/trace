package gin

import (
	"errors"
	"net/http"
	"tempotest/traceing"

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

		e := traceing.Start(
			traceing.ExtractHttpHeader(savedCtx, c.Request.Header),
			c.FullPath(),
			traceing.Attribute("http.url", path),
			traceing.Attribute("http.method", c.Request.Method),
			traceing.Attribute("http.user_agent", c.Request.UserAgent()),
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
				traceing.Attribute("http.status_code", status),
				traceing.Attribute("http.response_content_length", size),
			)
		default:
			e.EndError(
				errors.New(http.StatusText(status)),
				traceing.Attribute("http.status_code", status),
				traceing.Attribute("http.response_content_length", size),
			)
		}

	}
}
