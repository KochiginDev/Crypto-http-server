package httpapi

import (
	"context"
	"crypthttp/domain"
	"errors"
	"log/slog"
	"net/http"
)

type Error struct {
	Error string `json:"error"`
}

func ErrorWriter(resp http.ResponseWriter, req *http.Request, err error) {
	var status int
	var msg string
	switch {
	case errors.Is(err, context.Canceled):
		return
	case errors.Is(err, domain.ErrInvalidInput):
		status = 400
		msg = "Invalid Input"
	case errors.Is(err, domain.ErrUnknownSymbol):
		status = 400
		msg = "Unknown Symbol"
	case errors.Is(err, domain.ErrNotFound):
		status = 404
		msg = "Not Found"
	case errors.Is(err, domain.ErrAlreadyExists):
		status = 409
		msg = "Already Exists"
	case errors.Is(err, domain.ErrInvalidCredentials):
		status = 401
		msg = "Invalid Credentials"
	default:
		slog.Error("Internal error", "method", req.Method, "path", req.URL.String(), "err", err)
		status = 500
		msg = "Internal Error"
	}
	Err := Error{
		Error: msg,
	}
	WriteJSON(resp, status, Err)
}
