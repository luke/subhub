package pusher

import (
	"github.com/igm/sockjs-go/sockjs"
	"net/http"
	"sync"
)

type handler struct {
	prefix      string
	options     sockjs.Options
	handlerFunc func(Session)

	sessionsMux sync.Mutex
	sessions    map[string]*session
}

// NewHandler creates new HTTP handler that conforms to the basic net/http.Handler interface.
// It takes path prefix, options and sockjs handler function as parameters
func NewHandler(prefix string, opts sockjs.Options, handleFunc func(Session)) http.Handler {
	return newHandler(prefix, opts, handleFunc)
}

func newHandler(prefix string, opts sockjs.Options, handlerFunc func(Session)) *handler {
	h := &handler{
		prefix:      prefix,
		options:     opts,
		handlerFunc: handlerFunc,
		sessions:    make(map[string]*session),
	}
	return h
}

func (h *handler) Prefix() string { return h.prefix }

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.handleWebsocket(rw, req)
}
