package http

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	"net/http"
	"context"
	log "github.com/sirupsen/logrus"
)

func shutdown(s *http.Server) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	var timeout time.Duration

	if s.ReadHeaderTimeout != 0 {
		timeout = s.ReadHeaderTimeout
	} else if s.IdleTimeout != 0 {
		timeout = s.IdleTimeout
	} else {
		timeout = s.ReadTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Infof("Shutdown with timeout: %s", timeout)

	if err := s.Shutdown(ctx); err != nil {
		log.Error(err)
	} else {
		log.Info("Server stopped")
	}
}
