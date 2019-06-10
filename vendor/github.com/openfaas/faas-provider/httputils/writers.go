package httputils

import (
	"fmt"
	"net/http"
)

// ErrorF sets the response status code and write formats the provided message as the
// response body
func ErrorF(w http.ResponseWriter, statusCode int, msg string, args ...interface{}) {
	http.Error(w, fmt.Sprintf(msg, args...), statusCode)
}
