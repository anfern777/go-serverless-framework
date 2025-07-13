package common

import "slices"

type LanguageCode string

const (
	PK_SEPARATOR string = "#"

	// LAMBDA ENV KEYS
	LAMBDAENV_BUCKET_NAME    string = "BUCKET_NAME"
	LAMBDA_ENV_TABLE_NAME    string = "TABLE_NAME"
	LAMBDA_ENV_GSI_NAME      string = "GSI_NAME"
	LAMBDA_ENV_SQS_QUEUE_URL string = "SQS_QUEUE_URL"

	// PATH PARAMETER NAMES
	PATH_PARAM_ID    = "id"
	PATH_PARAM_EMAIL = "email"

	// QUERY PARAMETER NAMES
	QUERY_PARAM_LANGUAGE = "lang"

	// DOCUMENTS
	SK_SEPARATOR           string = "-"
	DOCUMENT_LETTER_PREFIX string = "Document"

	// MESSAGES
	MESSAGE_LETTER_PREFIX string = "Message"

	// AUTHORIZATION MIDDLEWARE
	AUHT_SCOPE string = "scope"

	// APPLICATIONS
	PK_APP_PREFIX string = "APP"

	// POSTS
	PK_POST_PREFIX    string = "Post"
	SK_POST_SEPARATOR string = "#"

	// Language Codes
	EN LanguageCode = "en"
	DE LanguageCode = "de"
	TL LanguageCode = "tl"
	AT LanguageCode = "at"
)

var allowedLanguages = []LanguageCode{EN, AT, DE, TL}

func IsValidLanguage(lang LanguageCode) bool {
	return slices.Contains(allowedLanguages, lang)
}
