package checkered

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func EnableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		EnableCORS(w)
		next.ServeHTTP(w, r)
	})
}

// Parses JSON data from a request. Returns an error and HTTP status code if the
// parsing fails.
func ParseJsonRequestData[T any](req *http.Request) (T, error, int) {
	var data T
	// Parse the request body
	body, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return data, fmt.Errorf("failed to read the request body: %w", err), http.StatusInternalServerError
	}

	// Parse the JSON request data
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&data)
	if err != nil {
		return data, fmt.Errorf("invalid request data: %w", err), http.StatusBadRequest
	}
	return data, nil, 0
}

// Parses JSON data from a response. Returns an error if the parsing fails.
func ParseJsonResponseData[T any](res *http.Response) (T, error) {
	var data T
	// Parse the response body
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return data, fmt.Errorf("failed to read the response body: %w", err)
	}

	// Parse the JSON response data
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&data)
	if err != nil {
		return data, fmt.Errorf("invalid response data: %w", err)
	}
	return data, nil
}
