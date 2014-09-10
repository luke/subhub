package server

const (
	EVENT_CONNECTION_ESTABLISHED          = "pusher:connection_established"
	EVENT_PING                            = "pusher:ping"
	EVENT_PONG                            = "pusher:pong"
	EVENT_SUBSCRIBE                       = "pusher:subscribe"
	EVENT_UNSUBSCRIBE                     = "pusher:unsubscribe"
	EVENT_ERROR                           = "pusher:error"
	EVENT_INTERNAL_SUBSCRIPTION_SUCCEEDED = "pusher_internal:subscription_succeeded"
	// ??? is there unsubscription_succeeded too
	EVENT_INTERNAL_MEMBER_ADDED   = "pusher_internal:member_added"
	EVENT_INTERNAL_MEMBER_REMOVED = "pusher_internal:member_removed"
)

const (
	//Error Codes
	//4000-4099
	//Indicates an error resulting in the connection being closed by Pusher, and that attempting to reconnect using the same parameters will not succeed.
	ERROR_4000_APP_SSL_ONLY           = 4000 //4000: Application only accepts SSL connections, reconnect using wss://
	ERROR_4001_APP_NOT_FOUND          = 4001 //4001: Application does not exist
	ERROR_4003_APP_DISABLED           = 4003 //4003: Application disabled
	ERROR_4004_APP_OVER_QUOTA         = 4004 //4004: Application is over connection quota
	ERROR_4005_PATH_NOT_FOUND         = 4005 //4005: Path not found
	ERROR_4006_INVALID_VERSION_FORMAT = 4006 //4006: Invalid version string format
	ERROR_4007_BAD_PROTOCOL_VERSION   = 4007 //4007: Unsupported protocol version
	ERROR_4008_NO_PROTOCOL_VERSION    = 4008 //4008: No protocol version supplied
	//4100-4199
	//Indicates an error resulting in the connection being closed by Pusher, and that the client may reconnect after 1s or more.
	ERROR_4100_OVER_CAPACITY = 4100 //4100: Over capacity
	//4200-4299
	//Indicates an error resulting in the connection being closed by Pusher, and that the client may reconnect immediately.
	ERROR_4200_CLOSED_RECONNECT  = 4200 //4200: Generic reconnect immediately
	ERROR_4201_CLOSED_NO_PONG    = 4201 //4201: Pong reply not received: ping was sent to the client, but no reply was received - see ping and pong messages
	ERROR_4202_CLOSED_INACTIVITY = 4202 //4202: Closed after inactivity: Client has been inactive for a long time (currently 24 hours) and client does not support ping. Please upgrade to a newer WebSocket draft or implement version 5 or above of this protocol.
	//4300-4399
	//Any other type of error.
	ERROR_4301_RATE_LIMIT = 4301 //4301: Client event rejected due to rate limit
)

const RAW_CONNECTION_ESTABLISHED = `{"event":"pusher:connection_established","data":"{\"socket_id\":\"%s\",\"activity_timeout\":120}"}`
const RAW_PING = "{\"event\":\"pusher:ping\",\"data\":\"{}\"}"
const RAW_PONG = "{\"event\":\"pusher:pong\",\"data\":\"{}\"}"
const RAW_SUBSCRIPTION_SUCCEEDED = `{"event":"pusher_internal:subscription_succeeded","data":"{\"channel\":\"%s\"}"}`
const RAW_CHANNEL_EVENT = `{"event":"%s","channel":"%s","data":"%s"}`

type ConnectionEstablishedData struct {
	SocketId        string `json:"socket_id"`
	ActivityTimeout string `json:"activity_timeout"`
}

type SubscribeData struct {
	Channel     string                 `json:"channel"`
	Auth        string                 `json:"auth,omitempty"`
	ChannelData map[string]interface{} `json:"channel_data,omitempty"`
}

type UnsubscribeData struct {
	Channel string `json:"channel"`
}

type ErrorData struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type SubscriptionSucceededData struct {
	Presence map[string]interface{} `json:"presence"`
	// {
	// "ids": ["1","2"],
	// "hash": {"1": {"foo":"bar"}, "2": {"bar":"baz"}},
	// "count": "2"
	// }
}

type MemberAddedData struct {
	UserId   string                 `json:"user_id"`
	UserInfo map[string]interface{} `json:"user_info,omitempty"`
}

type MemberRemovedData struct {
	UserId string `json:"user_id"`
}
