package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

type Publicer interface {
	Public() any
}

type privilegedResponse struct {
	reason string
	value  any
}

func allowPrivileged(reason string, value any) privilegedResponse {
	return privilegedResponse{
		reason: reason,
		value:  value,
	}
}

func writePublic(w http.ResponseWriter, status int, value Publicer) {
	writeRawJSON(w, status, value.Public())
}

func writePublicList[T Publicer](w http.ResponseWriter, status int, values []T) {
	publicValues := make([]any, 0, len(values))

	for _, value := range values {
		publicValues = append(publicValues, value.Public())
	}

	writeRawJSON(w, status, publicValues)
}

func writePrivileged(w http.ResponseWriter, status int, response privilegedResponse) {
	if response.reason == "" {
		fmt.Println("privileged response missing reason")
		writeInternalError(w)
		return
	}

	writeRawJSON(w, status, response.value)
}

func writeRawJSON(w http.ResponseWriter, status int, value any) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(value); err != nil {
		fmt.Println("write json error:", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		_ = json.NewEncoder(w).Encode(errorResponse{
			Error: "internal server error",
		})

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(buf.Bytes()); err != nil {
		fmt.Println("write response error:", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeRawJSON(w, status, errorResponse{
		Error: message,
	})
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "internal server error")
}
