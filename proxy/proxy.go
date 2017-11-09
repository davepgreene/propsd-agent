package proxy

import (
	"net/url"
	"net/http"
	"github.com/vulcand/oxy/testutils"
	"time"
	"io/ioutil"
	"strings"
	log "github.com/sirupsen/logrus"
)

type Proxy struct {
	url *url.URL
	handler http.Handler
	client http.Client
	data string
}

func New(url string, handler http.Handler) *Proxy {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	return &Proxy{
		url: testutils.ParseURI(url),
		client: client,
		handler: handler,
		data: "",
	}
}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	req, err := http.NewRequest("GET", p.url.String(), r.Body)
	if err != nil {
		// generating the request failed
	}

	var bodyStr string

	resp, err := p.client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Error connecting to proxied target. Falling back to cached data.")
		bodyStr = p.data
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr = string(body)
		p.data = bodyStr
	}

	r.Body = ioutil.NopCloser(strings.NewReader(bodyStr))
	r.ContentLength = int64(len(bodyStr))
	p.handler.ServeHTTP(rw, r)
}