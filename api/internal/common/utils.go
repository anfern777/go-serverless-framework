package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type RequestError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func (re *RequestError) Error() string {
	return re.Message
}

type SQSBatchItemFailureError struct {
	Message    string `json:"message"`
	MessageID  string `json:"messageID"`
	StatusCode int    `json:"statusCode"`
}

func (sqse *SQSBatchItemFailureError) Error() string {
	return sqse.Message
}

func RequestErrorResponse(statusCode int, message string, logger services.Logger, args ...any) (*events.APIGatewayV2HTTPResponse, error) {
	errorResponse := &RequestError{
		Message:    message,
		StatusCode: statusCode,
	}

	jsonResponse, err := json.MarshalIndent(errorResponse, "", "\t")
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       message,
		}, err
	}

	logger.Log(
		message,
		slog.LevelError,
		"StatusCode", statusCode,
		args,
	)

	if statusCode >= 500 {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: statusCode,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(jsonResponse),
		}, errorResponse
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(jsonResponse),
	}, nil
}

func GetSsmClient() (*ssm.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("utils: failed to load ssm config: %v", err)
	}

	return ssm.NewFromConfig(cfg), nil
}

func GetSesClient() (*ses.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("utils: failed to load ses config: %v", err)
	}

	return ses.NewFromConfig(cfg), nil
}
func GetCognitoClient() (*cognito.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("utils: failed to load cognito config: %v", err)
	}

	return cognito.NewFromConfig(cfg), nil
}

func GetContentTypeFromExtension(extension string) (*string, error) {
	var contentType string
	switch extension {
	case ".pdf":
		contentType = "application/pdf"
	case ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".doc":
		contentType = "application/msword"
	case ".png":
		contentType = "image/png"
	case ".mp3":
		contentType = "audio/mpeg"
	default:
		return nil, errors.New("unknown extension")
	}
	return &contentType, nil
}

func GetDocumentExtension(fileName string) (string, error) {
	ext := filepath.Ext(fileName)
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" && ext != ".png" && ext != ".mp3" {
		return "", fmt.Errorf("validation Error - invalid file extension")
	}
	return ext, nil
}

func GetFileContent(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to perform get request: %w", err)
	}
	defer response.Body.Close()

	fileContent, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to perform get request: %w", err)
	}

	return fileContent, nil
}

func PtrTo[T any](value T) *T {
	return &value
}

func GetValueFromGenericStruct(obj any, fieldName string) (any, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("item must be a struct or a pointer to a struct, got %T", val.Kind())
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("item struct does not have a field named %s", fieldName)
	}

	var value any
	if field.CanInterface() {
		value = field.Interface()
	} else {
		return nil, fmt.Errorf("PK field is not exportable")
	}
	return value, nil
}

func CreatePK(prefix string, id string) string {
	return strings.Join([]string{prefix, PK_SEPARATOR, id}, "")
}
