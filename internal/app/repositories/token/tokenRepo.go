package repositories

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
)

type TokenRepo struct {
	redis *redis.Client
}

func NewTokenRepo(redisClient *redis.Client) *TokenRepo {
	return &TokenRepo{
		redis: redisClient,
	}
}

func (t *TokenRepo) StoreToken(userId string, tokenUuid string, expired time.Time) error {
	err := t.redis.Set(tokenUuid, userId, expired.Sub(time.Now()))
	if err.Err() != nil {
		return err.Err()
	}

	return nil
}

func (t *TokenRepo) FetchUser(tokenUuid string) (string, error) {
	userId, err := t.redis.Get(tokenUuid).Result()
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (t *TokenRepo) DeleteToken(tokenUuid string) (int64, error) {
	deleted, err := t.redis.Del(tokenUuid).Result()
	if err != nil {
		return 0, err
	}

	return deleted, nil
}

func (t *TokenRepo) DeleteAllToken(userId string) error {
	length, err := t.redis.SCard(userId).Result()
	if err != nil {
		return err
	}

	for i := 0; i < int(length); i++ {
		if err := t.redis.SPop(userId).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (t *TokenRepo) AddToken(userId string, tokenString string) error {
	added, err := t.redis.SAdd(userId, tokenString).Result()
	if err != nil {
		return err
	}

	if added == 0 {
		return fmt.Errorf("Add token error")
	}

	return nil
}

func (t *TokenRepo) IsForgotTokenOf(forgotToken string, userId string) (bool, error) {
	return t.redis.SIsMember(userId, forgotToken).Result()
}
