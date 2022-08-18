package server

import (
	"github.com/julienschmidt/httprouter"
	"github.com/sonyamoonglade/notification-service/config"
	"net/http"
	"time"
)

func NewServer(cfg *config.AppConfig) (*http.Server, *httprouter.Router) {
	h := httprouter.New()
	srv := &http.Server{
		Addr:           ":" + cfg.AppPort,
		Handler:        h,
		ReadTimeout:    time.Second * 15,
		WriteTimeout:   time.Second * 15,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 2 << 10,
		ConnContext:    nil,
	}
	return srv, h
}
