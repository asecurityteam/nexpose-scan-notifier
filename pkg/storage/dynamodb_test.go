package storage

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamoDBTimestampStorage_FetchTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := NewMockDynamoDBAPI(ctrl)
	dynamoTimestampStorage := &DynamoDBTimestampStorage{
		db:                mockDB,
		tableName:         defaultDynamoDBTableName,
		partitionKeyName:  defaultDynamoDBPartitionKeyName,
		partitionKeyValue: defaultDynamoDBLastProcessedPartionKey,
		timestampKeyName:  defaultDynamoDBTimestampKeyName,
	}

	ts := time.Now()
	tests := []struct {
		name        string
		response    *dynamodb.GetItemOutput
		responseErr error
		expected    time.Time
		errExpected bool
	}{
		{
			name: "success",
			response: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					defaultDynamoDBPartitionKeyName: {
						S: aws.String(defaultDynamoDBLastProcessedPartionKey),
					},
					defaultDynamoDBTimestampKeyName: {
						S: aws.String(ts.Format(time.RFC3339Nano)),
					},
				},
			},
			responseErr: nil,
			expected:    ts,
			errExpected: false,
		},
		{
			name:        "error fetching timestamp",
			response:    &dynamodb.GetItemOutput{},
			responseErr: fmt.Errorf("get item error"),
			expected:    time.Time{},
			errExpected: true,
		},
		{
			name: "no timestamp found",
			response: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					defaultDynamoDBPartitionKeyName: {
						S: aws.String(defaultDynamoDBLastProcessedPartionKey),
					},
				},
			},
			responseErr: nil,
			expected:    time.Time{},
			errExpected: true,
		},
		{
			name: "cannot parse timestamp",
			response: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					defaultDynamoDBPartitionKeyName: {
						S: aws.String(defaultDynamoDBLastProcessedPartionKey),
					},
					defaultDynamoDBTimestampKeyName: {
						S: aws.String("2019-05-29 1PM"),
					},
				},
			},
			responseErr: nil,
			expected:    time.Time{},
			errExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB.EXPECT().GetItemWithContext(gomock.Any(), gomock.Any()).Return(tt.response, tt.responseErr)
			actual, err := dynamoTimestampStorage.FetchTimestamp(context.Background())
			require.Equal(t, tt.expected.Format(time.RFC3339Nano), actual.Format(time.RFC3339Nano))
			if tt.errExpected {
				require.Error(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}

func TestDynamoDBTimestampStorage_StoreTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := NewMockDynamoDBAPI(ctrl)
	dynamoTimestampStorage := &DynamoDBTimestampStorage{
		db:                mockDB,
		tableName:         defaultDynamoDBTableName,
		partitionKeyName:  defaultDynamoDBPartitionKeyName,
		partitionKeyValue: defaultDynamoDBLastProcessedPartionKey,
		timestampKeyName:  defaultDynamoDBTimestampKeyName,
	}

	ts := time.Now()
	putItemInput := &dynamodb.PutItemInput{
		TableName: aws.String(defaultDynamoDBTableName),
		Item: map[string]*dynamodb.AttributeValue{
			defaultDynamoDBPartitionKeyName: {
				S: aws.String(defaultDynamoDBLastProcessedPartionKey),
			},
			defaultDynamoDBTimestampKeyName: {
				S: aws.String(ts.Format(time.RFC3339Nano)),
			},
		},
	}

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "success",
			err:  nil,
		},
		{
			name: "error storing timestamp",
			err:  fmt.Errorf("dynamodb error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB.EXPECT().PutItemWithContext(gomock.Any(), putItemInput).Return(&dynamodb.PutItemOutput{}, tt.err)
			actual := dynamoTimestampStorage.StoreTimestamp(context.Background(), ts)
			if tt.err != nil {
				require.Error(t, actual)
				return
			}
			require.Nil(t, actual)
		})
	}
}

func TestDynamoDBDependencyCheck(t *testing.T) {
	tests := []struct {
		name          string
		returnedError error
		expectedErr   bool
	}{
		{
			name:          "success",
			returnedError: nil,
			expectedErr:   false,
		},
		{
			name:          "failure",
			returnedError: errors.New("🐖"),
			expectedErr:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			ctrl := gomock.NewController(tt)
			mockDB := NewMockDynamoDBAPI(ctrl)
			mockDB.EXPECT().DescribeTable(gomock.Any()).Return(nil, test.returnedError)

			dynamoTimestampStorage := &DynamoDBTimestampStorage{
				db:                mockDB,
				tableName:         defaultDynamoDBTableName,
				partitionKeyName:  defaultDynamoDBPartitionKeyName,
				partitionKeyValue: defaultDynamoDBLastProcessedPartionKey,
				timestampKeyName:  defaultDynamoDBTimestampKeyName,
			}
			err := dynamoTimestampStorage.CheckDependencies(context.Background())
			assert.Equal(tt, test.expectedErr, err != nil)
		})
	}
}
