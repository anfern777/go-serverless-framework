package services

type EmailType int

const (
	EMAIL_TYPE_JOB_ADVERTISEMENT_ADMIN EmailType = iota
	EMAIL_TYPE_JOB_ADVERTISEMENT_USER
	EMAIL_TYPE_APPLICATION_ADMIN
	EMAIL_TYPE_APPLICATION_USER
	EMAIL_TYPE_EMPLOYERS_DECLARATION
	EMAIL_TYPE_NEW_MESSAGE_REVEIVED_ADMIN
	EMAIL_TYPE_NEW_MESSAGE_REVEIVED_USER
)

type Attachment struct {
	Name        string
	ContentType string
}

type EmailerInputParams struct {
	Template    string
	Subject     string
	Source      string
	Destination string
	Data        map[string]string
	Attachments []Attachment
}

type Emailer interface {
	Email(EmailerInputParams) error
}
