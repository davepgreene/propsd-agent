package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/davepgreene/propsd-agent/utils"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/thoas/stats"
	"github.com/urfave/negroni"
	"github.com/davepgreene/propsd-agent/sources"
	"github.com/justinas/alice"
)

// Handler returns an http.Handler for the API.
func Handler(metadata *sources.Metadata) {
	r := mux.NewRouter()
	statsMiddleware := stats.New()
	r.HandleFunc("/stats", newAdminHandler(statsMiddleware).ServeHTTP)

	v1 := r.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/metadata", newMetadataHandler(metadata).ServeHTTP)

	prox := proxy(viper.GetString("propsd.upstream"))
	chain := alice.New(metadataMiddleware(metadata)).Append(prox)

	// Conqueso handler
	v1.Handle("/conqueso", chain.ThenFunc(newConquesoHandler().ServeHTTP))

	// Properties handlers
	v1.Handle("/properties", chain.ThenFunc(newPropertiesHandler().ServeHTTP))


	// Define our 404 handler
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	// Add middleware handlers
	n := negroni.New()
	n.Use(negroni.NewRecovery())

	if viper.GetBool("log.requests") {
		n.Use(negronilogrus.NewCustomMiddleware(utils.GetLogLevel(), utils.GetLogFormatter(), "requests"))
	}

	n.Use(statsMiddleware)
	n.UseHandler(r)

	// Set up connection
	conn := fmt.Sprintf("%s:%d", viper.GetString("service.host"), viper.GetInt("service.port"))
	log.Info(fmt.Sprintf("Listening on %s", conn))

	// Bombs away!
	server := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Addr:              conn,
		Handler:           n,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	shutdown(server)
}

// notFoundHandler provides a standard response for unhandled paths
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	w.Write([]byte(""))
}
