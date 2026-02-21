package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"game-server/internal/model"
	"game-server/internal/service"
)

type HTTPHandler struct {
	userService *service.UserService
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{
		userService: service.NewUserService(),
	}
}

func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.Username == "" || req.Password == "" {
		h.writeResponse(w, http.StatusBadRequest, "username and password are required", nil)
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		h.writeResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeResponse(w, http.StatusOK, "register success", user)
}

func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	user, token, err := h.userService.Login(&req)
	if err != nil {
		h.writeResponse(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	h.writeResponse(w, http.StatusOK, "login success", &model.UserLoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *HTTPHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeResponse(w, http.StatusBadRequest, "invalid user id", nil)
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		h.writeResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeResponse(w, http.StatusOK, "success", user)
}

func (h *HTTPHandler) writeResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&model.UserResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
