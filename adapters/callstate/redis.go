package callstate

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/neoms/adapters"
	"github.com/neoms/config"
	"github.com/neoms/logger"
	"time"
)

var ctx = context.Background()

type CallState struct {
	client *redis.Client
}

func NewCallStateAdapter() (adapters.CallStateAdapter,error) {
	redisHostPort := config.Config.Redis.RedisHost + ":" + config.Config.Redis.RedisPort
	logger.Logger.Debug("Redis Config :", redisHostPort)
	 return &CallState{
		client: redis.NewClient(&redis.Options{
			Addr:         redisHostPort,
			MinIdleConns: 5,
			MaxRetries:   3,
			Password:     "", // no password set
			DB:           0,  // use default DB
		}),
	}, nil
}

func (cs CallState) Get(callUUID string) ([]byte, error) {
	val, err := cs.client.Get(ctx, callUUID).Bytes()
	if err != nil {
		return []byte("UNKNOWN"), err
	}
	return val, nil
}

func (cs CallState) Set(callUuid string, state []byte, expired ...int) error {
	expire := 0 * time.Second
	if len(expired) > 0 {
		expire = time.Duration(expired[0])
	}
	err := cs.client.Set(ctx, callUuid, state, expire*time.Second).Err()
	return err
}

func (cs CallState) Del(callUUID string) error {
	return cs.client.Del(ctx, callUUID).Err()
}



func (cs CallState) KeyExist(key string) (bool, error) {
	val, err := cs.client.Exists(ctx, key).Result()
	if val != 1 || err != nil {
		return false, err
	}
	return true, nil
}

func (cs CallState) SetRecordingJob(state []byte) error {
	recordingJobKey := fmt.Sprint("tiniyo_namespace:jobs:s3_upload")
	recordingRangeKey := "tiniyo_namespace:known_jobs"
	var err error
	if err = cs.client.LPush(ctx, recordingJobKey, state).Err(); err == nil {
		err = cs.client.SAdd(ctx, recordingRangeKey, "s3_upload").Err()
	}
	return err
}

func (cs CallState) AddSetMember(key string, member string, expired ...int) error {
	err := cs.client.ZAdd(ctx, key, &redis.Z{
		Score:  0,
		Member: member,
	}).Err()
	return err
}

func (cs CallState) GetMembersScore(key string) (map[string]int64, error) {
	//parentSidRelationKey := fmt.Sprintf("parent:%s",callUuid)
	var resultState = make(map[string]int64)
	if result, err := cs.client.ZRangeWithScores(ctx, key, 0, -1).Result(); err == nil {
		for _, v := range result {
			resultState[v.Member.(string)] = int64(v.Score)
		}
		return resultState, err
	} else {
		return resultState, err
	}
}

func (cs CallState) IncrKeyMemberScore(key string, member string, score int) (int64, error) {
	val, err := cs.client.ZIncr(ctx, key, &redis.Z{
		Score:  float64(score),
		Member: member,
	}).Result()
	return int64(val), err
}

func (cs CallState) DelKeyMember(key string, member string) error {
	err := cs.client.ZRem(ctx, key, member).Err()
	return err
}
