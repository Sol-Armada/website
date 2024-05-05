package main

import (
	"context"
	"log/slog"

	"github.com/sol-armada/sol-bot/rsi"
)

var rsiActions = map[string]Action{
	"check_handle": checkHandle,
}

func checkHandle(ctx context.Context, c *Client, arg any) CommandResponse {
	logger := slog.Default()

	cr := CommandResponse{
		Thing:  "rsi",
		Action: "check_handle",
	}

	handle, ok := arg.(string)
	if !ok {
		return CommandResponse{
			Thing:  "rsi",
			Action: "check_handle",
			Error:  "bad_handle",
		}
	}

	logger.Info("checking handle", "handle", handle)

	cr.Result = rsi.ValidHandle(handle)

	logger.Info("checked handle", "handle", handle, "valid", cr.Result)

	return cr
}
