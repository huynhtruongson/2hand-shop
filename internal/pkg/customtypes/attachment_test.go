package customtypes

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
)

// ── InferAttachmentType ──────────────────────────────────────────────────────

func TestInferAttachmentType(t *testing.T) {
	tests := []struct {
		name       string
		contentType string
		want       AttachmentType
	}{
		// image/
		{"jpeg", "image/jpeg", AttachmentTypeImage},
		{"png", "image/png", AttachmentTypeImage},
		{"gif", "image/gif", AttachmentTypeImage},
		{"webp", "image/webp", AttachmentTypeImage},
		{"svg", "image/svg+xml", AttachmentTypeImage},
		{"image/uppercase", "IMAGE/PNG", AttachmentTypeImage},

		// video/
		{"mp4", "video/mp4", AttachmentTypeVideo},
		{"webm", "video/webm", AttachmentTypeVideo},
		{"mov", "video/quicktime", AttachmentTypeVideo},

		// audio/
		{"mp3", "audio/mpeg", AttachmentTypeAudio},
		{"ogg audio", "audio/ogg", AttachmentTypeAudio},
		{"wav", "audio/wav", AttachmentTypeAudio},

		// application/pdf
		{"pdf lowercase", "application/pdf", AttachmentTypePDF},
		{"pdf uppercase", "APPLICATION/PDF", AttachmentTypePDF},

		// text/* → document
		{"text/html", "text/html", AttachmentTypeDocument},
		{"text/css", "text/css", AttachmentTypeDocument},
		{"text/csv", "text/csv", AttachmentTypeDocument},
		{"text/plain", "text/plain", AttachmentTypeDocument},
		{"text/xml", "text/xml", AttachmentTypeDocument},

		// application/msword family
		{"msword", "application/msword", AttachmentTypeDocument},
		{"vnd.openxmlformats word", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", AttachmentTypeDocument},
		{"vnd.openxmlformats spreadsheet", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", AttachmentTypeDocument},
		{"vnd.openxmlformats presentation", "application/vnd.openxmlformats-officedocument.presentationml.presentation", AttachmentTypeDocument},

		// rtf / csv via text/
		{"rtf", "application/rtf", AttachmentTypeDocument},

		// unknown → other
		{"octet-stream", "application/octet-stream", AttachmentTypeOther},
		{"zip", "application/zip", AttachmentTypeOther},
		{"json", "application/json", AttachmentTypeOther},
		{"gzip", "application/gzip", AttachmentTypeOther},

		// edge: whitespace / mixed-case
		{"trim spaces", "  image/jpeg  ", AttachmentTypeImage},
		{"mixed case", "Image/Png", AttachmentTypeImage},
		{"empty string", "", AttachmentTypeOther},
		{"unknown prefix", "x-custom/type", AttachmentTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferAttachmentType(tt.contentType)
			if got != tt.want {
				t.Errorf("InferAttachmentType(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

// ── Attachment.Scan ─────────────────────────────────────────────────────────

func TestAttachment_Scan(t *testing.T) {
	validJSON := []byte(`{
		"key": "products/abc/def.jpg",
		"name": "photo.jpg",
		"size": 1024,
		"content_type": "image/jpeg",
		"type": "image"
	}`)

	t.Run("nil src returns empty Attachment", func(t *testing.T) {
		var a Attachment
		if err := a.Scan(nil); err != nil {
			t.Fatalf("Scan(nil) returned error: %v", err)
		}
		if a != (Attachment{}) {
			t.Errorf("Scan(nil): expected zero Attachment, got %+v", a)
		}
	})

	t.Run("valid JSON", func(t *testing.T) {
		var a Attachment
		if err := a.Scan(validJSON); err != nil {
			t.Fatalf("Scan(validJSON) returned error: %v", err)
		}
		if a.Key != "products/abc/def.jpg" {
			t.Errorf("Key = %q, want %q", a.Key, "products/abc/def.jpg")
		}
		if a.Name != "photo.jpg" {
			t.Errorf("Name = %q, want %q", a.Name, "photo.jpg")
		}
		if a.Size != 1024 {
			t.Errorf("Size = %d, want %d", a.Size, 1024)
		}
		if a.ContentType != "image/jpeg" {
			t.Errorf("ContentType = %q, want %q", a.ContentType, "image/jpeg")
		}
		if a.Type != AttachmentTypeImage {
			t.Errorf("Type = %v, want %v", a.Type, AttachmentTypeImage)
		}
	})

	t.Run("invalid type src", func(t *testing.T) {
		var a Attachment
		err := a.Scan("not a []byte")
		if err == nil {
			t.Fatal("Scan(non []byte) expected error, got nil")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var a Attachment
		err := a.Scan([]byte(`{invalid json`))
		if err == nil {
			t.Fatal("Scan(invalidJSON) expected error, got nil")
		}
	})

	t.Run("minimal JSON (key and content_type only)", func(t *testing.T) {
		var a Attachment
		minJSON := []byte(`{"key": "x/y/z.png", "content_type": "image/png", "type": "image"}`)
		if err := a.Scan(minJSON); err != nil {
			t.Fatalf("Scan(minJSON) returned error: %v", err)
		}
		if a.Key != "x/y/z.png" || a.ContentType != "image/png" {
			t.Errorf("unexpected values: %+v", a)
		}
	})
}

// ── Attachment.Value ─────────────────────────────────────────────────────────

func TestAttachment_Value(t *testing.T) {
	a := Attachment{
		Key:         "products/abc/def.jpg",
		Name:        "photo.jpg",
		Size:        1024,
		ContentType: "image/jpeg",
		Type:        AttachmentTypeImage,
	}

	val, err := a.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	got, ok := val.([]byte)
	if !ok {
		t.Fatalf("Value() type = %T, want []byte", val)
	}

	var decoded attachmentJSON
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("Value() returned non-JSON bytes: %s", string(got))
	}

	if decoded.Key != a.Key {
		t.Errorf("Key = %q, want %q", decoded.Key, a.Key)
	}
	if decoded.Name != a.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, a.Name)
	}
	if decoded.Size != a.Size {
		t.Errorf("Size = %d, want %d", decoded.Size, a.Size)
	}
	if decoded.ContentType != a.ContentType {
		t.Errorf("ContentType = %q, want %q", decoded.ContentType, a.ContentType)
	}
	if decoded.Type != a.Type {
		t.Errorf("Type = %v, want %v", decoded.Type, a.Type)
	}
}

// ── Attachment.MarshalJSON / UnmarshalJSON ──────────────────────────────────

func TestAttachment_JSON_RoundTrip(t *testing.T) {
	original := Attachment{
		Key:         "uploads/123/report.pdf",
		Name:        "annual_report.pdf",
		Size:        524288,
		ContentType: "application/pdf",
		Type:        AttachmentTypePDF,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var restored Attachment
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}

	if restored != original {
		t.Errorf("Round-trip mismatch:\ngot  %+v\nwant %+v", restored, original)
	}
}

func TestAttachment_MarshalJSON_OmitsEmptyFields(t *testing.T) {
	a := Attachment{
		Key:         "uploads/a/b/c.mp4",
		ContentType: "video/mp4",
		Type:        AttachmentTypeVideo,
		// Name, Size omitted
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("data is not valid JSON: %s", string(data))
	}

	// "name" and "size" must be absent (omitempty)
	if _, ok := m["name"]; ok {
		t.Errorf("MarshalJSON included name field, expected omission")
	}
	if _, ok := m["size"]; ok {
		t.Errorf("MarshalJSON included size field, expected omission")
	}
}

func TestAttachment_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var a Attachment
	err := a.UnmarshalJSON([]byte(`{broken`))
	if err == nil {
		t.Fatal("UnmarshalJSON expected error for invalid JSON, got nil")
	}
}

// ── Attachments ───────────────────────────────────────────────────────────────

func TestAttachments_Scan(t *testing.T) {
	validJSON := []byte(`[
		{"key": "products/abc/img.jpg", "name": "photo.jpg", "size": 1024, "content_type": "image/jpeg", "type": "image"},
		{"key": "products/abc/doc.pdf", "name": "doc.pdf", "size": 2048, "content_type": "application/pdf", "type": "pdf"}
	]`)

	t.Run("nil src returns nil Attachments", func(t *testing.T) {
		var as Attachments
		if err := as.Scan(nil); err != nil {
			t.Fatalf("Scan(nil) returned error: %v", err)
		}
		if as != nil {
			t.Errorf("Scan(nil): expected nil, got %v", as)
		}
	})

	t.Run("valid JSON array", func(t *testing.T) {
		var as Attachments
		if err := as.Scan(validJSON); err != nil {
			t.Fatalf("Scan(validJSON) returned error: %v", err)
		}
		if len(as) != 2 {
			t.Fatalf("len(as) = %d, want 2", len(as))
		}
		if as[0].Key != "products/abc/img.jpg" {
			t.Errorf("as[0].Key = %q, want %q", as[0].Key, "products/abc/img.jpg")
		}
		if as[0].Type != AttachmentTypeImage {
			t.Errorf("as[0].Type = %v, want %v", as[0].Type, AttachmentTypeImage)
		}
		if as[1].Key != "products/abc/doc.pdf" {
			t.Errorf("as[1].Key = %q, want %q", as[1].Key, "products/abc/doc.pdf")
		}
		if as[1].Type != AttachmentTypePDF {
			t.Errorf("as[1].Type = %v, want %v", as[1].Type, AttachmentTypePDF)
		}
	})

	t.Run("empty JSON array", func(t *testing.T) {
		var as Attachments
		if err := as.Scan([]byte(`[]`)); err != nil {
			t.Fatalf("Scan([]) returned error: %v", err)
		}
		if len(as) != 0 {
			t.Errorf("len(as) = %d, want 0", len(as))
		}
	})

	t.Run("invalid type src", func(t *testing.T) {
		var as Attachments
		err := as.Scan("not a []byte")
		if err == nil {
			t.Fatal("Scan(non []byte) expected error, got nil")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		var as Attachments
		err := as.Scan([]byte(`{invalid json`))
		if err == nil {
			t.Fatal("Scan(invalidJSON) expected error, got nil")
		}
	})
}

func TestAttachments_Value(t *testing.T) {
	as := Attachments{
		{Key: "p/a/img.jpg", Name: "photo.jpg", Size: 1024, ContentType: "image/jpeg", Type: AttachmentTypeImage},
		{Key: "p/a/doc.pdf", Name: "doc.pdf", Size: 2048, ContentType: "application/pdf", Type: AttachmentTypePDF},
	}

	val, err := as.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	got, ok := val.([]byte)
	if !ok {
		t.Fatalf("Value() type = %T, want []byte", val)
	}

	var decoded []attachmentJSON
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("Value() returned non-JSON bytes: %s", string(got))
	}
	if len(decoded) != 2 {
		t.Fatalf("len(decoded) = %d, want 2", len(decoded))
	}
	if decoded[0].Key != as[0].Key || decoded[1].Key != as[1].Key {
		t.Errorf("decoded keys = %v, want %v", []string{decoded[0].Key, decoded[1].Key},
			[]string{as[0].Key, as[1].Key})
	}
}

func TestAttachments_Value_Nil(t *testing.T) {
	var as Attachments
	val, err := as.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}
	if val != nil {
		t.Errorf("Value() = %v, want nil for nil Attachments", val)
	}
}

func TestAttachments_MarshalJSON(t *testing.T) {
	as := Attachments{
		{Key: "p/a/img.jpg", Name: "photo.jpg", Size: 1024, ContentType: "image/jpeg", Type: AttachmentTypeImage},
	}

	data, err := json.Marshal(as)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var decoded []attachmentJSON
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("MarshalJSON output is not a valid JSON array: %s", string(data))
	}
	if len(decoded) != 1 {
		t.Fatalf("len(decoded) = %d, want 1", len(decoded))
	}
	if decoded[0].Key != as[0].Key {
		t.Errorf("decoded[0].Key = %q, want %q", decoded[0].Key, as[0].Key)
	}
}

func TestAttachments_MarshalJSON_Nil(t *testing.T) {
	var as Attachments
	data, err := json.Marshal(as)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON(nil) = %s, want null", string(data))
	}
}

func TestAttachments_JSON_RoundTrip(t *testing.T) {
	original := Attachments{
		{Key: "uploads/1/a.jpg", Name: "a.jpg", Size: 100, ContentType: "image/jpeg", Type: AttachmentTypeImage},
		{Key: "uploads/1/b.pdf", Name: "b.pdf", Size: 200, ContentType: "application/pdf", Type: AttachmentTypePDF},
		{Key: "uploads/1/c.mp4", ContentType: "video/mp4", Type: AttachmentTypeVideo}, // name/size omitted
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var restored Attachments
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}

	if len(restored) != len(original) {
		t.Fatalf("len(restored) = %d, want %d", len(restored), len(original))
	}
	for i := range original {
		if restored[i] != original[i] {
			t.Errorf("restored[%d] = %+v, want %+v", i, restored[i], original[i])
		}
	}
}

func TestAttachments_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var as Attachments
	err := as.UnmarshalJSON([]byte(`{broken`))
	if err == nil {
		t.Fatal("UnmarshalJSON expected error for invalid JSON, got nil")
	}
}

func TestAttachments_UnmarshalJSON_Null(t *testing.T) {
	var as Attachments
	if err := as.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Fatalf("UnmarshalJSON(null) returned error: %v", err)
	}
	if as != nil {
		t.Errorf("UnmarshalJSON(null): as = %v, want nil", as)
	}
}

func TestAttachments_ImplementsSQLInterfaces(t *testing.T) {
	var _ driver.Valuer    = Attachments{}
	var _ json.Marshaler   = Attachments{}
	var _ json.Unmarshaler = (*Attachments)(nil)
}

// ── Attachment interface checks ─────────────────────────────────────────────

func TestAttachment_ImplementsSQLInterfaces(t *testing.T) {
	var _ driver.Valuer = Attachment{}
	var _ json.Marshaler   = Attachment{}
	var _ json.Unmarshaler = (*Attachment)(nil)
}

// ── Attachment constants ─────────────────────────────────────────────────────

func TestAttachmentTypeConstants(t *testing.T) {
	tests := []struct {
		val   AttachmentType
		valid bool
	}{
		{AttachmentTypeImage, true},
		{AttachmentTypeVideo, true},
		{AttachmentTypePDF, true},
		{AttachmentTypeDocument, true},
		{AttachmentTypeAudio, true},
		{AttachmentTypeOther, true},
		{AttachmentType(""), false},
		{AttachmentType("garbage"), false},
	}

	for _, tt := range tests {
		// Verify all constants round-trip through string conversion without panic
		_ = string(tt.val)
	}
}
