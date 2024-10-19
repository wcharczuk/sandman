package web

import (
	"encoding/gob"
	"net/http"
)

// WriteGob marshalls an object to gob.
func WriteGob(w http.ResponseWriter, statusCode int, response interface{}) error {
	w.Header().Set(HeaderContentType, ContentTypeApplicationGob)
	w.WriteHeader(statusCode)
	return gob.NewEncoder(w).Encode(response)
}
