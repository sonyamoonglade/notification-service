package server

import (
	"github.com/julienschmidt/httprouter"
	"github.com/sonyamoonglade/notification-service/config"
	"net/http"
)

func NewServer(cfg *config.AppConfig) (*http.Server, *httprouter.Router) {
	h := httprouter.New()
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      h,
		ReadTimeout:  15,
		WriteTimeout: 15,
		IdleTimeout:  60,
	}
	return srv, h
}
