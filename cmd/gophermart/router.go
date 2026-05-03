package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/liebeSonne/gophermart/api/server"
	"github.com/liebeSonne/gophermart/internal/handler"
)

func initRouter() (http.Handler, error) {
	s := handler.NewServer()
	r := chi.NewMux()
	h := server.HandlerFromMux(s, r)
	return h, nil
}
