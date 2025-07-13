package models

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/anfern777/go-serverless-framework-api/internal/common"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Post struct {
	PK        string `dynamodbav:"PK" json:"id" validate:"required"`
	SK        string `dynamodbav:"SK" json:"-" validate:"required"`
	GSI_PK    string `dynamodbav:"GSI_PK" json:"-"` // matches entity type (prefix) + language (en | tg | de | at)
	CreatedAt string `dynamodbav:"CreatedAt" json:"createdAt" validate:"required"`
	Title     string `dynamodbav:"Title" json:"title" validate:"required"`
	Content   string `dynamodbav:"Content" json:"content" validate:"required"`
}

func (p *Post) GenerateKeys(language common.LanguageCode) error {
	if !common.IsValidLanguage(language) {
		return errors.New("validation error: invalid language")
	}

	p.PK = fmt.Sprintf("%s%s", common.PK_POST_PREFIX, uuid.New().String())
	p.SK = p.PK
	p.GSI_PK = GetPostSK(common.PK_POST_PREFIX, language)
	return nil
}

func (p *Post) GenerateAttributes() {
	p.CreatedAt = time.Now().Format(time.RFC3339)
}

func (p *Post) Validate(docs *[]Document) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(p)
	if err != nil {
		return fmt.Errorf("validation failed -> missing or invalid fields: %w", err)
	}
	mandatoryParts := p.GetMandatoryDocuments()
	count := 0
	for _, d := range *docs {
		if slices.Contains(mandatoryParts, string(d.SK)) {
			count++
		}
	}
	if count != len(mandatoryParts) {
		return fmt.Errorf("wrong number of documents")
	}

	return nil
}

func (Post) GetMandatoryDocuments() []string {
	return []string{
		fmt.Sprint(DocumentTypeThumbnail),
	}
}

func GetPostSK(prefix string, language common.LanguageCode) string {
	return common.PK_POST_PREFIX + string(language) + common.SK_POST_SEPARATOR
}
