package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	dynamoDBConfig := DynamoDBTimestampStorageConfig{}
	require.Equal(t, "DynamoDB", dynamoDBConfig.Name())
}

func TestComponentDefaultConfig(t *testing.T) {
	component := &DynamoDBTimestampStorageComponent{}
	config := component.Settings()
	require.Empty(t, config.Endpoint)
	require.Empty(t, config.Region)
	require.Equal(t, config.TableName, defaultDynamoDBTableName)
	require.Equal(t, config.PartitionKeyName, defaultDynamoDBPartitionKeyName)
	require.Equal(t, config.PartitionKeyValue, defaultDynamoDBLastProcessedPartionKey)
	require.Equal(t, config.TimestampKeyName, defaultDynamoDBTimestampKeyName)
}

func TestNexposeClientConfigWithValues(t *testing.T) {
	component := &DynamoDBTimestampStorageComponent{}
	config := &DynamoDBTimestampStorageConfig{
		Endpoint:          "http://localhost:8000",
		Region:            "us-west-2",
		TableName:         "tableName",
		PartitionKeyName:  "partitionKeyName",
		PartitionKeyValue: "partitionKeyValue",
		TimestampKeyName:  "timestampKeyName",
	}
	dynamoDBTimestampStorage, err := component.New(context.Background(), config)

	require.Equal(t, "tableName", dynamoDBTimestampStorage.tableName)
	require.Equal(t, "partitionKeyName", dynamoDBTimestampStorage.partitionKeyName)
	require.Equal(t, "partitionKeyValue", dynamoDBTimestampStorage.partitionKeyValue)
	require.Equal(t, "timestampKeyName", dynamoDBTimestampStorage.timestampKeyName)
	require.Nil(t, err)
}
