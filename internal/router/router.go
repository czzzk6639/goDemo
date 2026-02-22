package router

import (
	"log"
	"net/http"

	"game-server/internal/handler"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Router struct {
	handler   *handler.HTTPHandler
	wsHandler *handler.WSHandler
}

func NewRouter() *Router {
	return &Router{
		handler:   handler.NewHTTPHandler(),
		wsHandler: handler.NewWSHandler(),
	}
}

func (r *Router) Setup() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/register", r.handler.Register)
	mux.HandleFunc("POST /api/login", r.handler.Login)
	mux.HandleFunc("GET /api/user/{id}", r.handler.GetUser)

	mux.HandleFunc("/ws", r.handleWebSocket)

	mux.Handle("/", http.FileServer(http.Dir("web")))

	return mux
}

func (r *Router) handleWebSocket(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	go r.wsHandler.HandleWS(conn)
}
