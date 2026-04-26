package web

import (
	"encoding/json"
	"net/http"
)

// WriteJSON marshalls an object to json.
func WriteJSON(w http.ResponseWriter, statusCode int, response interface{}) error {
	w.Header().Set(HeaderContentType, ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(response)
}
