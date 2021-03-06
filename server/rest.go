package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"time"
)

func newAuthMiddleware(s *server) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		// Set example variable
		c.Set("example", "12345")

		q := c.Request.URL.Query()
		// before request

		//Authentication
		//The following query parameters must be included with all requests, and are used to authenticate the request

		//auth_key	Your application key
		key := q["auth_key"][0]
		//auth_timestamp	The number of seconds since January 1, 1970 00:00:00 GMT. The server will only accept requests where the timestamp is within 600s of the current time
		timestamp := q["auth_timestamp"][0]
		//auth_version	Authentication version, currently 1.0
		version := q["auth_version"][0]
		//body_md5	If the request body is nonempty (for example for POST requests to `/events`), this parameter must contain the hexadecimal MD5 hash of the body
		bodyMD5 := q["body_md5"][0]
		//Once all the above parameters have been added to the request, a signature is calculated

		//auth_signature	Authentication signature, described below
		signature := q["auth_signature"][0]

		// POST\n/apps/3/events\nauth_key=278d425bdf160c739803&auth_timestamp=1353088179&auth_version=1.0&body_md5=ec365a775a4cd0599faeb73354201b6f

		input := fmt.Sprintf("%s\n%s\nauth_key=%s&auth_timestamp=%s&auth_version=%s&body_md5=%s", c.Request.Method, c.Request.URL.Path, key, timestamp, version, bodyMD5)

		secret, _ := s.lookupAuthSecret(key)

		valid := signature == hmacSha256HexSignature([]byte(input), []byte(secret))

		// todo: check md5 is actually correct

		// validate the signature..
		// valid := false

		if !valid {
			// is this the right status?
			c.Fail(401, errors.New("Invalid signature"))
		} else {
			c.Next()
		}

		// after request
		latency := time.Since(t)
		log.Print(latency)

		// access the status we are sending
		status := c.Writer.Status()
		log.Println(status)
	}
}

type EventJSON struct {
	Name     string   `json:"name" binding:"required"`
	Data     string   `json:"data" binding:"required"` // limited to 10KB
	Channels []string `json:"channels"`                // limited to 10 channels
	Channel  string   `json:"channel"`                 //  (can be used instead of channels)
	SocketId string   `json:"socket_id"`               // excludes the event from being sent to a specific connection
}

func (s *server) newRestApiHandler() http.Handler {

	r := gin.Default()
	r.Use(newAuthMiddleware(s))

	r.POST("/apps/:app_id/events", func(c *gin.Context) {
		// POST

		// The event data should not be larger than 10KB.
		// If you attempt to POST an event with a larger data parameter you will receive a 413 error code.

		var json EventJSON
		c.Bind(&json)

		// TODO: process

		// Response is an empty JSON hash.
		c.JSON(200, gin.H{})
	})

	r.GET("/apps/:app_id/channels", func(c *gin.Context) {

		q := c.Request.URL.Query()
		// filter_by_prefix	 Filter the returned channels by a specific prefix.
		// For example in order to return only presence channels you would set filter_by_prefix=presence-
		filterPrefix := q.Get("filter_by_prefix")
		// info	A comma separated list of attributes which should be returned for each channel.
		// If this parameter is missing, an empty hash of attributes will be returned for each channel.
		info := strings.Split(q.Get("info"), ",")
		// available attributes
		// user_count	Integer	Presence	Number of distinct users currently subscribed to this channel (a single user may be subscribed many times, but will only count as one)

		log.Println(filterPrefix)
		log.Println(info)

		channelMap := gin.H{"example": gin.H{"user_count": 0}}

		c.JSON(200, gin.H{"channels": channelMap})
		//{
		//  "channels": {
		//    "presence-foobar": {
		//      user_count: 42
		//    },
		//    "presence-another": {
		//      user_count: 123
		//    }
		//  }
		//}
	})

	r.GET("/apps/:app_id/channels/:channel_name", func(c *gin.Context) {

		q := c.Request.URL.Query()
		// info
		info := strings.Split(q.Get("info"), ",")
		//user_count	Integer	Presence	Number of distinct users currently subscribed to this channel (a single user may be subscribed many times, but will only count as one)
		//subscription_count	Integer	All	[BETA] Number of connections currently subscribed to this channel. This attribute is not available by default

		log.Println(info)

		c.JSON(200, gin.H{"occupied": true, "user_count": 42, "subscription_count": 42})
		//{
		//  occupied: true,
		//  user_count: 42,
		//  subscription_count: 42
		//}

	})

	r.GET("/apps/:app_id/channels/:channel_name/users", func(c *gin.Context) {

		// Note that only presence channels allow this functionality,
		// and a request to any other kind of channel will result in a 400 HTTP code."
		c.JSON(200, gin.H{"users": "..."})
		//{
		//  "users": [
		//    { "id": 1 },
		//    { "id": 2 }
		//  ]
		//}

	})

	return r

}
