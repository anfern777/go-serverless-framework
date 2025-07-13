package implementations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"gopkg.in/gomail.v2"
)

type SesEmailerImpl struct {
	client    SesClientAPI
	fsService services.FileStorageService
}

type SesClientAPI interface {
	SendRawEmail(ctx context.Context, params *ses.SendRawEmailInput, optFns ...func(*ses.Options)) (*ses.SendRawEmailOutput, error)
}

// TODO: make fs service optional and add opt function
type SesEmailerImplOption func(e *SesEmailerImpl)

func WithFsService(fs services.FileStorageService) SesEmailerImplOption {
	return func(e *SesEmailerImpl) {
		e.fsService = fs
	}
}

func NewSesEmailerImpl(opts ...SesEmailerImplOption) (*SesEmailerImpl, error) {
	client, err := common.GetSesClient()
	if err != nil {
		return nil, fmt.Errorf("failed go get ses client: %w", err)
	}

	e := &SesEmailerImpl{
		client: client,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

func (e *SesEmailerImpl) Email(params services.EmailerInputParams) error {
	re := regexp.MustCompile(`\{\{\.([^}]+)\}\}`)
	replacedTemplate := re.ReplaceAllStringFunc(params.Template, func(match string) string {
		key := strings.TrimPrefix(match, "{{.")
		key = strings.TrimSuffix(key, "}}")

		if val, ok := params.Data[key]; ok {
			return val
		}
		return ""
	})

	sendEmailRawInput := &ses.SendRawEmailInput{
		Destinations: []string{params.Destination},
		Source:       &params.Source,
	}

	rawMessage := gomail.NewMessage()
	rawMessage.SetHeader("From", params.Source)
	rawMessage.SetHeader("To", params.Destination)
	rawMessage.SetHeader("Subject", params.Subject)
	rawMessage.SetBody("text/html", replacedTemplate)

	if len(params.Attachments) > 0 {
		for _, att := range params.Attachments {
			pr, err := e.fsService.GrantReadAccess(att.Name)
			if err != nil {
				return fmt.Errorf("failed to grant file write access: %w", err)
			}
			fileContent, err := common.GetFileContent(pr.URL)
			if err != nil {
				return fmt.Errorf("failed to get file content: %w", err)
			}
			log.Printf("Fetched file content for attachment '%s'. Length: %d bytes. Is nil: %t", att.Name, len(fileContent), fileContent == nil)
			if fileContent != nil {
				rawMessage.Attach(att.Name,
					gomail.SetHeader(map[string][]string{"Content-Type": {att.ContentType}}),
					gomail.SetCopyFunc(func(w io.Writer) error {
						_, err := w.Write(fileContent)
						return err
					}))
			}
		}
	}

	var emailRaw bytes.Buffer
	if _, err := rawMessage.WriteTo(&emailRaw); err != nil {
		return fmt.Errorf("failed to write raw email message: %w", err)
	}

	sendEmailRawInput.RawMessage = &types.RawMessage{Data: emailRaw.Bytes()}

	_, err := e.client.SendRawEmail(context.TODO(), sendEmailRawInput)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
