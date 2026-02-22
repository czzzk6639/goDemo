package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"game-server/pkg/redis"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

const (
	SessionKeyPrefix = "session:"
	SessionTTL       = 24 * time.Hour
)

type Session struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Token     string `json:"token"`
	CreatedAt int64  `json:"created_at"`
}

type SessionService struct{}

func NewSessionService() *SessionService {
	return &SessionService{}
}

func (s *SessionService) Create(userID int64, username, token string) error {
	sess := &Session{
		UserID:    userID,
		Username:  username,
		Token:     token,
		CreatedAt: time.Now().Unix(),
	}

	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	ctx := context.Background()
	key := SessionKeyPrefix + token

	return redis.Client.Set(ctx, key, data, SessionTTL).Err()
}

func (s *SessionService) Get(token string) (*Session, error) {
	ctx := context.Background()
	key := SessionKeyPrefix + token

	data, err := redis.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, ErrSessionNotFound
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}

	return &sess, nil
}

func (s *SessionService) Validate(token string) (*Session, error) {
	sess, err := s.Get(token)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (s *SessionService) Delete(token string) error {
	ctx := context.Background()
	key := SessionKeyPrefix + token
	return redis.Client.Del(ctx, key).Err()
}

func (s *SessionService) Refresh(token string) error {
	ctx := context.Background()
	key := SessionKeyPrefix + token
	return redis.Client.Expire(ctx, key, SessionTTL).Err()
}

func (s *SessionService) SetUserOnline(userID int64, token string) error {
	ctx := context.Background()
	key := fmt.Sprintf("online:%d", userID)
	return redis.Client.Set(ctx, key, token, SessionTTL).Err()
}

func (s *SessionService) SetUserOffline(userID int64) error {
	ctx := context.Background()
	key := fmt.Sprintf("online:%d", userID)
	return redis.Client.Del(ctx, key).Err()
}

func (s *SessionService) IsUserOnline(userID int64) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("online:%d", userID)
	n, err := redis.Client.Exists(ctx, key).Result()
	return n > 0, err
}
