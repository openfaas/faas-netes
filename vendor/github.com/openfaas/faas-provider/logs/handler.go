package logs

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/openfaas/faas-provider/httputils"
)

var upgrader = websocket.Upgrader{} // use default options

// Requestor submits queries the logging system.
// This will be passed to the log handler constructor.
type Requestor interface {
	// Query submits a log request to the actual logging system.
	Query(context.Context, Request) (<-chan Message, error)
}

// NewSimpleLogHandlerFunc creates and http HandlerFunc from the supplied log Requestor.
func NewSimpleLogHandlerFunc(requestor Requestor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		cn, ok := w.(http.CloseNotifier)
		if !ok {
			log.Println("LogHandler: response is not a CloseNotifier, required for streaming response")
			http.NotFound(w, r)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Println("LogHandler: response is not a Flusher, required for streaming response")
			http.NotFound(w, r)
			return
		}

		logRequest, err := parseRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			httputils.WriteError(w, http.StatusUnprocessableEntity, "could not parse the log request")
			return
		}

		ctx, cancelQuery := context.WithCancel(r.Context())
		defer cancelQuery()
		messages, err := requestor.Query(ctx, logRequest)
		if err != nil {
			// add smarter error handling here
			httputils.WriteError(w, http.StatusInternalServerError, "function log request failed")
			return
		}

		// Send the initial headers saying we're gonna stream the response.
		w.Header().Set("Connection", "Keep-Alive")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		sent := 0
		jsonEncoder := json.NewEncoder(w)

		if logRequest.Limit > 0 {
			log.Printf("LogHandler: watch for and stream `%d` log messages\n", logRequest.Limit)
		}

		msgFilter := newFilter(logRequest)

		for messages != nil {
			select {
			case <-cn.CloseNotify():
				log.Println("LogHandler: client stopped listening")
				return
			case msg, ok := <-messages:
				if !ok {
					log.Println("LogHandler: end of log stream")
					messages = nil
					return
				}

				if !msgFilter(&msg) {
					continue
				}
				// serialize and write the msg to the http ResponseWriter
				err := jsonEncoder.Encode(msg)
				if err != nil {
					// can't actually write the status header here so we should json serialize an error
					// and return that because we have already sent the content type and status code
					log.Printf("LogHandler: failed to serialize log message: '%s'\n", msg.String())
					log.Println(err.Error())
					// write json error message here ?
					jsonEncoder.Encode(Message{Text: "failed to serialize log message"})
					return
				}

				flusher.Flush()

				if logRequest.Limit > 0 {
					sent++
					if sent >= logRequest.Limit {
						log.Printf("LogHandler: reached message limit '%d'\n", logRequest.Limit)
						return
					}
				}
			}
		}

		return
	}
}

// NewLogHandlerFunc creates and http HandlerFunc from the supplied log Requestor.
func NewLogHandlerFunc(requestor Requestor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			log.Println("LogHandler: response is not a Hijacker, required for streaming response")
			http.NotFound(w, r)
			return
		}

		logRequest, err := parseRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			httputils.WriteError(w, http.StatusUnprocessableEntity, "could not parse the log request")
			return
		}

		ctx, cancelQuery := context.WithCancel(r.Context())
		defer cancelQuery() // allows us to cancel the query stream by simply returning from the handler

		messages, err := requestor.Query(ctx, logRequest)
		if err != nil {
			// add smarter error handling here?
			httputils.WriteError(w, http.StatusInternalServerError, "function log request failed")
			return
		}

		// Send the initial headers saying we're gonna stream the response.
		w.Header().Set(http.CanonicalHeaderKey("Connection"), "Keep-Alive")
		w.Header().Set(http.CanonicalHeaderKey("Transfer-Encoding"), "chunked")
		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/x-ndjson")
		w.WriteHeader(http.StatusOK)

		conn, buf, err := hijacker.Hijack()
		if err != nil {
			log.Println("LogHandler: failed to hijack connection for streaming response")
			return
		}
		defer conn.Close()
		conn.SetWriteDeadline(time.Time{}) // allow arbitrary time between log writes
		buf.Flush()                        // will write the headers and the initial 200 response

		// using NewChunkedWriter ensures that data is written with the correct chunked format,
		// e.g. line endings, it will not ensure proper closing trailers
		jsonEncoder := json.NewEncoder(httputil.NewChunkedWriter(buf))

		defer func() {
			// try to write the required closing newliens for chunked encoding
			// see the RFC https://tools.ietf.org/html/rfc7230#section-4.1 specification of
			// the last chunk `last-chunk     = 1*("0") [ chunk-ext ] CRLF`
			buf.WriteString("0\r\n\r\n")
			buf.Flush()
		}()

		if logRequest.Limit > 0 {
			log.Printf("LogHandler: watch for and stream `%d` log messages\n", logRequest.Limit)
		}

		sent := 0 // used to enforce number of logs limit

		closed := closeNotify(ctx, conn)
		msgFilter := newFilter(logRequest)

		for messages != nil {
			select {
			case <-closed:
				log.Println("LogHandler: connection closed")
				return
			case <-ctx.Done():
				log.Println("LogHandler: context done")
				return
			case msg, ok := <-messages:
				if !ok {
					log.Println("LogHandler: end of log stream")
					messages = nil
					return
				}

				if !msgFilter(&msg) {
					continue
				}
				// serialize and write the msg to the http ResponseWriter
				err := jsonEncoder.Encode(msg)
				if err != nil {
					// can't actually write the status header here because we already sent a 200

					// we can:
					// 1. json serialize an error and return that
					// 2. use Trailers to send error information, this is expected behavior for
					//    chunked encoding, but probably hard for the client to parse
					log.Printf("LogHandler: failed to serialize log message: '%s'\n", msg.String())
					log.Println(err.Error())
					// write json error message here ?
					jsonEncoder.Encode(Message{Text: "failed to serialize log message"})
					return
				}
				// actually send the log line to the client
				buf.Flush()

				// only track logs sent _if_ we need to
				if logRequest.Limit > 0 {
					sent++
					if sent >= logRequest.Limit {
						log.Printf("LogHandler: reached message limit '%d'\n", logRequest.Limit)
						return
					}
				}
			}
		}

		return
	}
}

// closeNotify will watch the connection and notify when then connection is closed
func closeNotify(ctx context.Context, c net.Conn) <-chan error {
	notify := make(chan error, 1)

	go func() {
		buf := make([]byte, 1)
		// blocks until non-zero read or error.  From the fd.Read docs:
		// If the caller wanted a zero byte read, return immediately
		// without trying (but after acquiring the readLock).
		// Otherwise syscall.Read returns 0, nil which looks like
		// io.EOF.
		// It is important that `buf` is allocated a non-zero size
		n, err := c.Read(buf)
		if err != nil {
			log.Printf("LogHandler: test connection: %s\n", err)
			notify <- err
			return
		}
		if n > 0 {
			log.Printf("LogHandler: unexpected data: %s\n", buf[:n])
			return
		}
	}()
	return notify
}

// NewWSLogHandlerFunc creates and http HandlerFunc from the supplied log Requestor.
func NewWSLogHandlerFunc(requestor Requestor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("LogHandler: unable to upgrade response to a websocket: %s\n", err.Error())
			http.NotFound(w, r)
			return
		}
		defer c.Close()

		logRequest, err := parseRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			httputils.WriteError(w, http.StatusUnprocessableEntity, "could not parse the log request")
			return
		}

		ctx, cancelQuery := context.WithCancel(r.Context())
		defer cancelQuery()
		messages, err := requestor.Query(ctx, logRequest)
		if err != nil {
			// add smarter error handling here
			httputils.WriteError(w, http.StatusInternalServerError, "function log request failed")
			return
		}

		sent := 0

		if logRequest.Limit > 0 {
			log.Printf("LogHandler: watch for and stream `%d` log messages\n", logRequest.Limit)
		}

		msgFilter := newFilter(logRequest)

		for messages != nil {
			select {
			case msg, ok := <-messages:
				if !ok {
					log.Println("LogHandler: end of log stream")
					messages = nil
					return
				}

				if !msgFilter(&msg) {
					continue
				}

				// serialize and write the msg to the http ResponseWriter
				err := c.WriteJSON(msg)
				if err != nil {
					// can't actually write the status header here so we should json serialize an error
					// and return that because we have already sent the content type and status code
					log.Printf("LogHandler: failed to serialize log message: '%s'\n", msg.String())
					log.Println(err.Error())
					// write json error message here ?
					c.WriteJSON(Message{Text: "failed to serialize log message"})
					return
				}

				if logRequest.Limit > 0 {
					sent++
					if sent >= logRequest.Limit {
						log.Printf("LogHandler: reached message limit '%d'\n", logRequest.Limit)
						return
					}
				}
			}
		}

		return
	}
}

// parseRequest extracts the logRequest from the GET variables or from the POST body
func parseRequest(r *http.Request) (logRequest Request, err error) {
	switch r.Method {
	case http.MethodGet:
		query := r.URL.Query()
		logRequest.Name = getValue(query, "name")
		logRequest.Instance = getValue(query, "instance")
		limitStr := getValue(query, "limit")
		if limitStr != "" {
			logRequest.Limit, err = strconv.Atoi(limitStr)
			if err != nil {
				return logRequest, err
			}
		}
		// ignore error because it will default to false if we can't parse it
		logRequest.Follow, _ = strconv.ParseBool(getValue(query, "follow"))
		logRequest.Invert, _ = strconv.ParseBool(getValue(query, "invert"))

		sinceStr := getValue(query, "since")
		if sinceStr != "" {
			since, err := time.Parse(time.RFC3339, sinceStr)
			logRequest.Since = &since
			if err != nil {
				return logRequest, err
			}
		}

		// don't use getValue here so that we can detect if the value is nil or empty
		patterns := query["pattern"]
		if len(patterns) > 0 {
			logRequest.Pattern = &(patterns[len(patterns)-1])
		}

	case http.MethodPost:
		err = json.NewDecoder(r.Body).Decode(&logRequest)
	}

	return logRequest, err
}

// getValue returns the value for the given key. If the key has more than one value, it returns the
// last value. if the value does not exist, it returns the empty string.
func getValue(queryValues url.Values, name string) string {
	values := queryValues[name]
	if len(values) == 0 {
		return ""
	}

	return values[len(values)-1]
}

// Filter implements the filter logic for the Requestor interface
func newFilter(r Request) func(m *Message) bool {

	var pattern *regexp.Regexp
	if r.Pattern != nil {
		var err error
		pattern, err = regexp.Compile(*r.Pattern)
		if err != nil {
			log.Printf("LogRequestor: failed to compile request Pattern: %s", err.Error())
		}
	}

	return func(m *Message) bool {
		if r.Invert {
			return matchesInstance(r.Instance, m) && !matchesPattern(pattern, m)
		}
		return matchesInstance(r.Instance, m) && matchesPattern(pattern, m)
	}
}

func matchesInstance(instance string, m *Message) bool {
	if instance == "" {
		return true
	}

	return instance == m.Instance
}

func matchesPattern(r *regexp.Regexp, m *Message) bool {
	if r == nil {
		return true
	}

	return r.MatchString(m.Text)
}
