package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strings"
)

func (s *server) verifyAuth(auth string, message string) bool {
	log.Println("message", message)
	parts := strings.Split(auth, ":")
	authKey := parts[0]
	authSecret, _ := s.lookupAuthSecret(authKey)
	log.Println("key", authKey)
	log.Println("secret", authSecret)
	hmac0 := parts[1]
	hmac1 := hmacSha256HexSignature([]byte(message), []byte(authSecret))
	log.Println("hmac", hmac0, hmac1)
	return hmac0 == hmac1
}

func (s *server) verifyToken(tokenData string) {

	token, err := jwt.Parse(tokenData, func(token *jwt.Token) (interface{}, error) {
		return s.lookupAuthSecret(token.Header["kid"].(string))
	})

	if err == nil && token.Valid {
		log.Println("token is valid !")
	} else {
		log.Println("token is invalid :(")
	}

}

const REDIS_AUTH_KEYS = "subhub://auth/keys"

func (s *server) saveAuth(key string, secret string) error {
	_, err := s.redis.HSet(REDIS_AUTH_KEYS, key, secret)
	return err
}

func (s *server) lookupAuthSecret(key string) (string, error) {
	// todo: add a key cache here..
	secret, err := s.redis.HGet(REDIS_AUTH_KEYS, key)
	if err != nil {
		return "", err
	}
	return string(secret), nil
}

// checkMAC returns true if messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func hmacSha256HexSignature(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

type authHandlerFunc func(w http.ResponseWriter, r *http.Request)

func (s *server) newAuthHandlerFunc() authHandlerFunc {

	// todo save the key here too..
	authKey := "278d425bdf160c739803"
	authSecret := "7ad3773142a6692b25b8"

	s.saveAuth(authKey, authSecret)

	// var err error
	authSecret, _ = s.lookupAuthSecret(authKey)

	type Response struct {
		Auth        string `json:"auth"`
		ChannelData string `json:"channel_data,omitempty"`
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("auth handler called")
		// todo: implement me

		// parse post values
		err := r.ParseForm()

		if err != nil {
			log.Println("unable to parse form vars")
			return
		}

		socketId := r.Form.Get("socket_id")
		channelName := r.Form.Get("channel_name")
		callback := r.Form.Get("callback")

		if ok := s.authorizeSocketAccessToChannel(socketId, channelName, r); !ok {
			log.Println("auth denied")
			return
		}

		var channelData string // = nil

		message := fmt.Sprintf("%s:%s", socketId, channelName)

		if strings.HasPrefix(channelName, CHANNEL_PREFIX_PRESENSE) {
			// double escaped json
			channelData = fmt.Sprintf("{\"user_id\":\"sock%s\",\"user_info\":{\"name\":\"Socket %s \"}}", socketId, socketId)
			message = fmt.Sprintf("%s:%s", message, channelData)
		}

		signature := hmacSha256HexSignature([]byte(message), []byte(authSecret))
		auth := fmt.Sprintf("%s:%s", authKey, signature)

		resp := &Response{Auth: auth, ChannelData: channelData}
		data, _ := json.Marshal(resp)

		log.Println("resp %+v", resp)
		log.Println("callback %s", callback)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Add("Access-Control-Allow-Methods", "*")
		w.Header().Add("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))

		if len(callback) > 0 {
			w.Header().Set("Content-Type", "text/javascript")
			data = []byte(fmt.Sprintf("%s(%s);\n", callback, data))
		} else {
			w.Header().Set("Content-Type", "application/json")
		}
		w.Write(data)
	}
	return handler
}

func (s *server) authorizeSocketAccessToChannel(socketId string, channelName string, r *http.Request) bool {
	// for now we return true always
	return true
}
