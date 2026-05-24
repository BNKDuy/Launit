package db

import (
	"context"
	"fmt"
	"orchestrator/internal/domain"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoTokenRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoTokenRepository(client *dynamodb.Client, tableName string) *DynamoTokenRepository {
	return &DynamoTokenRepository{
		client:    client,
		tableName: tableName,
	}
}

var _ TokenRepository = (*DynamoTokenRepository)(nil)

func (r *DynamoTokenRepository) getKey(tokenValue string) string {
	return "TOKEN#" + tokenValue
}

func (r *DynamoTokenRepository) SaveToken(ctx context.Context, token domain.Token) error {
	key := r.getKey(token.Value)

	unixString := strconv.FormatInt(token.CreatedAt.Unix(), 10)

	item := map[string]types.AttributeValue{
		"pk":         &types.AttributeValueMemberS{Value: key},
		"username":   &types.AttributeValueMemberS{Value: token.Username},
		"created_at": &types.AttributeValueMemberS{Value: unixString},
		"valid":      &types.AttributeValueMemberBOOL{Value: token.Valid},
	}

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})

	return err
}

func (r *DynamoTokenRepository) FindToken(ctx context.Context, tokenValue string) (domain.Token, error) {
	key := map[string]types.AttributeValue{
		"pk": &types.AttributeValueMemberS{Value: r.getKey(tokenValue)},
	}

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	})
	if err != nil {
		return domain.Token{}, err
	}
	if out.Item == nil {
		return domain.Token{}, fmt.Errorf("token not found")
	}

	var username string
	if val, ok := out.Item["username"].(*types.AttributeValueMemberS); ok {
		username = val.Value
	}

	var createdAt time.Time
	if val, ok := out.Item["created_at"].(*types.AttributeValueMemberS); ok {
		if seconds, parseErr := strconv.ParseInt(val.Value, 10, 64); parseErr == nil {
			createdAt = time.Unix(seconds, 0)
		}
	}

	var valid bool
	if val, ok := out.Item["valid"].(*types.AttributeValueMemberBOOL); ok {
		valid = val.Value
	}

	return domain.Token{
		Value:     tokenValue,
		Username:  username,
		CreatedAt: createdAt,
		Valid:     valid,
	}, nil
}
