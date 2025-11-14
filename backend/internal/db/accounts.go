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

func (db *DynamoDB) CreateAccount(ctx context.Context, account *models.Account) error {
	now := time.Now().Unix()
	account.CreatedAt = now
	account.UpdatedAt = now

	item, err := attributevalue.MarshalMap(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.AccountsTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (db *DynamoDB) GetAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	result, err := db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.AccountsTable),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	var account models.Account
	if err := attributevalue.UnmarshalMap(result.Item, &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &account, nil
}

func (db *DynamoDB) GetAccountByID(ctx context.Context, accountID string) (*models.Account, error) {
	result, err := db.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(db.AccountsTable),
		IndexName:              aws.String("account_id-index"),
		KeyConditionExpression: aws.String("account_id = :account_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":account_id": &types.AttributeValueMemberS{Value: accountID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query account: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var account models.Account
	if err := attributevalue.UnmarshalMap(result.Items[0], &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &account, nil
}

func (db *DynamoDB) UpdateAccount(ctx context.Context, account *models.Account) error {
	account.UpdatedAt = time.Now().Unix()

	item, err := attributevalue.MarshalMap(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	_, err = db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.AccountsTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}
