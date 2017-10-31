package http

import (
	"net/http"
	prox "github.com/davepgreene/propsd-agent/proxy"
	"github.com/davepgreene/propsd-agent/sources"
	"github.com/justinas/alice"
	"github.com/spf13/viper"
	"encoding/json"
	"io/ioutil"
	"bytes"
	"strconv"
)

type metadataProxyHandler struct {
	handler http.Handler
	metadata *sources.Metadata
}

func (p *metadataProxyHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	properties := make(map[string]interface{})
	properties["instance"] = p.metadata.Properties()
	properties["image"] = viper.GetStringMap("properties.image")
	propertiesJSON, _ := json.Marshal(properties)

	r.Body = ioutil.NopCloser(bytes.NewReader(propertiesJSON))
	r.ContentLength = int64(len(propertiesJSON))
	r.Header.Set("Content-Length", strconv.Itoa(len(propertiesJSON)))
	p.handler.ServeHTTP(rw, r)
}

func metadataMiddleware(m *sources.Metadata) alice.Constructor {
	return func(handler http.Handler) http.Handler {
		return &metadataProxyHandler{handler, m}
	}
}

func proxy(url string) alice.Constructor {
	return func(handler http.Handler) http.Handler {
		return prox.New(url, handler)
	}
}