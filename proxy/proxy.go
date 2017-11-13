package proxy

import (
	"net/url"
	"net/http"
	"time"
	"io/ioutil"
	"strings"
	log "github.com/sirupsen/logrus"
)

const (
	UpstreamHeader = "X-Upstream-Proxy-Invalid"
)

type Proxy struct {
	url string
	handler http.Handler
	client http.Client
	data string
}

func New(url string, handler http.Handler) *Proxy {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	return &Proxy{
		url: url,
		client: client,
		handler: handler,
		data: "",
	}
}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// We can swallow any errors in request creation because the only ones we could generate
	// are invalid URLs. If that's the case, we should let error handling for the http client
	// making the request take care of any issues. That way the client (the one connecting to
	// this agent) completes the request.
	req, _ := http.NewRequest("GET", p.url, r.Body)

	var bodyStr string

	resp, err := p.client.Do(req)

	if err != nil {
		// We need to know what kind of error we're running into. If it's a url.Error,
		// this isn't an issue upstream, it's a config issue.
		switch t := err.(type) {
		case *url.Error:
			log.WithFields(log.Fields{
				"error": t,
			}).Error("Upstream URL cannot be parsed.")
		default:
			log.WithFields(log.Fields{
				"error": t,
			}).Warn("Error connecting to proxied target. Falling back to cached data.")
		}

		// Because the stats middleware injects its own ResponseWriter implementation, we
		// can't just wrap http.ResponseWriter in our own implementation where we track the
		// state of the upstream server. Instead we have to write a header to the response
		// as a flag.
		rw.Header().Add(UpstreamHeader, "true")
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