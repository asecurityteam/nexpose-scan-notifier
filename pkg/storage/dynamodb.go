package storage

import (
	"context"
	"time"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type dynamoResult struct {
	PartitionKey string `json:"partitionKey"`
	Timestamp    string `json:"timestamp"`
}

// DynamoDBTimestampStorage provides persistence and retrieval of last processed scan timestamps from
// a DynamoDB table.
type DynamoDBTimestampStorage struct {
	db                dynamodbiface.DynamoDBAPI
	tableName         string
	partitionKeyName  string
	partitionKeyValue string
	timestampKeyName  string
}

// FetchTimestamp queries a DynamoDB table with a static partition key for the last processed timestamp.
func (s *DynamoDBTimestampStorage) FetchTimestamp(ctx context.Context) (time.Time, error) {
	item, err := s.db.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			s.partitionKeyName: {
				S: aws.String(s.partitionKeyValue),
			},
		},
	})
	if err != nil {
		return time.Time{}, err
	}

	var result dynamoResult
	err = dynamodbattribute.UnmarshalMap(item.Item, &result)
	if err != nil {
		return time.Time{}, err
	}

	if result.Timestamp == "" {
		return time.Time{}, domain.TimestampNotFound{}
	}

	ts, err := time.Parse(time.RFC3339Nano, result.Timestamp)
	if err != nil {
		return time.Time{}, err
	}

	return ts, nil
}

// StoreTimestamp upserts a timestamp to a DynamoDB table with a static partition key.
func (s *DynamoDBTimestampStorage) StoreTimestamp(ctx context.Context, ts time.Time) error {
	_, err := s.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item: map[string]*dynamodb.AttributeValue{
			s.partitionKeyName: {
				S: aws.String(s.partitionKeyValue),
			},
			s.timestampKeyName: {
				S: aws.String(ts.Format(time.RFC3339Nano)),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// CheckDependencies tries to communicate to the DB by trying to retrieve its tables
func (s *DynamoDBTimestampStorage) CheckDependencies(ctx context.Context) error {
	_, err := s.db.ListTables(&dynamodb.ListTablesInput{})
	return err
}
