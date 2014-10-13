package server

import (
	"fmt"
	"log"
)

const REDIS_APP_SETTINGS_HASH = "subhub://app/%s/settings"

type AppSettings struct {
	Name               string `json:"name"`
	ForceEncryption    bool   `json:"force_encryption"`
	EnableClientEvents bool   `json:"enable_client_events"`
	// EnableKeyspaceEvents bool `json:"enable_keyspace_events"`
}

func (s *server) createApp(appId string) {
	log.Println("create app", appId)
	// adds the app, creates the keys, returns info
	key := fmt.Sprintf(REDIS_APP_SETTINGS_HASH, appId)
	err := s.redis.HMSetJSON(key, &AppSettings{Name: "foo", ForceEncryption: true, EnableClientEvents: false})
	if err != nil {
		log.Println("error", err)
	}
	// return err
}

func (s *server) loadApp(appId string) *AppSettings {
	key := fmt.Sprintf(REDIS_APP_SETTINGS_HASH, appId)
	settings := &AppSettings{}
	err := s.redis.HGetAllJSON(key, settings)
	if err != nil {
		log.Println("error fetching settings")
	}
	return settings
}

func (s *server) deleteApp(appId string) {
	_ = fmt.Sprintf(REDIS_APP_SETTINGS_HASH, appId)
}

func (s *server) saveAppSettings(appId string, settings *AppSettings) {
	_ = fmt.Sprintf(REDIS_APP_SETTINGS_HASH, appId)
}
