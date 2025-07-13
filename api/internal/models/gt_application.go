package models

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/anfern777/go-serverless-framework-api/internal/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type PreScreeningStatus string
type Accommodation struct {
	Location string `json:"location"`
	From     string `json:"from"`
	To       string `json:"to"`
}
type EmployerInfo struct {
	Name          string        `json:"name"`
	Accommodation Accommodation `json:"accommodation"`
}
type GermanTrainingInfo struct {
	ExamDate string `json:"examDate"`
}

const (
	PQ   PreScreeningStatus = "PQ"
	PA   PreScreeningStatus = "PA"
	PFA  PreScreeningStatus = "PFA"
	DGKP PreScreeningStatus = "DGKP"
	NQ   PreScreeningStatus = "NQ"
)

type Application struct {
	PK                 string              `dynamodbav:"PK" json:"id"`
	SK                 string              `dynamodbav:"SK" json:"-"`
	GSI_PK             string              `dynamodbav:"GSI_PK" json:"-"` // matches entity type (prefix)
	CreatedAt          string              `dynamodbav:"CreatedAt" json:"createdAt"`
	Name               string              `dynamodbav:"Name" json:"name" validate:"required"`
	Email              string              `dynamodbav:"Email" json:"email" validate:"required,email"`
	Message            string              `dynamodbav:"Message" json:"message,omitempty"`
	Analysed           *bool               `dynamodbav:"Analysed" json:"analysed"`
	PreScreeningStatus PreScreeningStatus  `dynamodbav:"PreScreeningStatus" json:"preScreeningStatus"`
	CognitoID          *string             `dynamodbav:"CognitoID" json:"-"`
	GermanTrainingInfo *GermanTrainingInfo `dynamodbav:"GermanTrainingInfo" json:"germanTrainingInfo"`
	EmployerInfo       *EmployerInfo       `dynamodbav:"EmployerInfo" json:"employerInfo"`
}

func (app *Application) GenerateKeys() {
	app.PK = fmt.Sprintf("%s%s%s", common.PK_APP_PREFIX, common.PK_SEPARATOR, uuid.New().String())
	app.SK = app.PK
	app.GSI_PK = common.PK_APP_PREFIX
	app.CognitoID = aws.String("NA")
}

func (app *Application) GenerateAttributes() {
	app.Analysed = aws.Bool(false)
	app.PreScreeningStatus = PreScreeningStatus(PQ)
	app.CreatedAt = time.Now().Format(time.RFC3339)
}

func (app *Application) Validate(docs []Document) error {
	validate := validator.New()
	err := validate.Struct(app)
	if err != nil {
		return fmt.Errorf("application validation failed: missing or invalid fields: %w", err)
	}

	mandatoryDocuments := app.GetMandatoryDocuments()
	count := 0
	for _, d := range docs {
		sufix, exists := strings.CutPrefix(d.SK, common.DOCUMENT_LETTER_PREFIX+"-")
		if !exists {
			return fmt.Errorf("wrong document prefix")
		}
		if slices.Contains(mandatoryDocuments, string(sufix)) {
			count++
		}
	}
	if count != len(mandatoryDocuments) {
		return fmt.Errorf("wrong number of documents")
	}

	return nil
}

func (Application) GetMandatoryDocuments() []string {
	return []string{
		fmt.Sprint(DocumentTypeCV), fmt.Sprint(DocumentTypeConsent),
	}
}

func (app Application) GetEmailData() (string, string, string) {
	return app.Name, app.Email, app.Message
}
