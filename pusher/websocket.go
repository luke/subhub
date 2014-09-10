package pusher

import (
	"log"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketReadBufSize is a parameter that is used for WebSocket Upgrader.
// https://github.com/gorilla/websocket/blob/master/server.go#L230
var WebSocketReadBufSize = 4096

// WebSocketWriteBufSize is a parameter that is used for WebSocket Upgrader
// https://github.com/gorilla/websocket/blob/master/server.go#L230
var WebSocketWriteBufSize = 4096

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  WebSocketReadBufSize,
	WriteBufferSize: WebSocketWriteBufSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *handler) handleWebsocket(rw http.ResponseWriter, req *http.Request) {
	// conn, err := websocket.Upgrade(rw, req, nil, WebSocketReadBufSize, WebSocketWriteBufSize)

	req.ParseForm()
	sid := req.Form.Get("sid") // , _ := h.parseSessionID(req.URL)
	if sid == "" {
		sid = uuid.NewRandom().String()
	}
	rw.Header().Add("sid", sid)

	conn, err := wsUpgrader.Upgrade(rw, req, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(rw, `Can "Upgrade" only to "WebSocket".`, http.StatusBadRequest)
		return
	} else if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	path := req.RequestURI
	sess := newSession(sid, h.options.DisconnectDelay, h.options.HeartbeatDelay, path)
	if h.handlerFunc != nil {
		go h.handlerFunc(sess)
	}

	receiver := newWsReceiver(conn)
	sess.attachReceiver(receiver)
	readCloseCh := make(chan struct{})
	go func() {
		for {

			// todo: refactor needed
			// kind of silly to be doing all this
			// read > accept > encode > decode > yada lark
			// can just do.. c.ws.ReadJSON(&msg)

			t, m, err := conn.ReadMessage()
			if err != nil {
				log.Println("error reading message, calling close", err)
				close(readCloseCh)
				return
			}
			if t != websocket.TextMessage {
				log.Println("only accept text messages, closing")
				close(readCloseCh)
				return
			}
			sess.accept(string(m))
		}
	}()

	select {
	case <-readCloseCh:
	case <-receiver.doneNotify():
	}
	sess.close()
	conn.Close()
}

type wsReceiver struct {
	conn    *websocket.Conn
	closeCh chan struct{}
}

func newWsReceiver(conn *websocket.Conn) *wsReceiver {
	return &wsReceiver{
		conn:    conn,
		closeCh: make(chan struct{}),
	}
}

func (w *wsReceiver) sendBulk(messages ...string) {
	if len(messages) > 0 {
		for _, message := range messages {
			w.sendFrame(message)
		}
	}
}

func (w *wsReceiver) sendFrame(frame string) {
	if err := w.conn.WriteMessage(websocket.TextMessage, []byte(frame)); err != nil {
		log.Println("error sending message, calling close", err.Error())
		w.close()
	}
}

func (w *wsReceiver) close() {
	select {
	case <-w.closeCh: // already closed
	default:
		close(w.closeCh)
	}
}
func (w *wsReceiver) doneNotify() <-chan struct{}        { return w.closeCh }
func (w *wsReceiver) interruptedNotify() <-chan struct{} { return nil }
