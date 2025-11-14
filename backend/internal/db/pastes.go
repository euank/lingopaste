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

func (db *DynamoDB) CreatePasteMeta(ctx context.Context, meta *models.PasteMeta) error {
	meta.CreatedAt = time.Now().Unix()

	item, err := attributevalue.MarshalMap(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal paste meta: %w", err)
	}

	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.PastesTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create paste meta: %w", err)
	}

	return nil
}

func (db *DynamoDB) GetPasteMeta(ctx context.Context, pasteID string) (*models.PasteMeta, error) {
	result, err := db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.PastesTable),
		Key: map[string]types.AttributeValue{
			"paste_id": &types.AttributeValueMemberS{Value: pasteID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get paste meta: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var meta models.PasteMeta
	if err := attributevalue.UnmarshalMap(result.Item, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal paste meta: %w", err)
	}

	return &meta, nil
}

func (db *DynamoDB) UpdatePasteMeta(ctx context.Context, meta *models.PasteMeta) error {
	item, err := attributevalue.MarshalMap(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal paste meta: %w", err)
	}

	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.PastesTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update paste meta: %w", err)
	}

	return nil
}

func (db *DynamoDB) AddTranslationLanguage(ctx context.Context, pasteID, language string) error {
	// First, get the current paste to check if language already exists
	meta, err := db.GetPasteMeta(ctx, pasteID)
	if err != nil {
		return fmt.Errorf("failed to get paste meta: %w", err)
	}
	if meta == nil {
		return fmt.Errorf("paste not found")
	}

	// Check if language already in list
	for _, lang := range meta.AvailableTranslations {
		if lang == language {
			return nil // Already exists, nothing to do
		}
	}

	// Append to list
	_, err = db.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(db.PastesTable),
		Key: map[string]types.AttributeValue{
			"paste_id": &types.AttributeValueMemberS{Value: pasteID},
		},
		UpdateExpression: aws.String("SET available_translations = list_append(if_not_exists(available_translations, :empty_list), :lang)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lang":       &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: language}}},
			":empty_list": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add translation language: %w", err)
	}

	return nil
}
