package httpapi

import (
	"crypthttp/domain"
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	return nil
}
func decodeJSON(r *http.Request, w http.ResponseWriter, creds any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		return fmt.Errorf("%w : %v", domain.ErrInvalidCredentials, err)
	}
	return nil
}
