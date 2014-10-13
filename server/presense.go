package server

import (
	"encoding/json"
	"fmt"
	"github.com/screencloud/subhub/pubsub"
	"log"
	"strconv"
)

const REDIS_CHANNEL_MEMBERS_HASH = "subhub://channel/%s/members"

func (s *server) presenseMemberAdded(sock *socket, channel string, userId string, userData interface{}) {
	userDataJSON, _ := json.Marshal(userData)
	msg := &pubsub.Message{
		Name: EVENT_INTERNAL_MEMBER_ADDED, //  "pusher_internal:member_removed",
		Data: fmt.Sprintf("{\"user_id\": \"%s\", \"user_data\": %s}", userId, userDataJSON),
	}
	key := fmt.Sprintf(REDIS_CHANNEL_MEMBERS_HASH, channel)
	log.Println("save", key, userId, string(userDataJSON))
	resp, err := s.redis.HSet(key, userId, string(userDataJSON))
	log.Println("resp", resp)
	if err != nil {
		log.Println("problem adding member to hash", err, resp)
	}
	s.pubsub.Publish(sock, channel, msg)
}

func (s *server) presenseMemberRemoved(sock *socket, channel string, userId string) {
	msg := &pubsub.Message{
		Name: EVENT_INTERNAL_MEMBER_REMOVED, //  "pusher_internal:member_removed",
		Data: fmt.Sprintf("{\"user_id\": \"%s\"}", userId),
	}
	s.redis.HDel(fmt.Sprintf(REDIS_CHANNEL_MEMBERS_HASH, channel), userId)
	s.pubsub.Publish(sock, channel, msg)
}

func (s *server) handleSubscribePresense(sock *socket, channel string, channelData string) {

	_, alreadySubscribed := sock.presense[channel]

	if alreadySubscribed {
		log.Println("already subscribed")
		return
	}

	// notify other members we have been added
	memberData := &MemberAddedData{}
	log.Println("channelData", channelData)
	err := json.Unmarshal([]byte(channelData), memberData)
	if err != nil {
		log.Println("error decoding json", err)
	}
	log.Printf("memberData %+v", memberData)
	userId := fmt.Sprintf("%v", memberData.UserId)
	s.presenseMemberAdded(sock, channel, userId, memberData.UserInfo)

	// add to the sock
	sock.presense[channel] = userId

	// read from master here as we cant be sure its synced to client
	key := fmt.Sprintf(REDIS_CHANNEL_MEMBERS_HASH, channel)
	log.Println("key", key)
	members, err := s.redis.HGetAll(key)

	log.Printf("redis hgetall %+v %+v", members, err)

	var ids []string
	var hash = make(map[string]interface{})
	for id := range members {
		ids = append(ids, id)
		var v interface{}
		_ = json.Unmarshal([]byte(members[id]), &v)
		hash[id] = v
	}

	presence := make(map[string]interface{})
	presence["ids"] = ids
	presence["hash"] = hash
	presence["count"] = len(ids)

	s.pubsub.Subscribe(sock, channel)

	data := make(map[string]interface{})
	data["presence"] = presence

	// resp := &PresenseSubscriptionSucceededData{Presense: presense}
	jsonEncodedData, _ := json.Marshal(data)

	resp := fmt.Sprintf(RAW_SUBSCRIPTION_SUCCEEDED, channel, strconv.Quote(string(jsonEncodedData)))
	log.Println("resp", resp)
	sock.session.Send(resp)

}

func (s *server) handleUnsubscribePresense(sock *socket, channel string) {
	// lookup the userId
	userId := sock.presense[channel]
	// remove from the map
	delete(sock.presense, channel)
	s.presenseMemberRemoved(sock, channel, userId)
}
