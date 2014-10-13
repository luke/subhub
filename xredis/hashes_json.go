package xredis

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"log"
)

const DEFAULT_TAG_NAME = "json"

func (r *Redis) HMSetJSON(key string, obj interface{}) error {
	log.Println("hmset json")
	structs.DefaultTagName = DEFAULT_TAG_NAME
	pairs := make(map[string]string)
	s := structs.New(obj)
	fields := s.Fields()
	for _, f := range fields {
		fmt.Printf("field name: %+v\n", f.Name())
		if f.IsExported() {
			val, err := json.Marshal(f.Value())
			if err != nil {
				log.Println("problem encoding value as json", err)
				continue
			}
			log.Println("encoded value: ", string(val))
			pairs[f.Name()] = string(val)
		}
	}
	return r.HMSet(key, pairs)
}

func (r *Redis) HGetAllJSON(key string, obj interface{}) error {
	structs.DefaultTagName = DEFAULT_TAG_NAME
	s := structs.New(obj)
	hash, err := r.HGetAll(key)
	if err == nil {
		for k, v := range hash {
			f := s.Field(k)
			val := f.Value()
			jsonErr := json.Unmarshal([]byte(v), &val)
			if jsonErr != nil {
				log.Println("problem decoding value from json", err)
				continue
			}
			f.Set(val)
		}
	}
	return err
}
