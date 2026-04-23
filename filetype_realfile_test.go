package filetype

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_RealFiles(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		// OOXML (ZIP-based)
		{"docx.docx", "docx"},
		{"xlsx.xlsx", "xlsx"},
		{"pptx.pptx", "pptx"},

		// ODF (ZIP-based)
		{"odt.odt", "odt"},
		{"ods.ods", "ods"},
		{"odp.odp", "odp"},

		// OLE2 (Compound Binary)
		{"doc.doc", "doc"},
		{"test_xls.xls", "xls"},
		{"ppt.ppt", "ppt"},

		// WPS Office
		{"测试.wps", "wps"},
		{"工作簿1.et", "et"},
		{"pptx.dps", "dps"},

		// Rich Text
		{"rtf.rtf", "rtf"},
		{"test.pdf", "pdf"},

		// Text / Markup
		{"txt.txt", "txt"},
		{"xml.xml", "xml"},
		{"html.html", "html"},

		// Images
		{"jpg.jpg", "jpg"},
		{"png.png", "png"},
		{"gif.gif", "gif"},
		{"bmp.bmp", "bmp"},
		{"tif.tif", "tif"},
		{"webp.webp", "webp"},
		{"sample.webp", "webp"},
		{"pcx.pcx", "pcx"},

		// Archives
		{"zip.zip", "zip"},
		{"test.7z", "7z"},
		{"gz.tar.gz", "gz"},
		{"test.rar", "rar"},
		{"tar.tar", "tar"},
		{"xz.txt.xz", "xz"},
	}

	for _, tt := range tests {
		t.Run(tt.filename+"→"+tt.want, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("failed to read testdata/%s: %v", tt.filename, err)
			}
			got := Detect(tt.filename, data)
			if got != tt.want {
				t.Errorf("Detect(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectByContent_RealFiles(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		// Content-only detection (no filename hint)
		{"docx.docx", "docx"},
		{"xlsx.xlsx", "xlsx"},
		{"pptx.pptx", "pptx"},
		{"odt.odt", "odt"},
		{"ods.ods", "ods"},
		{"odp.odp", "odp"},
		{"doc.doc", "doc"},
		{"ppt.ppt", "ppt"},
		{"rtf.rtf", "rtf"},
		{"jpg.jpg", "jpg"},
		{"png.png", "png"},
		{"gif.gif", "gif"},
		{"bmp.bmp", "bmp"},
		{"tif.tif", "tif"},
		{"webp.webp", "webp"},
		{"zip.zip", "zip"},
		{"test_xls.xls", "xls"},
		{"test.7z", "7z"},
		{"gz.tar.gz", "gz"},
		{"test.rar", "rar"},
		{"tar.tar", "tar"},
		{"xz.txt.xz", "xz"},
		{"pcx.pcx", "pcx"},
		{"html.html", "html"},
		{"xml.xml", "xml"},
		{"test.pdf", "pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.filename+"→"+tt.want, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("failed to read testdata/%s: %v", tt.filename, err)
			}
			got := DetectByContent(data)
			if got != tt.want {
				t.Errorf("DetectByContent(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
