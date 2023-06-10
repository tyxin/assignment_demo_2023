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

	if err := client.Ping(ctx).Err(); err != nil {
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
		Score:  float64(message.Timestamp),
		Member: text,
	}

	_, err = c.cli.ZAdd(ctx, roomID, *member).Result()

	if err != nil {
		return err
	}

	return nil
}

func (c *RedisClient) GetMessagesByRoomID(ctx context.Context, roomID string, start, end int64, reverse bool) ([]*Message, error) {

	var (
		rawText  []string
		messages []*Message
		err      error
	)

	if reverse {
		rawText, err = c.cli.ZRevRange(ctx, roomID, start, end).Result()

		if err != nil {
			return nil, err
		}
	} else {
		rawText, err = c.cli.ZRevRange(ctx, roomID, start, end).Result()

		if err != nil {
			return nil, err
		}
	}

	for _, msg := range rawText {
		temp := &Message{}
		err := json.Unmarshal([]byte(msg), temp)

		if err != nil {
			return nil, err
		}

		messages = append(messages, temp)
	}

	return messages, nil
}
