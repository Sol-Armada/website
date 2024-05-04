package main

import (
	"context"
)

var contractsActions = map[string]Action{
	"list":   getContracts,
	"get":    getContract,
	"create": createContract,
	"update": updateContract,
	"delete": deleteContract,
}

func getContracts(ctx context.Context, c *Client, token any) CommandResponse {
	// logger := slog.Default()
	return CommandResponse{}
}

func getContract(ctx context.Context, c *Client, token any) CommandResponse {
	// logger := slog.Default()
	return CommandResponse{}
}

func createContract(ctx context.Context, c *Client, token any) CommandResponse {
	// logger := slog.Default()
	return CommandResponse{}
}

func updateContract(ctx context.Context, c *Client, token any) CommandResponse {
	// logger := slog.Default()
	return CommandResponse{}
}

func deleteContract(ctx context.Context, c *Client, token any) CommandResponse {
	// logger := slog.Default()
	return CommandResponse{}
}
