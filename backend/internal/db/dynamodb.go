package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDB struct {
	Client          *dynamodb.Client
	AccountsTable   string
	PastesTable     string
	RateLimitsTable string
}

func NewDynamoDB(ctx context.Context, region, accountsTable, pastesTable, rateLimitsTable string) (*DynamoDB, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDB{
		Client:          client,
		AccountsTable:   accountsTable,
		PastesTable:     pastesTable,
		RateLimitsTable: rateLimitsTable,
	}, nil
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
