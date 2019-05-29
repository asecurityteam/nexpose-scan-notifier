package storage

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	defaultDynamoDBTableName               = "ScanTimestamp"
	defaultDynamoDBPartitionKeyName        = "partitionkey"
	defaultDynamoDBLastProcessedPartionKey = "lastProcessed"
	defaultDynamoDBTimestampKeyName        = "timestamp"
)

// DynamoDBTimestampStorageConfig holds configuration required to send Nexpose assets
// to a queue via an HTTP Producer
type DynamoDBTimestampStorageConfig struct {
	TableName         string
	PartitionKeyName  string
	PartitionKeyValue string
	TimestampKeyName  string
	Region            string
	Endpoint          string
}

// Name is used by the settings library and will add a "DYNAMODB"
// prefix to DynamoDBTimestampStorageConfig environment variables
func (c *DynamoDBTimestampStorageConfig) Name() string {
	return "DynamoDB"
}

// DynamoDBTimestampStorageComponent satisfies the settings library Component
// API, and may be used by the settings.NewComponent function.
type DynamoDBTimestampStorageComponent struct{}

// Settings can be used to populate default values if there are any
func (*DynamoDBTimestampStorageComponent) Settings() *DynamoDBTimestampStorageConfig {
	return &DynamoDBTimestampStorageConfig{
		TableName:         defaultDynamoDBTableName,
		PartitionKeyName:  defaultDynamoDBPartitionKeyName,
		PartitionKeyValue: defaultDynamoDBLastProcessedPartionKey,
		TimestampKeyName:  defaultDynamoDBTimestampKeyName,
	}
}

// New constructs a DynamoDBTimestampStorage from a config.
func (*DynamoDBTimestampStorageComponent) New(_ context.Context, c *DynamoDBTimestampStorageConfig) (*DynamoDBTimestampStorage, error) {
	awsConfig := aws.NewConfig()
	awsConfig.Region = aws.String(c.Region)
	awsConfig.Endpoint = aws.String(c.Endpoint)
	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	db := dynamodb.New(awsSession)
	return &DynamoDBTimestampStorage{
		db:                db,
		tableName:         c.TableName,
		partitionKeyName:  c.PartitionKeyName,
		partitionKeyValue: c.PartitionKeyValue,
		timestampKeyName:  c.TimestampKeyName,
	}, nil
}
