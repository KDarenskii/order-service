package mzerolog

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/KDarenskii/order-service/internal/pkg/http/httph"
)

type middleware struct {
	log zerolog.Logger

	fromOptions struct {
		skipper func(r *http.Request) bool
	}
}

func (m *middleware) Callback(c *gin.Context) {
	const (
		tailSuccess    = " finished with no error"
		tailClientFail = " finished with client error"
		tailServerFail = " finished (or aborted) with server error"
	)

	startTime := time.Now()

	c.Next()

	err := httph.ErrorGet(c.Request)

	execTime := time.Since(startTime)

	if m.fromOptions.skipper(c.Request) {
		return
	}

	status := c.Writer.Status()

	var mb strings.Builder
	mb.Grow(48 + len(c.Request.RequestURI))
	mb.WriteString(c.Request.Method)
	mb.WriteByte(' ')
	mb.WriteString(c.Request.RequestURI)

	var ev *zerolog.Event
	switch {
	case status >= http.StatusInternalServerError:
		mb.WriteString(tailServerFail)
		ev = m.log.Error()
	case status >= http.StatusBadRequest:
		mb.WriteString(tailClientFail)
		ev = m.log.Warn()
	default:
		mb.WriteString(tailSuccess)
		ev = m.log.Debug()
	}

	ev.Err(err)
	ev.Ctx(c.Request.Context())
	ev.Int("status", status)
	ev.Str("exec_time", execTime.String())
	ev.Str("client_ip", c.ClientIP())
	ev.Msg(mb.String())
}

func NewMiddleware(opts ...Option) func(*gin.Context) {
	m := middleware{
		log: log.Logger,
	}
	m.fromOptions.skipper = defaultSkipper

	for _, opt := range opts {
		opt(&m)
	}

	return m.Callback
}

func defaultSkipper(_ *http.Request) bool {
	return false
}
