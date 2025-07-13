package repository

import (
	"context"
	"testing"

	"github.com/anfern777/go-serverless-framework-api/internal/models"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type MockDatabaseAPI struct {
	TransactWriteItemsFunc       func(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	TransactWriteItemsFuncCalled bool

	DeleteItemFunc       func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	DeleteItemFuncCalled bool

	QueryFunc       func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	QueryFuncCalled bool

	BatchWriteItemFunc       func(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	BatchWriteItemFuncCalled bool

	UpdateItemFunc       func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	UpdateItemFuncCalled bool

	GetItemFunc       func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	GetItemFuncCalled bool

	PutItemFunc       func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	PutItemFuncCalled bool
}

func (m *MockDatabaseAPI) TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error) {
	m.TransactWriteItemsFuncCalled = true
	if m.TransactWriteItemsFunc != nil {
		return m.TransactWriteItemsFunc(ctx, params, optFns...)
	}
	return &dynamodb.TransactWriteItemsOutput{}, nil
}

func (m *MockDatabaseAPI) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	m.DeleteItemFuncCalled = true
	if m.DeleteItemFunc != nil {
		return m.DeleteItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.DeleteItemOutput{}, nil
}

func (m *MockDatabaseAPI) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	m.QueryFuncCalled = true
	if m.QueryFunc != nil {
		m.QueryFunc(ctx, params, optFns...)
	}
	return &dynamodb.QueryOutput{}, nil
}

func (m *MockDatabaseAPI) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	m.BatchWriteItemFuncCalled = true
	if m.BatchWriteItemFunc != nil {
		return m.BatchWriteItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.BatchWriteItemOutput{}, nil
}

func (m *MockDatabaseAPI) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	m.UpdateItemFuncCalled = true
	if m.UpdateItemFunc != nil {
		return m.UpdateItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.UpdateItemOutput{}, nil
}

func (m *MockDatabaseAPI) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	m.GetItemFuncCalled = true
	if m.GetItemFunc != nil {
		return m.GetItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.GetItemOutput{}, nil
}

func (m *MockDatabaseAPI) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	m.GetItemFuncCalled = true
	if m.GetItemFunc != nil {
		return m.PutItemFunc(ctx, params, optFns...)
	}
	return &dynamodb.PutItemOutput{}, nil
}

func TestChildRepositoryImpl_Save(t *testing.T) {
	documentTests := []struct {
		name               string
		item               models.Document
		expectedFuncCalled bool
	}{
		{
			name: "happy path",
			item: models.Document{},

			expectedFuncCalled: true,
		},
	}

	mockDbApi := &MockDatabaseAPI{
		TransactWriteItemsFuncCalled: false,
		TransactWriteItemsFunc:       nil,
	}

	mock := &ChildRepositoryImpl[models.Document]{
		databaseAPI: mockDbApi,
		table:       "test-table",
		index:       nil,
	}

	for _, tt := range documentTests {
		t.Run(tt.name, func(t *testing.T) {
			err := mock.Save(context.TODO(), tt.item)
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if err == nil {
				if tt.expectedFuncCalled != mockDbApi.TransactWriteItemsFuncCalled {
					t.Errorf("expected TransactWriteItems to be called, but got no call")
				}
			}
		})
	}
}

func GetAllByParentPKRepositoryImpl_Delete(t *testing.T) {
	documentTests := []struct {
		name               string
		parentPK           string
		sk                 string
		expectedFuncCalled bool
	}{
		{
			name:     "happy path",
			parentPK: "test-parent-pk",
			sk:       "test-sk",

			expectedFuncCalled: true,
		},
	}

	mockDbApi := &MockDatabaseAPI{
		DeleteItemFuncCalled: false,
		DeleteItemFunc:       nil,
	}

	mock := &ChildRepositoryImpl[models.Document]{
		databaseAPI: mockDbApi,
		table:       "test-table",
		index:       nil,
	}

	for _, tt := range documentTests {
		t.Run(tt.name, func(t *testing.T) {
			err := mock.Delete(tt.parentPK, tt.sk)
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if err == nil {
				if tt.expectedFuncCalled != mockDbApi.DeleteItemFuncCalled {
					t.Errorf("expected DeleteItem to be called, but got no call")
				}
			}
		})
	}
}

func TestChildRepositoryImpl_GetAllByParentPK(t *testing.T) {
	documentTests := []struct {
		name               string
		parentPK           string
		sk                 string
		expectedFuncCalled bool
	}{
		{
			name:     "happy path",
			parentPK: "test-parent-pk",
			sk:       "test-sk",

			expectedFuncCalled: true,
		},
	}

	mockDbApi := &MockDatabaseAPI{
		QueryFuncCalled: false,
		QueryFunc:       nil,
	}

	mock := &ChildRepositoryImpl[models.Document]{
		databaseAPI: mockDbApi,
		table:       "test-table",
		index:       nil,
	}

	for _, tt := range documentTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mock.GetAllByParentPK(tt.parentPK)
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if err == nil {
				if tt.expectedFuncCalled != mockDbApi.QueryFuncCalled {
					t.Errorf("expected Query to be called, but got no call")
				}
			}
		})
	}
}

// TODO: test is not complete - assert that BatchWriteItem was called
func TestChildRepositoryImpl_BatchDelete(t *testing.T) {
	documentTests := []struct {
		name               string
		parentPK           string
		sk                 string
		expectedFuncCalled bool
	}{
		{
			name:     "happy path",
			parentPK: "test-parent-pk",

			expectedFuncCalled: true,
		},
	}

	mockDbApi := &MockDatabaseAPI{
		QueryFunc:       nil,
		QueryFuncCalled: false,
	}

	mock := &ChildRepositoryImpl[models.Document]{
		databaseAPI: mockDbApi,
		table:       "test-table",
		index:       nil,
	}

	for _, tt := range documentTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mock.BatchDelete(tt.parentPK)
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if err == nil {
				if tt.expectedFuncCalled != mockDbApi.QueryFuncCalled {
					t.Errorf("expected BatchWriteItem to be called, but got no call")
				}
			}
		})
	}
}

func TestChildRepositoryImpl_BatchSave(t *testing.T) {
	documentTests := []struct {
		name               string
		item               []models.Document
		expectedFuncCalled bool
	}{
		{
			name: "happy path",
			item: []models.Document{},

			expectedFuncCalled: true,
		},
	}

	mockDbApi := &MockDatabaseAPI{
		BatchWriteItemFunc:       nil,
		BatchWriteItemFuncCalled: false,
	}

	mock := &ChildRepositoryImpl[models.Document]{
		databaseAPI: mockDbApi,
		table:       "test-table",
		index:       nil,
	}

	for _, tt := range documentTests {
		t.Run(tt.name, func(t *testing.T) {
			err := mock.BatchSave(tt.item)
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if err == nil {
				if tt.expectedFuncCalled != mockDbApi.BatchWriteItemFuncCalled {
					t.Errorf("expected BatchWriteItem to be called, but got no call")
				}
			}
		})
	}
}

func TestChildRepositoryImpl_UpdateProperty(t *testing.T) {
	documentTests := []struct {
		name               string
		parentPK           string
		sk                 string
		property           string
		propertyType       string
		atValue            types.AttributeValue
		expectedFuncCalled bool
	}{
		{
			name:         "happy path",
			parentPK:     "test-parent-pk",
			sk:           "test-sk",
			property:     "Title",
			propertyType: "s",
			atValue:      &types.AttributeValueMemberS{Value: "test-value"},

			expectedFuncCalled: true,
		},
	}

	mockDbApi := &MockDatabaseAPI{
		UpdateItemFunc:       nil,
		UpdateItemFuncCalled: false,
	}

	mock := &ChildRepositoryImpl[models.Document]{
		databaseAPI: mockDbApi,
		table:       "test-table",
		index:       nil,
	}

	for _, tt := range documentTests {
		t.Run(tt.name, func(t *testing.T) {
			err := mock.UpdateProperty(tt.parentPK, tt.sk, tt.property, tt.propertyType, tt.atValue)
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if err == nil {
				if tt.expectedFuncCalled != mockDbApi.UpdateItemFuncCalled {
					t.Errorf("expected BatchWriteItem to be called, but got no call")
				}
			}
		})
	}
}

func GetAllByParentPKRepositoryImpl_GetByPrimaryKey(t *testing.T) {
	documentTests := []struct {
		name               string
		parentPK           string
		sk                 string
		expectedFuncCalled bool
	}{
		{
			name:     "happy path",
			parentPK: "test-parent-pk",
			sk:       "test-sk",

			expectedFuncCalled: true,
		},
	}

	mockDbApi := &MockDatabaseAPI{
		GetItemFuncCalled: false,
		GetItemFunc:       nil,
	}

	mock := &ChildRepositoryImpl[models.Document]{
		databaseAPI: mockDbApi,
		table:       "test-table",
		index:       nil,
	}

	for _, tt := range documentTests {
		t.Run(tt.name, func(t *testing.T) {
			err := mock.Delete(tt.parentPK, tt.sk)
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if err == nil {
				if tt.expectedFuncCalled != mockDbApi.GetItemFuncCalled {
					t.Errorf("expected DeleteItem to be called, but got no call")
				}
			}
		})
	}
}
