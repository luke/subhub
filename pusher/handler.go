package pusher

import (
	//	"errors"
	"github.com/igm/sockjs-go/sockjs"
	"net/http"
	//"net/url"
	// "regexp"
	// "strings"
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

//func (h *handler) parseSessionID(url *url.URL) (string, error) {
//	session := regexp.MustCompile(h.prefix + "/(?P<server>[^/.]+)/(?P<session>[^/.]+)/.*")
//	matches := session.FindStringSubmatch(url.Path)
//	if len(matches) == 3 {
//		return matches[2], nil
//	}
//	return "", errors.New("unable to parse URL for session")
//}

//func (h *handler) sessionById(sessionID string) (*session, error) {
//	h.sessionsMux.Lock()
//	defer h.sessionsMux.Unlock()
//	sess, exists := h.sessions[sessionID]
//	if !exists {
//		sess = newSession(sessionID, h.options.DisconnectDelay, h.options.HeartbeatDelay)
//		h.sessions[sessionID] = sess
//		if h.handlerFunc != nil {
//			go h.handlerFunc(sess)
//		}
//		go func() {
//			<-sess.closedNotify()
//			h.sessionsMux.Lock()
//			delete(h.sessions, sessionID)
//			h.sessionsMux.Unlock()
//		}()
//	}
//	return sess, nil
//}
