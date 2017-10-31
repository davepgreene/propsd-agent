package http

import (
	"github.com/gorilla/handlers"
	"net/http"
	"io/ioutil"
)

type propertiesHandler struct{}

func newPropertiesHandler() http.Handler {
	return handlers.MethodHandler{
		"GET": &propertiesHandler{},
	}
}

func (h *propertiesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	rw.Write(body)
}
