// Package customtypes provides shared value objects used across services,
// such as file attachment metadata backed by S3 and stored as PostgreSQL JSONB.
package customtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// ── AttachmentType ─────────────────────────────────────────────────────────────

// AttachmentType classifies an attachment into a broad media category.
type AttachmentType string

const (
	AttachmentTypeImage    AttachmentType = "image"
	AttachmentTypeVideo    AttachmentType = "video"
	AttachmentTypePDF      AttachmentType = "pdf"
	AttachmentTypeDocument AttachmentType = "document"
	AttachmentTypeAudio    AttachmentType = "audio"
	AttachmentTypeOther    AttachmentType = "other"
)

// mimeSubtypeGroups maps MIME prefixes to their AttachmentType.
// The first matching prefix wins, so specific types (e.g. pdf) must come
// before their parent prefix (e.g. application/).
var mimeSubtypeGroups = map[string]AttachmentType{
	"image/":                         AttachmentTypeImage,
	"video/":                         AttachmentTypeVideo,
	"audio/":                         AttachmentTypeAudio,
	"application/pdf":                AttachmentTypePDF,
	"text/":                          AttachmentTypeDocument,
	"application/msword":             AttachmentTypeDocument,
	"application/vnd.openxmlformats": AttachmentTypeDocument, // .docx, .xlsx, .pptx, etc.
	"application/rtf":                AttachmentTypeDocument,
	"text/csv":                       AttachmentTypeDocument,
	"text/plain":                     AttachmentTypeDocument,
	// application/octet-stream falls through to AttachmentTypeOther (default)
}

// InferAttachmentType returns the logical type for a given MIME content-type string.
func InferAttachmentType(contentType string) AttachmentType {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	for prefix, t := range mimeSubtypeGroups {
		if strings.HasPrefix(ct, prefix) {
			return t
		}
	}
	return AttachmentTypeOther
}

// ── Attachment ─────────────────────────────────────────────────────────────────

// Attachment is an immutable value object representing an S3 object.
// It implements sql.Scanner and driver.Valuer so it can be stored as a
// PostgreSQL JSONB column, and json.Marshaler/json.Unmarshaler for
// serialisation in API responses and RabbitMQ events.
type Attachment struct {
	// Key is the S3 object key (required).
	Key string `json:"key"`
	// Bucket is the S3 bucket name (required).
	// Name is the original filename (optional).
	Name string `json:"name,omitempty"`
	// Size is the file size in bytes (optional).
	Size int64 `json:"size,omitempty"`
	// ContentType is the MIME type, e.g. "image/png" (required).
	ContentType string `json:"content_type"`
	// Type is the logical media category, derived from ContentType.
	Type AttachmentType `json:"type"`
	// UploadedAt is the upload timestamp (required).

	// s3BaseURL is the S3 endpoint used by String(). It defaults to the
	// classic S3 global endpoint and can be overridden via WithS3BaseURL.
}

// attachmentJSON is the internal serialisation struct (avoids infinite recursion
// in Scan/Value by not exposing s3BaseURL).
type attachmentJSON struct {
	Key         string         `json:"key"`
	Name        string         `json:"name,omitempty"`
	Size        int64          `json:"size,omitempty"`
	ContentType string         `json:"content_type"`
	Type        AttachmentType `json:"type"`
}

func (a Attachment) toJSON() attachmentJSON {
	return attachmentJSON{
		Key:         a.Key,
		Name:        a.Name,
		Size:        a.Size,
		ContentType: a.ContentType,
		Type:        a.Type,
	}
}

// ── sql.Scanner ───────────────────────────────────────────────────────────────

// Scan implements database/sql.Scanner. It deserialises a JSONB value from
// PostgreSQL back into an Attachment.
func (a *Attachment) Scan(src any) error {
	if src == nil {
		*a = Attachment{}
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Attachment.Scan: expected []byte, got %T", src)
	}
	var aj attachmentJSON
	if err := json.Unmarshal(b, &aj); err != nil {
		return fmt.Errorf("Attachment.Scan: failed to unmarshal JSON: %w", err)
	}
	*a = Attachment{
		Key:         aj.Key,
		Name:        aj.Name,
		Size:        aj.Size,
		ContentType: aj.ContentType,
		Type:        aj.Type,
	}
	return nil
}

// ── driver.Valuer ─────────────────────────────────────────────────────────────

// Value implements database/sql/driver.Valuer. It serialises the Attachment
// as a JSONB value for PostgreSQL.
func (a Attachment) Value() (driver.Value, error) {
	return json.Marshal(a.toJSON())
}

// ── json.Marshaler / json.Unmarshaler ─────────────────────────────────────────

// MarshalJSON implements json.Marshaler.
func (a Attachment) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.toJSON())
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Attachment) UnmarshalJSON(data []byte) error {
	var aj attachmentJSON
	if err := json.Unmarshal(data, &aj); err != nil {
		return err
	}
	*a = Attachment{
		Key:         aj.Key,
		Name:        aj.Name,
		Size:        aj.Size,
		ContentType: aj.ContentType,
		Type:        aj.Type,
	}
	return nil
}

// ── Attachments ────────────────────────────────────────────────────────────────

// Attachments is a typed slice of Attachment value objects.
// It implements sql.Scanner, driver.Valuer, json.Marshaler, and
// json.Unmarshaler so it can be stored as a PostgreSQL JSONB column
// and serialised in API responses and RabbitMQ events.
type Attachments []Attachment

// attachmentsJSON is the internal serialisation slice alias.
type attachmentsJSON []attachmentJSON

func (as Attachments) toJSON() attachmentsJSON {
	if as == nil {
		return nil
	}
	out := make(attachmentsJSON, len(as))
	for i, a := range as {
		out[i] = a.toJSON()
	}
	return out
}

// Scan implements database/sql.Scanner. It deserialises a JSONB array
// from PostgreSQL back into an Attachments slice.
func (as *Attachments) Scan(src any) error {
	if src == nil {
		*as = nil
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Attachments.Scan: expected []byte, got %T", src)
	}
	var ajs attachmentsJSON
	if err := json.Unmarshal(b, &ajs); err != nil {
		return fmt.Errorf("Attachments.Scan: failed to unmarshal JSON: %w", err)
	}
	if ajs == nil {
		*as = nil
		return nil
	}
	out := make(Attachments, len(ajs))
	for i, aj := range ajs {
		out[i] = Attachment{
			Key:         aj.Key,
			Name:        aj.Name,
			Size:        aj.Size,
			ContentType: aj.ContentType,
			Type:        aj.Type,
		}
	}
	*as = out
	return nil
}

// Value implements database/sql/driver.Valuer. It serialises the Attachments
// slice as a JSONB array for PostgreSQL.
func (as Attachments) Value() (driver.Value, error) {
	if as == nil {
		return nil, nil
	}
	return json.Marshal(as.toJSON())
}

// MarshalJSON implements json.Marshaler.
func (as Attachments) MarshalJSON() ([]byte, error) {
	if as == nil {
		return []byte("null"), nil
	}
	return json.Marshal(as.toJSON())
}

// UnmarshalJSON implements json.Unmarshaler.
func (as *Attachments) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*as = nil
		return nil
	}
	var ajs attachmentsJSON
	if err := json.Unmarshal(data, &ajs); err != nil {
		return err
	}
	if ajs == nil {
		*as = nil
		return nil
	}
	out := make(Attachments, len(ajs))
	for i, aj := range ajs {
		out[i] = Attachment{
			Key:         aj.Key,
			Name:        aj.Name,
			Size:        aj.Size,
			ContentType: aj.ContentType,
			Type:        aj.Type,
		}
	}
	*as = out
	return nil
}
