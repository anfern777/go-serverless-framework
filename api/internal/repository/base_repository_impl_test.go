package repository

import (
	"testing"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/models"
)

func TestIndex(t *testing.T) {
	tests := []struct {
		name  string
		index *string

		expectedIndex *string
	}{
		{
			name:  "Index is set",
			index: common.PtrTo("test-index"),

			expectedIndex: common.PtrTo("test-index"),
		},
		{
			name:  "Index is not set",
			index: nil,

			expectedIndex: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := BaseRepositoryImpl[models.Application]{
				index: tt.index,
			}
			result := mockRepo.Index()
			if tt.expectedIndex == nil {
				if result != nil {
					t.Errorf("Expected index %s, but got %s", *tt.expectedIndex, *result)
				}
			} else {
				if *tt.expectedIndex != *result {
					t.Errorf("Expected index %s, but got %s", *tt.expectedIndex, *result)
				}
			}
		})
	}
}
