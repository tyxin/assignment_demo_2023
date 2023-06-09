package main

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	cli *redis.Client
}

func (c *RedisClient) InitClient(ctx context.Context, address, password string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	if err := r.Ping(ctx).Err(); err != nil {
		return err
	}

	c.cli = client

	return nil
}

func (c *RedisClient) SaveMessage(ctx context.Context, roomID string, message *Message) error {

	text, err := json.Marshal(message)

	if err != nil {
		return err
	}

	member := &redis.Z{
		Score:  message.Timestamp,
		Member: text,
	}

	_, err = c.cli.ZAdd(ctx, roomID, *member).Result()

	if err != nil {
		return err
	}

	return nil
}
