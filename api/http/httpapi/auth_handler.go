package httpapi

import (
	"crypthttp/usecases"
	"net/http"
)

type AuthHandler struct {
	Auth usecases.Auth
}

func NewAuthHandler(auth usecases.Auth) *AuthHandler {
	return &AuthHandler{auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds LoginRequest
	err := decodeJSON(r, w, &creds)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	token, err := h.Auth.Register(r.Context(), creds.Username, creds.Password)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	err = WriteJSON(w, 201, TokenResponse{Token: token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds LoginRequest
	err := decodeJSON(r, w, &creds)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	token, err := h.Auth.Login(r.Context(), creds.Username, creds.Password)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	err = WriteJSON(w, 200, TokenResponse{Token: token})
}
