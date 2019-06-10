// +build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/storage"
	"github.com/asecurityteam/settings"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/require"
)

const (
	dynamoTableName        = "ScanTimestamp"
	dynamoPartitionKeyName = "partitionkey"
)

var dynamoDBTimestampStorage *storage.DynamoDBTimestampStorage

// createTable creates a DynamoDB table
func createTable(ctx context.Context, db *dynamodb.DynamoDB) {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(dynamoPartitionKeyName),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(dynamoPartitionKeyName),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(dynamoTableName),
	}
	if _, e := db.CreateTableWithContext(ctx, input); e != nil {
		panic(fmt.Sprintf("create table err: %s", e.Error()))
	}
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	// create aws session and dynamodb connection
	awsConfig := aws.NewConfig()
	awsConfig.Region = aws.String(os.Getenv("DYNAMODB_REGION"))
	awsConfig.Endpoint = aws.String(os.Getenv("DYNAMODB_ENDPOINT"))
	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		panic(err.Error())
	}
	db := dynamodb.New(awsSession)

	// create ScanTimestamp table
	_, err = db.DescribeTableWithContext(ctx,
		&dynamodb.DescribeTableInput{TableName: aws.String(dynamoTableName)})
	switch err.(type) {
	case nil: // delete table and recreate if found
		if _, e := db.DeleteTableWithContext(ctx, &dynamodb.DeleteTableInput{TableName: aws.String(dynamoTableName)}); e != nil {
			panic(fmt.Sprintf("delete table err: %s", e.Error()))
		}
		createTable(ctx, db)
	case awserr.Error: // create table if not found
		if err.(awserr.Error).Code() == dynamodb.ErrCodeResourceNotFoundException {
			createTable(ctx, db)
		}
	default:
		panic(err.Error())
	}

	// create DynamoDB timestamp fetcher/storer
	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}
	dynamoDBComponent := &storage.DynamoDBTimestampStorageComponent{}
	dynamoDBTimestampStorage = new(storage.DynamoDBTimestampStorage)
	if err = settings.NewComponent(ctx, source, dynamoDBComponent, dynamoDBTimestampStorage); err != nil {
		panic(err.Error())
	}

	// run test
	os.Exit(m.Run())
}

func TestDynamoDBTimestampStore_StoreAndRetrieveTimestamp(t *testing.T) {
	// store a timestamp in the dynamodb table
	now := time.Now()
	e := dynamoDBTimestampStorage.StoreTimestamp(context.Background(), now)
	require.Nil(t, e)

	// fetch the timestamp
	ts, e := dynamoDBTimestampStorage.FetchTimestamp(context.Background())
	require.Equal(t, ts.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	require.Nil(t, e)

	// persist a new timestamp
	e = dynamoDBTimestampStorage.StoreTimestamp(context.Background(), now.Add(1*time.Hour))
	require.Nil(t, e)

	// fetch the new timestamp
	ts, e = dynamoDBTimestampStorage.FetchTimestamp(context.Background())
	require.Equal(t, ts.Format(time.RFC3339Nano), now.Add(1*time.Hour).Format(time.RFC3339Nano))
	require.Nil(t, e)
}
