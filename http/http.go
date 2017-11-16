package http

import (
	"time"
	"net/http"
	"github.com/thoas/stats"
	"github.com/davepgreene/propsd-agent/sources"
	prox "github.com/davepgreene/propsd-agent/proxy"
	"io/ioutil"
	"github.com/gorilla/handlers"
)

type Status struct {
	Version string `json:"version,omitempty"`
	Uptime string `json:"uptime,omitempty"`
	Code int `json:"code,omitempty"`
	Metadata bool `json:"metadata"`
	Proxy bool `json:"proxy"`
	Body bool `json:"body"`
}

type statusHandler struct {
	metadata *sources.Metadata
	stats *stats.Stats
	fn func(*statusHandler, http.ResponseWriter, *http.Request)
}

func newStatusHandler(metadata *sources.Metadata, s *stats.Stats, fn func(h *statusHandler, w http.ResponseWriter, r *http.Request)) http.Handler {
	return handlers.MethodHandler{
		"GET": &statusHandler{metadata, s, fn},
	}
}

func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.fn(h, w, r)
}

func (h *statusHandler) GenerateStatus(w http.ResponseWriter, r *http.Request) (Status, int) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)

	// See comment in proxy/proxy.go
	upstream := w.Header().Get(prox.UpstreamHeader)
	w.Header().Del(prox.UpstreamHeader)

	status := Status{
		Version: "0.0.0",
		Uptime: h.stats.Uptime.Format(time.RFC3339),
		Metadata: h.metadata.Ok(),
		Proxy: upstream == "true",
		Body: len(body) != 0,
	}

	if !status.Metadata || !status.Proxy || !status.Body {
		return status, http.StatusInternalServerError
	}

	return status, http.StatusOK
}