// Package logs provides the standard interface and handler for OpenFaaS providers to expose function logs.
//
package logs

import (
	"fmt"
	"time"
)

// Request is the query to return the function logs.
type Request struct {
	// Name is the function name and is required
	Name string `json:"name"`
	// Instance is the optional container name, that allows you to request logs from a specific function instance
	Instance string `json:"instance"`
	// Since is the optional datetime value to start the logs from
	Since *time.Time `json:"since"`
	// Limit sets the maximum number of log messages to return, <=0 means unlimited
	Limit int `json:"limit"`
	// Follow is allows the user to request a stream of logs
	Follow bool `json:"follow"`
	// Pattern is an optional regexp value to filter the log messages
	Pattern *string `json:"pattern"`
	// Invert allows you to control if the Pattern should be matched or negated
	Invert bool `json:"invert"`
}

// String implements that Stringer interface and prints the log Request in a consistent way that
// allows you to safely compare if two requests have the same value.
func (r Request) String() string {
	pattern := ""
	if r.Pattern != nil {
		pattern = *r.Pattern
	}
	return fmt.Sprintf("name:%s instance:%s since:%v limit:%d follow:%v pattern:%v invert:%v", r.Name, r.Instance, r.Since, r.Limit, r.Follow, pattern, r.Invert)
}

// Message is a specific log message from a function container log stream
type Message struct {
	// Name is the function name
	Name string `json:"name"`
	// instance is the name/id of the specific function instance
	Instance string `json:"instance"`
	// Timestamp is the timestamp of when the log message was recorded
	Timestamp time.Time `json:"timestamp"`
	// Text is the raw log message content
	Text string `json:"text"`
}

// String implements the Stringer interface and allows for nice and simple string formatting of a log Message.
func (m Message) String() string {
	return fmt.Sprintf("%s %s (%s) %s", m.Timestamp.String(), m.Name, m.Instance, m.Text)
}
