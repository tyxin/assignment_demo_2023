package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {

	if err := validateSendRequest(req); err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	message := &Message{
		Message:   req.Message.GetText(),
		Sender:    req.Message.GetSender(),
		Timestamp: timestamp,
	}

	roomID, err := getRoomID(req.Message.GetChat())

	if err != nil {
		return nil, ctx.Err()
	}

	err = rdb.SaveMessage(ctx, roomID, message)

	if err != nil {
		return nil, ctx.Err()
	}

	resp := rpc.NewSendResponse()
	resp.Code, resp.Msg = 0, "success"
	return resp, nil

}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {

	roomID, err := getRoomID(req.GetChat())

	if err != nil {
		return nil, err
	}

	limit := int64(req.GetLimit())
	if limit == 0 {
		limit = 10
	}
	start := req.GetCursor()

	end := start + limit

	messages, err := rdb.GetMessagesByRoomID(ctx, roomID, start, end, req.GetReverse())

	if err != nil {
		return nil, err
	}

	respMessages := make([]*rpc.Message, 0)
	var count int64 = 0
	var nextCursor int64 = 0
	var hasMore bool = false

	for _, msg := range messages {
		if count+1 > limit {
			hasMore = true
			nextCursor = end
			break
		}
		temp := &rpc.Message{
			Chat:     req.GetChat(),
			Text:     msg.Message,
			Sender:   msg.Sender,
			SendTime: msg.Timestamp,
		}
		respMessages = append(respMessages, temp)
		count += 1
	}

	resp := rpc.NewPullResponse()
	resp.Messages = respMessages
	resp.Code, resp.Msg = 0, "success"
	resp.HasMore = &hasMore
	resp.NextCursor = &nextCursor

	return resp, nil
}

func validateSendRequest(req *rpc.SendRequest) error {

	senders := strings.Split(req.Message.Chat, ":")

	if len(senders) != 2 {
		return fmt.Errorf("invalid Chat ID '%s', should be in format (user1:user2)", req.Message.GetChat())
	}

	sender1, sender2 := senders[0], senders[1]

	if (req.Message.GetSender() != sender1) && (req.Message.GetSender() != sender2) {
		return fmt.Errorf("sender '%s' not in room", req.Message.GetSender())
	}

	return nil

}

func getRoomID(chat string) (string, error) {

	var roomID string
	senders := strings.Split(chat, ":")

	if len(senders) != 2 {
		return "", fmt.Errorf("invalid Chat ID '%s', should be in the form (user1:user2)", chat)
	}

	sender1, sender2 := senders[0], senders[1]

	if comp := strings.Compare(sender1, sender2); comp == 1 {
		roomID = fmt.Sprintf("%s:%s", sender2, sender1)
	} else {
		roomID = fmt.Sprintf("%s:%s", sender1, sender2)
	}

	return roomID, nil
}
