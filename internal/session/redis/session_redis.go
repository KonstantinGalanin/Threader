package repository

import (
	"encoding/json"
	"fmt"

	"github.com/KonstantinGalanin/redditclone/internal/myerrors"
	"github.com/KonstantinGalanin/redditclone/internal/session"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)


type SessionManagerRedis struct {
	redisConn redis.Conn
}

func NewSessionManagerRedis(conn redis.Conn) *SessionManagerRedis {
	return &SessionManagerRedis{
		redisConn: conn,
	}
}

func (sm *SessionManagerRedis) Create(in *session.Session) (*session.SessionID, error) {
	id := session.SessionID{
		ID: uuid.New().String(),
	}
	dataSerialized, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("create session %w", err)
	}
	mkey := "session:" + id.ID
	result, err := redis.String(sm.redisConn.Do("SET", mkey, dataSerialized, "EX", 86400))
	if err != nil {
		return nil, fmt.Errorf("create session %w", err)
	}
	if result != "OK" {
		return nil, myerrors.ErrRedisSetNotOk
	}

	return &id, nil
}

func (sm *SessionManagerRedis) Check(in *session.SessionID) (*session.Session, error) {
	mkey := "session:" + in.ID
	data, err := redis.Bytes(sm.redisConn.Do("GET", mkey))
	if err != nil {
		return nil, fmt.Errorf("cant unpack session data: %w", err)
	}
	sess := &session.Session{}
	err = json.Unmarshal(data, sess)
	if err != nil {
		return nil, fmt.Errorf("unmarshal session key: %w", err)
	}

	return sess, nil
}

func (sm *SessionManagerRedis) Delete(in *session.SessionID) error {
	mkey := "session:" + in.ID
	_, err := redis.Int(sm.redisConn.Do("DEL", mkey))
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}
