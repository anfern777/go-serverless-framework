package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/anfern777/go-serverless-framework-api/internal/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-playground/validator/v10"
)

type DocStatus int
type DocumentType string

const (
	DocStatusRequested DocStatus = iota
	DocStatusApproved
	DocStatusRejected
	DocStatusUnderAnalysis
	DocStatusNotApplicable
)

// String returns the canonical string representation for DocStatus.
func (ds DocStatus) String() string {
	switch ds {
	case DocStatusRequested:
		return "REQUESTED"
	case DocStatusApproved:
		return "APPROVED"
	case DocStatusRejected:
		return "REJECTED"
	case DocStatusUnderAnalysis:
		return "UNDER_ANALYSIS"
	case DocStatusNotApplicable:
		return "NOT_APPLICABLE"
	default:
		return fmt.Sprintf("UnknownDocStatus(%d)", ds)
	}
}

const (
	// Pre-Screening
	DocumentTypeConsent DocumentType = "Consent"
	DocumentTypeCV      DocumentType = "CV"

	// Companies
	DocumentTypeJobAdvertisement DocumentType = "JobAd"

	// German Training
	DocumentTypeGLTEP DocumentType = "GLTEP"
	DocumentTypeB2EC  DocumentType = "B2EC"

	// Employer
	DocumentTypeEmployersDeclaration DocumentType = "EmpDecl"

	// Nostrification/RWR
	// School Documents
	DocumentTypeHSD        DocumentType = "HSD"
	DocumentTypeF137       DocumentType = "Form137"
	DocumentTypeCMIEnglish DocumentType = "CMIEnglish"
	DocumentTypeCGHS       DocumentType = "CGHS"
	DocumentTypeCD         DocumentType = "CD"
	DocumentTypeTOR        DocumentType = "TOR"
	DocumentTypeTORCTC     DocumentType = "TORCTC"
	DocumentTypeRLE        DocumentType = "RLE"
	DocumentTypeRLECTC     DocumentType = "RLECTC"
	DocumentTypeCGC        DocumentType = "CGC"
	DocumentTypeCMI        DocumentType = "CMI"
	DocumentTypeF137CTC    DocumentType = "Form137CTC"
	DocumentTypeHSDCTC     DocumentType = "HSDCTC"

	// Professional Documents
	DocumentTypePRCID       DocumentType = "PRCID"
	DocumentTypePRCIDCTC    DocumentType = "PRCIDCTC"
	DocumentTypeCGS         DocumentType = "CGS"
	DocumentTypePRCBCRating DocumentType = "PRCBCRating"
	DocumentTypePRCBCPasser DocumentType = "PRCBCPasser"
	DocumentTypeCE          DocumentType = "CE"

	// Personal Documents
	DocumentTypePassportCTC         DocumentType = "PassportCTC"
	DocumentTypeMarriageCertificate DocumentType = "MarriageCert" // Note: Constant name reflects corrected spelling
	DocumentTypeNBIClearance        DocumentType = "NBIClearance"
	DocumentTypePicture             DocumentType = "Picture"
	DocumentTypeCVGerman            DocumentType = "CVGerman"
	DocumentTypeAFN                 DocumentType = "AFN"
	DocumentTypePOAN                DocumentType = "POAN"
	DocumentTypePOARWR              DocumentType = "POARWR"
	DocumentTypeAFRWR               DocumentType = "AFRWR"
	DocumentTypeCertificateEnglish  DocumentType = "CertEnglish"

	DocumentTypeThumbnail DocumentType = "Thumbnail"
	DocumentTypeAudioRec  DocumentType = "PostAudioRec"
)

type Document struct {
	PK          string    `dynamodbav:"PK" json:"id" validate:"required"`           // matches PK of parent entity
	SK          string    `dynamodbav:"SK" json:"documentType" validate:"required"` // document type
	Name        *string   `dynamodbav:"Name" json:"name" validate:"required"`
	ContentType *string   `dynamodbav:"ContentType" json:"contentType" validate:"required"`
	SizeBytes   *int64    `dynamodbav:"SizeBytes" json:"sizeBytes" validate:"required"`
	RequestedAt string    `dynamodbav:"RequestedAt" json:"requestedAt" validate:"required"`
	UploadedAt  *string   `dynamodbav:"uploadedAt" json:"uploadedAt" validate:"required"`
	Status      DocStatus `dynamodbav:"Status" json:"status" validate:"required"`
	Notes       string    `dynamodbav:"Notes" json:"Notes"`
}

func (doc *Document) GenerateKeys(parentEntityPK string, documentType DocumentType) {
	doc.PK = parentEntityPK
	doc.SK = CreateDocumentSK(documentType)
}

func (doc *Document) GenerateAttributesForDocumentRequest(applicationId string, note string) error {
	doc.Status = DocStatusRequested
	doc.RequestedAt = time.Now().Format(time.RFC3339)
	doc.Notes = note
	return nil
}

func (doc *Document) GenerateAttributesForDocumentUpload(fileName string, sizeBytes int64) error {
	extension, err := common.GetDocumentExtension(fileName)
	if err != nil {
		return fmt.Errorf("failed to get extension from document: %w", err)
	}

	contentType, err := common.GetContentTypeFromExtension(extension)
	if err != nil {
		return fmt.Errorf("failed to get content type from extension: %w", err)
	}
	doc.ContentType = contentType
	doc.SizeBytes = &sizeBytes
	doc.UploadedAt = aws.String(time.Now().Format(time.RFC3339))
	doc.Status = DocStatusUnderAnalysis

	doc.Name = aws.String(fmt.Sprintf("%s%s%s%s", doc.PK, common.SK_SEPARATOR, doc.SK, extension))
	return nil
}

func (d *Document) Validate() error {
	validate := validator.New()
	err := validate.Struct(d)
	if err != nil {
		return fmt.Errorf("document validation failed: missing or invalid fields: %w", err)
	}
	maxSize, err := d.getMaxFileSize()
	if err != nil {
		return fmt.Errorf("document validation failed: %w", err)
	}
	if *d.SizeBytes > maxSize {
		return fmt.Errorf("document validation failed: document size exceeds maximum allowed size: %d", maxSize)
	}
	return nil
}

func (d *Document) getMaxFileSize() (int64, error) {
	sufix, _ := strings.CutPrefix(d.SK, common.DOCUMENT_LETTER_PREFIX+common.SK_SEPARATOR)

	switch DocumentType(sufix) {
	case DocumentTypeConsent:
		return 5 * 1024 * 1024, nil
	case DocumentTypeCV:
		return 5 * 1024 * 1024, nil
	case DocumentTypeJobAdvertisement:
		return 5 * 1024 * 1024, nil
	case DocumentTypeThumbnail:
		return 1 * 1024 * 1024, nil
	case DocumentTypeGLTEP:
		return 5 * 1024 * 1024, nil
	case DocumentTypeAudioRec:
		return 40 * 1024 * 1024, nil

	// German Traning
	case DocumentTypeB2EC:
		return 5 * 1024 * 1024, nil

	// Employer
	case DocumentTypeEmployersDeclaration:
		return 5 * 1024 * 1024, nil

	// School Documents
	case DocumentTypeCMIEnglish:
		return 5 * 1024 * 1024, nil
	case DocumentTypeCGHS:
		return 5 * 1024 * 1024, nil
	case DocumentTypeCD:
		return 5 * 1024 * 1024, nil
	case DocumentTypeTOR:
		return 5 * 1024 * 1024, nil
	case DocumentTypeRLE:
		return 5 * 1024 * 1024, nil
	case DocumentTypeCGC:
		return 5 * 1024 * 1024, nil
	case DocumentTypeCMI:
		return 5 * 1024 * 1024, nil
	case DocumentTypeRLECTC:
		return 5 * 1024 * 1024, nil
	case DocumentTypeTORCTC:
		return 5 * 1024 * 1024, nil
	case DocumentTypeF137CTC:
		return 5 * 1024 * 1024, nil
	case DocumentTypeHSDCTC:
		return 5 * 1024 * 1024, nil

	// Professional Documents
	case DocumentTypePRCID:
		return 5 * 1024 * 1024, nil
	case DocumentTypePRCIDCTC:
		return 5 * 1024 * 1024, nil
	case DocumentTypeCGS:
		return 5 * 1024 * 1024, nil
	case DocumentTypePRCBCRating:
		return 5 * 1024 * 1024, nil
	case DocumentTypePRCBCPasser:
		return 5 * 1024 * 1024, nil
	case DocumentTypeCE:
		return 5 * 1024 * 1024, nil

	// Personal Documents
	case DocumentTypePassportCTC:
		return 5 * 1024 * 1024, nil
	case DocumentTypeMarriageCertificate:
		return 5 * 1024 * 1024, nil
	case DocumentTypeNBIClearance:
		return 5 * 1024 * 1024, nil
	case DocumentTypePicture:
		return 1 * 1024 * 1024, nil
	case DocumentTypeCVGerman:
		return 5 * 1024 * 1024, nil
	case DocumentTypeAFN:
		return 5 * 1024 * 1024, nil
	case DocumentTypePOAN:
		return 5 * 1024 * 1024, nil
	case DocumentTypePOARWR:
		return 5 * 1024 * 1024, nil
	case DocumentTypeAFRWR:
		return 5 * 1024 * 1024, nil
	case DocumentTypeCertificateEnglish:
		return 5 * 1024 * 1024, nil
	default:
		return -1, nil
	}
}

func CreateDocumentSK(documentType DocumentType) string {
	return common.DOCUMENT_LETTER_PREFIX + common.SK_SEPARATOR + string(documentType)
}
