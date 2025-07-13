package models

import (
	"fmt"
	"time"

	"github.com/anfern777/go-serverless-framework-api/internal/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MessageAuthor string

const (
	MESSAGE_AUTHOR_ADMIN string = "admin"
	MESSAGE_AUTHOR_USER  string = "user"
)

type Message struct {
	PK        string        `dynamodbav:"PK" json:"id" validate:"required"`
	SK        string        `dynamodbav:"SK" json:"-" validate:"required"` // Matches entity type + uuid = "Message#<uuid>"
	Author    MessageAuthor `dynamodbav:"Author" json:"author" validate:"required"`
	IsRead    *bool         `dynamodbav:"IsRead" json:"isRead" validate:"required"`
	Content   string        `dynamodbav:"Content" json:"content" validate:"required"`
	CreatedAt string        `dynamodbav:"CreatedAt" json:"createdAt" validate:"required"`
}

func (m *Message) GenerateKeys(parentEntityPK string) {
	m.PK = parentEntityPK
	m.SK = CreateMessagetSK()
}

func (m *Message) GenerateAttributes(author MessageAuthor) {
	m.Author = author
	m.CreatedAt = time.Now().Format(time.RFC3339)
	m.IsRead = aws.Bool(false)
}

func (m *Message) Validate() error {
	validate := validator.New()
	err := validate.Struct(m)
	if err != nil {
		return fmt.Errorf("message validation failed: missing or invalid fields: %w", err)
	}
	return nil
}

func CreateMessagetSK() string {
	return common.MESSAGE_LETTER_PREFIX + common.SK_SEPARATOR + uuid.New().String()
}
