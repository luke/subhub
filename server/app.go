package server

import (
	"fmt"
	"github.com/screencloud/subhub/uuid"
	"log"
)

const REDIS_APP_SETTINGS_HASH = "subhub://app/%s/settings"

type AppSettings struct {
	Name               string `json:"name"`
	ForceEncryption    bool   `json:"force_encryption"`
	EnableClientEvents bool   `json:"enable_client_events"`
	// EnableKeyspaceEvents bool `json:"enable_keyspace_events"`
}

func newId() string {
	return uuid.NewRandom().String()
}

func appKey(appId string) string {
	return fmt.Sprintf(REDIS_APP_SETTINGS_HASH, appId)
}

func (s *server) createApp() string {
	appId := newId()
	log.Println("create app", appId)
	key := appKey(appId)
	err := s.redis.HMSetJSON(key, &AppSettings{})
	if err != nil {
		log.Println("error", err)
	}
	// TODO: generate auth keys at the same time
	return appId
}

func (s *server) loadApp(appId string) *AppSettings {
	key := appKey(appId)
	settings := &AppSettings{}
	err := s.redis.HGetAllJSON(key, settings)
	if err != nil {
		log.Println("error fetching settings")
	}
	return settings
}

func (s *server) deleteApp(appId string) {
	key := appKey(appId)
	_, err := s.redis.Del(key)
	if err != nil {
		log.Println("error deleting app", err)
	}
}

func (s *server) saveApp(appId string, settings *AppSettings) {
	key := appKey(appId)
	err := s.redis.HMSetJSON(key, settings)
	if err != nil {
		log.Println("problem saving app", err)
	}
}

// TODO: add api for these methods with global auth, perhaps simple basic auth?

func (s *server) testApp() {
	appId := s.createApp()
	settings := s.loadApp(appId)
	log.Printf("%+v", settings)
	settings.Name = "test"
	s.saveApp(appId, settings)
	settings2 := s.loadApp(appId)
	log.Printf("%+v", settings2)
	s.deleteApp(appId)

}
