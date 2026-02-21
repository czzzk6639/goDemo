package router

import (
	"net/http"

	"game-server/internal/handler"
)

type Router struct {
	handler *handler.HTTPHandler
}

func NewRouter() *Router {
	return &Router{
		handler: handler.NewHTTPHandler(),
	}
}

func (r *Router) Setup() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", r.handler.Register)
	mux.HandleFunc("POST /api/login", r.handler.Login)
	mux.HandleFunc("GET /api/user/{id}", r.handler.GetUser)

	return mux
}
