package httputils

import (
	"fmt"
	"net/http"
)

// WriteError sets the response status code and write formats the provided message as the
// response body
func WriteError(w http.ResponseWriter, statusCode int, msg string, args ...interface{}) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(msg, args...)))
}
