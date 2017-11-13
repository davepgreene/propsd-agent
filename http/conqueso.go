package http

import (
	"github.com/gorilla/handlers"
	"net/http"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"github.com/Jeffail/gabs"
	"github.com/doublerebel/bellows"
	"strconv"
	"github.com/magiconair/properties"
	"strings"
)

type conquesoHandler struct{}

func newConquesoHandler() http.Handler {
	return handlers.MethodHandler{
		"GET": &conquesoHandler{},
	}
}

func (h *conquesoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)

	rw.Header().Set("Content-Type", "text/plain")
	var props []byte
	var status int

	if len(body) > 0 {
		props, status = TransformProperties(body)
	} else {
		props = []byte("")
		status = http.StatusGone
	}

	rw.WriteHeader(status)
	rw.Write(props)
}

func TransformProperties(p []byte) ([]byte, int) {
	log.Info("Transforming properties")

	jsonParsed, err := gabs.ParseJSON(p)
	if err != nil {
		log.Error(err)
		return []byte(""), http.StatusInternalServerError
	}

	// Delete instance and tags keys from the parsed object
	if err := jsonParsed.Delete("instance"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Deleting instance")
	}

	if err := jsonParsed.Delete("tags"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Deleting tags")
	}

	// Because bellows flattens to a map[string]interface{} we have
	// to range over each key/value pair in the flatmap, use a switch
	// statement over the type of each value and coerce it to a string
	// so we can pass a map[string]string to properties.LoadMap.
	flattened := bellows.Flatten(jsonParsed.Data())
	normalized := make(map[string]string)
	for k, v := range flattened {
		switch i := v.(type) {
		case int:
			normalized[string(k)] = strconv.Itoa(i)
		case string:
			normalized[string(k)] = strings.Replace(i, "\n", "\\n", -1)
		case []string:
			normalized[string(k)] = strings.Join(i, ",")
		default:
			normalized[string(k)] = i.(string)
		}
	}

	props := properties.LoadMap(normalized).String()
	props = strings.Replace(props, " = ", "=", -1)

	return []byte(props), http.StatusOK
}