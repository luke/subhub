package server

func (s *server) createWebhook(appId string, url string, eventId string) {

}

func (s *server) disableWebhook(webhookId string) {

}

func (s *server) enableWebhook(webhookId string) {

}

func (s *server) deleteWebhook(webhookId string) {

}

func (s *server) listWebhooks(appId string) {

}

func (s *server) callWebhooks(appId string, eventId string) {

}

// channel_occupied
// channel_vacated
// member_added
// member_removed
// client_event

//{
//  "name": "client_event",
//  "channel": "name of the channel the event was published on",
//  "event": "name of the event",
//  "data": "data associated with the event",
//  "socket_id": "socket_id of the sending socket",
//  "user_id": "user_id associated with the sending socket" # Only for presence channels
//}

//{
//  "time_ms": 1327078148132
//  "events": [
//    { "name": "event_name", "some": "data" }
//  ]
//}

//X-Pusher-Key: A Pusher app may have multiple tokens. The oldest active token will be used, identified by this key.
//X-Pusher-Signature: A HMAC SHA256 hex digest formed by signing the POST payload (body) with the token’s secret.
