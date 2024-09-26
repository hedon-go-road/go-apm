package dogapm

import (
	"encoding/json"
	"log"
	"net/http"
)

type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Body    any    `json:"body"`
}

type httpStatus struct{}

var HttpStatus = httpStatus{}

func (h *httpStatus) Ok(w http.ResponseWriter) {
	status := &Status{
		Code:    http.StatusOK,
		Message: "success",
	}

	data, _ := json.Marshal(status)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status.Code)
	_, err := w.Write(data)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *httpStatus) OkBody(w http.ResponseWriter, msg string, body any) {
	status := &Status{
		Code:    http.StatusOK,
		Message: msg,
		Body:    body,
	}

	data, _ := json.Marshal(status)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status.Code)
	_, err := w.Write(data)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *httpStatus) Fail(w http.ResponseWriter, msg string, body any) {
	status := &Status{
		Code:    http.StatusBadRequest,
		Message: msg,
		Body:    body,
	}

	data, _ := json.Marshal(status)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status.Code)
	_, err := w.Write(data)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *httpStatus) Error(w http.ResponseWriter, msg string, body any) {
	status := &Status{
		Code:    http.StatusInternalServerError,
		Message: msg,
		Body:    body,
	}

	data, _ := json.Marshal(status)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status.Code)
	_, err := w.Write(data)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
