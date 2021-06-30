// This is module to store/retrive infotmation of Telegraf instances and telegraf configuration
package main

import (
	"fmt"
	"container/list"
	"strings"

	"github.com/xujiajun/nutsdb"
)

type TelegrafInstanceInfo struct {
	ID string
	Name string
	DockerContainerID string
	Config string
	Description string
	Created string
}


// type TelegrafConfigInfo struct {
// 	ID string
// 	InstanceID string
// 	Created string
// }

type Storage interface {
	// Instance Handlers
	PutInstance(instanceInfo *TelegrafInstanceInfo) error
	RemoveInstance(instanceID string) error
	GetInstanceList() (list.List, error)

	// Configuration Handlers
	// PutConfig(config *TelegrafConfigInfo) error
	// RemoveConfig(configID string) error
	// GetConfigList() (list.List, error)
}

const LOCAL_STORAGE_DIR = "telegrafana-data"
const BUCKET_INSTANCE = "instance"
// const BUCKET_CONFIG = "config"
const LOCAL_STORAGE_KEY_SEP = "/"

// Instance fields
const BUCKET_INSTANCE_KEY_ID = "id"
const BUCKET_INSTANCE_KEY_NAME = "name"
const BUCKET_INSTANCE_KEY_CONTAINER = "container"
const BUCKET_INSTANCE_KEY_CONFIG = "config"
const BUCKET_INSTANCE_KEY_DESCRIPTION = "desc"
const BUCKET_INSTANCE_KEY_CREATED = "created"

type LocalStorage struct {
	db *nutsdb.DB
}

func NewLocalStorage() *LocalStorage {
	opt := nutsdb.DefaultOptions

	opt.Dir = LOCAL_STORAGE_DIR
	opt.SegmentSize = 1024 * 1024 // 1MB
	db, _ := nutsdb.Open(opt)

	return & LocalStorage {
		db: db,
	} 
}

func (s *LocalStorage)handleBucketKV(entries nutsdb.Entries) map[string]map[string]string {
	ret := map[string]map[string]string{}
	for _, entry := range entries {
		k := string(entry.Key)

		strs := strings.Split(k, LOCAL_STORAGE_KEY_SEP)

		// ID of the entry
		id := strs[0]

		// Field name
		f := strs[1]

		instance, exists := ret[id]
		if !exists {
			instance = map[string]string{}
			ret[id] = instance
		}

		instance[f] = string(entry.Value)
	}

	return ret
}

func (s *LocalStorage)toInstanceList(entries nutsdb.Entries) []TelegrafInstanceInfo {
	ret := make([]TelegrafInstanceInfo, 0, 100)
	for _, entry := range s.handleBucketKV(entries) {
		id, exists := entry[BUCKET_INSTANCE_KEY_ID]
		name, _ := entry[BUCKET_INSTANCE_KEY_NAME]
		container, _ := entry[BUCKET_INSTANCE_KEY_CONTAINER]
		config, _ := entry[BUCKET_INSTANCE_KEY_CONFIG]
		desc, _ := entry[BUCKET_INSTANCE_KEY_DESCRIPTION]
		created, _ := entry[BUCKET_INSTANCE_KEY_CREATED]

		if exists {
			ret = append(ret, TelegrafInstanceInfo{
				ID: id,
				Name: name,
				DockerContainerID: container,
				Config: config,
				Description: desc,
				Created: created,
			})
		}
	}

	return ret
}

func (s *LocalStorage) GetInstanceList() ([]TelegrafInstanceInfo, error) {
	var ret []TelegrafInstanceInfo = nil
	err := s.db.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(BUCKET_INSTANCE)
			if err != nil {
				return err
			}

			ret = s.toInstanceList(entries)
	
			return nil
		});
	
	return ret, err
}

func putFieldToBucket(tx *nutsdb.Tx, bucket string, id string, fieldKey string, fieldVal string) ([]byte, error) {
	key := getBucketKey(id, fieldKey)
	return key, tx.Put(bucket, key, []byte(fieldVal), 0)
}

func getBucketKey(id string, fieldKey string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", id, LOCAL_STORAGE_KEY_SEP, fieldKey))
}

func (s *LocalStorage) PutInstance(instanceInfo *TelegrafInstanceInfo) error {
	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			key1, err1 := putFieldToBucket(tx, BUCKET_INSTANCE, instanceInfo.ID, BUCKET_INSTANCE_KEY_ID, instanceInfo.ID)
			key2, err2 := putFieldToBucket(tx, BUCKET_INSTANCE, instanceInfo.ID, BUCKET_INSTANCE_KEY_NAME, instanceInfo.Name)
			key3, err3 := putFieldToBucket(tx, BUCKET_INSTANCE, instanceInfo.ID, BUCKET_INSTANCE_KEY_CONTAINER, instanceInfo.DockerContainerID)
			key4, err4 := putFieldToBucket(tx, BUCKET_INSTANCE, instanceInfo.ID, BUCKET_INSTANCE_KEY_CONFIG, instanceInfo.Config)
			key5, err5 := putFieldToBucket(tx, BUCKET_INSTANCE, instanceInfo.ID, BUCKET_INSTANCE_KEY_DESCRIPTION, instanceInfo.Description)
			key6, err6 := putFieldToBucket(tx, BUCKET_INSTANCE, instanceInfo.ID, BUCKET_INSTANCE_KEY_CREATED, instanceInfo.Created)

			if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
				// Delete all
				tx.Delete(BUCKET_INSTANCE, key1)
				tx.Delete(BUCKET_INSTANCE, key2)
				tx.Delete(BUCKET_INSTANCE, key3)
				tx.Delete(BUCKET_INSTANCE, key4)
				tx.Delete(BUCKET_INSTANCE, key5)
				tx.Delete(BUCKET_INSTANCE, key6)
			}
	
			return nil
		});
	
	return err
}

func (s *LocalStorage) RemoveInstance(instanceID string) error {
	err := s.db.Update(
		func(tx *nutsdb.Tx) error {
			tx.Delete(BUCKET_INSTANCE, getBucketKey(instanceID, BUCKET_INSTANCE_KEY_ID))
			tx.Delete(BUCKET_INSTANCE, getBucketKey(instanceID, BUCKET_INSTANCE_KEY_NAME))
			tx.Delete(BUCKET_INSTANCE, getBucketKey(instanceID, BUCKET_INSTANCE_KEY_CONTAINER))
			tx.Delete(BUCKET_INSTANCE, getBucketKey(instanceID, BUCKET_INSTANCE_KEY_CONFIG))
			tx.Delete(BUCKET_INSTANCE, getBucketKey(instanceID, BUCKET_INSTANCE_KEY_DESCRIPTION))
			tx.Delete(BUCKET_INSTANCE, getBucketKey(instanceID, BUCKET_INSTANCE_KEY_CREATED))
	
			return nil
		})
	
	return err
}

// Get the instance information from database
func (s *LocalStorage) GetInstance(instanceID string) (*TelegrafInstanceInfo, error) {
	var info *TelegrafInstanceInfo = nil
	var ret []TelegrafInstanceInfo
	err := s.db.View(
		func(tx *nutsdb.Tx) error {
			if entries, _, err := tx.PrefixScan(BUCKET_INSTANCE, []byte(instanceID + LOCAL_STORAGE_KEY_SEP), 0, 10); err != nil {
				return err
			} else {
				ret = s.toInstanceList(entries)
				return nil
			}
		})

	if len(ret) > 0 {
		info = &ret[0]
	}

	return info, err
}

func (s *LocalStorage) Stop() error {
	return s.db.Close()
}
