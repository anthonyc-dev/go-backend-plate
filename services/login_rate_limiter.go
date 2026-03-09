package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rest-api/configs"

	"github.com/redis/go-redis/v9"
)

var (
	ErrAccountLocked = errors.New("account temporarily locked due to too many failed login attempts")
)

type LoginAttemptLimiter struct {
	redis        *redis.Client
	maxAttempts  int
	blockMinutes int
}

func NewLoginAttemptLimiter() *LoginAttemptLimiter {
	maxAttempts := 5
	blockMinutes := 5

	if configs.AppEnv != nil {
		maxAttempts = configs.AppEnv.MaxLoginAttempts
		blockMinutes = configs.AppEnv.LoginBlockMinutes
	}

	return &LoginAttemptLimiter{
		redis:        configs.RedisClient,
		maxAttempts:  maxAttempts,
		blockMinutes: blockMinutes,
	}
}

func (l *LoginAttemptLimiter) ctx() context.Context {
	return context.Background()
}

func (l *LoginAttemptLimiter) attemptsKey(email string) string {
	return fmt.Sprintf("login:attempts:%s", email)
}

func (l *LoginAttemptLimiter) blockKey(email string) string {
	return fmt.Sprintf("login:blocked:%s", email)
}

func (l *LoginAttemptLimiter) CheckBlocked(email string) error {
	blocked, err := l.redis.Exists(l.ctx(), l.blockKey(email)).Result()
	if err != nil {
		return err
	}

	if blocked > 0 {
		ttl, _ := l.redis.TTL(l.ctx(), l.blockKey(email)).Result()
		if ttl > 0 {
			return fmt.Errorf("%w. Try again in %d minutes", ErrAccountLocked, int(ttl.Minutes()))
		}
		l.redis.Del(l.ctx(), l.blockKey(email))
	}

	return nil
}

func (l *LoginAttemptLimiter) RecordFailedAttempt(email string) (int, error) {
	key := l.attemptsKey(email)

	pipe := l.redis.Pipeline()
	incr := pipe.Incr(l.ctx(), key)
	pipe.Expire(l.ctx(), key, l.lockoutDuration())

	_, err := pipe.Exec(l.ctx())
	if err != nil {
		return 0, err
	}

	attempts := int(incr.Val())

	if attempts >= l.maxAttempts {
		blockKey := l.blockKey(email)
		l.redis.Set(l.ctx(), blockKey, "1", l.blockDuration())
		l.redis.Del(l.ctx(), key)
		return attempts, ErrAccountLocked
	}

	return attempts, nil
}

func (l *LoginAttemptLimiter) ResetAttempts(email string) error {
	l.redis.Del(l.ctx(), l.attemptsKey(email))
	l.redis.Del(l.ctx(), l.blockKey(email))
	return nil
}

func (l *LoginAttemptLimiter) GetRemainingAttempts(email string) int {
	count, err := l.redis.Get(l.ctx(), l.attemptsKey(email)).Int()
	if err != nil {
		return l.maxAttempts
	}
	return l.maxAttempts - count
}

func (l *LoginAttemptLimiter) lockoutDuration() time.Duration {
	return time.Duration(l.blockMinutes) * time.Minute * 2
}

func (l *LoginAttemptLimiter) blockDuration() time.Duration {
	return time.Duration(l.blockMinutes) * time.Minute
}

var loginLimiter *LoginAttemptLimiter

func GetLoginLimiter() *LoginAttemptLimiter {
	if loginLimiter == nil {
		loginLimiter = NewLoginAttemptLimiter()
	}
	return loginLimiter
}
