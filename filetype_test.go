package filetype

import (
	"testing"
)

func TestDetect_OLE2Integration(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		streams  []string
		want     string
	}{
		{
			name:     "DOC detected by content, filename ignored",
			filename: "report.xls",
			streams:  []string{"WordDocument"},
			want:     "doc",
		},
		{
			name:     "WPS with WordDocument stream → filename wins",
			filename: "document.wps",
			streams:  []string{"WordDocument"},
			want:     "wps",
		},
		{
			name:     "ET with Workbook stream → filename wins",
			filename: "data.et",
			streams:  []string{"Workbook"},
			want:     "et",
		},
		{
			name:     "DPS with PowerPoint Document stream → filename wins",
			filename: "slides.dps",
			streams:  []string{"PowerPoint Document"},
			want:     "dps",
		},
		{
			name:     "WPS with unrecognized stream → filename wins",
			filename: "document.wps",
			streams:  []string{"Contents"},
			want:     "wps",
		},
		{
			name:     "ET with unrecognized stream → filename wins",
			filename: "data.et",
			streams:  []string{"Contents"},
			want:     "et",
		},
		{
			name:     "DPS with unrecognized stream → filename wins",
			filename: "slides.dps",
			streams:  []string{"Contents"},
			want:     "dps",
		},
		{
			name:     "Unknown OLE2 with unknown extension",
			filename: "file.bin",
			streams:  []string{"Contents"},
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildOLE2(tt.streams)
			got := Detect(tt.filename, data)
			if got != tt.want {
				t.Errorf("Detect(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetect_TxtFile(t *testing.T) {
	got := Detect("readme.txt", []byte("hello world"))
	if got != "txt" {
		t.Errorf("Detect(readme.txt) = %q, want %q", got, "txt")
	}
}

func TestDetect_SuffixFallback(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"page.htm", "html"},
		{"page.xhtml", "html"},
		{"photo.jpeg", "jpg"},
		{"image.tiff", "tif"},
		{"archive.tgz", "gz"},
		{"data.xml", "xml"},
		{"unknown.bin", ""},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := Detect(tt.filename, nil)
			if got != tt.want {
				t.Errorf("Detect(%q, nil) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectByContent_ZIP(t *testing.T) {
	// Minimal ZIP with "word/" in the local file header
	data := []byte{'P', 'K', 3, 4}
	data = append(data, make([]byte, 26)...) // padding to filename
	data = append(data, []byte("word/document.xml")...)
	got := DetectByContent(data)
	if got != "docx" {
		t.Errorf("DetectByContent(zip with word/) = %q, want %q", got, "docx")
	}
}

func TestDetectByContent_MagicBytes(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    string
	}{
		{"RTF", []byte{0x7B, 0x5C, 0x72, 0x74, 0x66, 0x31}, "rtf"},
		{"7Z", []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C, 0x00}, "7z"},
		{"empty", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectByContent(tt.content)
			if got != tt.want {
				t.Errorf("DetectByContent() = %q, want %q", got, tt.want)
			}
		})
	}
}
