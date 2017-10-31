package http

import (
	"github.com/gorilla/handlers"
	"net/http"
	"github.com/davepgreene/propsd-agent/sources"
	"encoding/json"
)

type metadataHandler struct{
	metadata *sources.Metadata
}

func newMetadataHandler(m *sources.Metadata) http.Handler {
	return handlers.MethodHandler{
		"GET": &metadataHandler{m},
	}
}

func (h *metadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	b, _ := json.Marshal(h.metadata.Properties())

	w.Write(b)
}
