package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lingopaste/backend/internal/models"
)

func (db *DynamoDB) GetRateLimit(ctx context.Context, identifier, date string) (*models.RateLimit, error) {
	result, err := db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.RateLimitsTable),
		Key: map[string]types.AttributeValue{
			"identifier": &types.AttributeValueMemberS{Value: identifier},
			"date":       &types.AttributeValueMemberS{Value: date},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var limit models.RateLimit
	if err := attributevalue.UnmarshalMap(result.Item, &limit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limit: %w", err)
	}

	return &limit, nil
}

func (db *DynamoDB) IncrementRateLimit(ctx context.Context, identifier, date, limitType string) (int, error) {
	// TTL is 48 hours from now (to allow for timezone differences)
	ttl := time.Now().Add(48 * time.Hour).Unix()

	result, err := db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(db.RateLimitsTable),
		Key: map[string]types.AttributeValue{
			"identifier": &types.AttributeValueMemberS{Value: identifier},
			"date":       &types.AttributeValueMemberS{Value: date},
		},
		UpdateExpression: aws.String("ADD paste_count :inc SET limit_type = :limit_type, #ttl = :ttl"),
		ExpressionAttributeNames: map[string]string{
			"#ttl": "ttl",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc":        &types.AttributeValueMemberN{Value: "1"},
			":limit_type": &types.AttributeValueMemberS{Value: limitType},
			":ttl":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	var limit models.RateLimit
	if err := attributevalue.UnmarshalMap(result.Attributes, &limit); err != nil {
		return 0, fmt.Errorf("failed to unmarshal rate limit: %w", err)
	}

	return limit.PasteCount, nil
}
