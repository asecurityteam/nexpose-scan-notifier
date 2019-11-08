package v1

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDepCheckHandleSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	DynamoDBMockDependencyChecker := NewMockDependencyChecker(ctrl)
	DynamoDBMockDependencyChecker.EXPECT().CheckDependencies(context.Background()).Return(nil)
	NexposeClientMockDependencyChecker := NewMockDependencyChecker(ctrl)
	NexposeClientMockDependencyChecker.EXPECT().CheckDependencies(context.Background()).Return(nil)

	handler := &DependencyCheckHandler{
		DynamoDBDependencyChecker:      DynamoDBMockDependencyChecker,
		NexposeClientDependencyChecker: NexposeClientMockDependencyChecker,
	}
	err := handler.Handle(context.Background())

	assert.Nil(t, err)
}

func TestDepCheckHandleDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	DynamoDBMockDependencyChecker := NewMockDependencyChecker(ctrl)
	DynamoDBMockDependencyChecker.EXPECT().CheckDependencies(context.Background()).Return(fmt.Errorf("error"))
	NexposeClientMockDependencyChecker := NewMockDependencyChecker(ctrl)

	handler := &DependencyCheckHandler{
		DynamoDBDependencyChecker:      DynamoDBMockDependencyChecker,
		NexposeClientDependencyChecker: NexposeClientMockDependencyChecker,
	}
	err := handler.Handle(context.Background())

	assert.NotNil(t, err)
}

func TestDepCheckHandleNexposeClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	DynamoDBMockDependencyChecker := NewMockDependencyChecker(ctrl)
	DynamoDBMockDependencyChecker.EXPECT().CheckDependencies(context.Background()).Return(nil)
	NexposeClientMockDependencyChecker := NewMockDependencyChecker(ctrl)
	NexposeClientMockDependencyChecker.EXPECT().CheckDependencies(context.Background()).Return(fmt.Errorf("error"))

	handler := &DependencyCheckHandler{
		DynamoDBDependencyChecker:      DynamoDBMockDependencyChecker,
		NexposeClientDependencyChecker: NexposeClientMockDependencyChecker,
	}
	err := handler.Handle(context.Background())

	assert.NotNil(t, err)
}
