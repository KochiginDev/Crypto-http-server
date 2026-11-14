package types

import (
	"encoding/json"
	"fmt"
	"myapi/domain"
	"net/http"
)

type GetObjectHandlerRequest struct {
	Key string `json:"key"`
}

func CreateGetObjectHandlerRequest(r *http.Request) (*GetObjectHandlerRequest, error) {
	key := r.URL.Query().Get("key")
	if key == "" {
		return nil, fmt.Errorf("missing key")
	}
	return &GetObjectHandlerRequest{Key: key}, nil
}

type PostObjectHandlerRequest struct {
	domain.Object
}

func CreatePostObjectHandlerRequest(r *http.Request) (*PostObjectHandlerRequest, error) {
	var req PostObjectHandlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}
	return &req, nil
}

type DeleteObjectHandlerRequest struct {
	Key string `json:"key"`
}

func CreateDeleteObjectHandlerRequest(r *http.Request) (*DeleteObjectHandlerRequest, error) {
	key := r.URL.Query().Get("key")
	if key == "" {
		return nil, fmt.Errorf("missing key")
	}
	return &DeleteObjectHandlerRequest{Key: key}, nil
}

type PutObjectHandlerRequest struct {
	domain.Object
}

func CreatePutObjectHandlerRequest(r *http.Request) (*PutObjectHandlerRequest, error) {
	key := r.URL.Query().Get("key")
	if key == "" {
		return nil, fmt.Errorf("missing key")
	}
	var req PutObjectHandlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %v", err)
	}
	return &req, nil
}
