package filetype

import (
	"bytes"
	"net/http"
	"strings"
)

// magicDef defines a magic-byte signature at a given offset.
type magicDef struct {
	offset int
	magic  []byte
	ext    string
}

// magicTable holds signatures that net/http.DetectContentType cannot identify.
var magicTable = []magicDef{
	// RTF: {\rtf
	{0, []byte{0x7B, 0x5C, 0x72, 0x74, 0x66}, "rtf"},
	// 7Z: 37 7A BC AF 27 1C
	{0, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}, "7z"},
	// XZ: FD 37 7A 58 5A 00
	{0, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}, "xz"},
	// TIFF Little-Endian: 49 49 2A 00
	{0, []byte{0x49, 0x49, 0x2A, 0x00}, "tif"},
	// TIFF Big-Endian: 4D 4D 00 2A
	{0, []byte{0x4D, 0x4D, 0x00, 0x2A}, "tif"},
	// PCX: 0A (with extra validation)
	{0, []byte{0x0A}, "pcx"},
}

// TAR magic at offset 257
var tarMagic = []byte("ustar")

// mimeToExt maps standard MIME types to file extensions.
var mimeToExt = map[string]string{
	"application/pdf":              "pdf",
	"application/zip":              "zip",
	"application/x-gzip":           "gz",
	"application/x-rar-compressed": "rar",
	"image/jpeg":                   "jpg",
	"image/png":                    "png",
	"image/gif":                    "gif",
	"image/bmp":                    "bmp",
	"image/webp":                   "webp",
	"text/html; charset=utf-8":     "html",
	"text/xml; charset=utf-8":      "xml",
}

// ole2Extensions are valid OLE2-based file extensions.
var ole2Extensions = map[string]bool{
	"doc": true, "xls": true, "ppt": true,
	"wps": true, "et": true, "dps": true,
}

// wpsExtensions are WPS Office proprietary extensions.
// WPS files (.wps/.et/.dps) use the same OLE2 stream names and OOXML directory
// structure as MS Office, so content detection alone cannot distinguish them.
// For these extensions, the filename must take priority.
var wpsExtensions = map[string]bool{
	"wps": true, "et": true, "dps": true,
}

// suffixToExt maps filename suffixes to canonical extensions.
var suffixToExt = map[string]string{
	".htm":   "html",
	".xhtml": "html",
	".tiff":  "tif",
	".jpeg":  "jpg",
	".tgz":   "gz",
}

// IsZip checks the PK magic bytes at the start of data.
func IsZip(data []byte) bool {
	return len(data) >= 4 &&
		data[0] == 'P' && data[1] == 'K' && data[2] == 3 && data[3] == 4
}

// DetectZipSubtype inspects the first 4096 bytes of a ZIP to identify
// OOXML (docx/xlsx/pptx), ODF (odt/ods/odp), or plain zip.
func DetectZipSubtype(data []byte) string {
	search := data
	if len(search) > 4096 {
		search = search[:4096]
	}
	s := string(search)

	// OOXML
	if strings.Contains(s, "word/") {
		return "docx"
	}
	if strings.Contains(s, "xl/") {
		return "xlsx"
	}
	if strings.Contains(s, "ppt/") {
		return "pptx"
	}

	// ODF
	if strings.Contains(s, "application/vnd.oasis.opendocument.text") {
		return "odt"
	}
	if strings.Contains(s, "application/vnd.oasis.opendocument.spreadsheet") {
		return "ods"
	}
	if strings.Contains(s, "application/vnd.oasis.opendocument.presentation") {
		return "odp"
	}

	return "zip"
}

func isValidPCX(data []byte) bool {
	return len(data) >= 3 && data[1] <= 5 && data[2] <= 1
}

// DetectByContent identifies file type purely from content bytes.
// Returns extension string (e.g. "pdf", "docx") or "" if unknown.
func DetectByContent(content []byte) string {
	if len(content) == 0 {
		return ""
	}

	// 1. ZIP family (OOXML / ODF / plain ZIP)
	if IsZip(content) {
		return DetectZipSubtype(content)
	}

	// 2. OLE2 Compound Document
	if IsOLE2(content) {
		if ft := DetectOLE2Type(content); ft != "" {
			return ft
		}
		return "ole2"
	}

	// 3. Custom magic bytes table
	for _, m := range magicTable {
		end := m.offset + len(m.magic)
		if len(content) > end && bytes.Equal(content[m.offset:end], m.magic) {
			if m.ext == "pcx" && !isValidPCX(content) {
				continue
			}
			return m.ext
		}
	}

	// 4. TAR (magic at offset 257)
	if len(content) >= 262 && bytes.Equal(content[257:262], tarMagic) {
		return "tar"
	}

	// 5. stdlib fallback
	mime := http.DetectContentType(content)
	if ext, ok := mimeToExt[mime]; ok {
		return ext
	}

	return ""
}

// Detect identifies file type using both filename and content.
// Content-based detection takes priority, with filename as fallback
// for ambiguous formats (e.g. WPS Office files).
func Detect(filename string, body []byte) string {
	if strings.HasSuffix(strings.ToLower(filename), ".txt") {
		return "txt"
	}

	// WPS Office formats have identical binary structure to MS Office.
	// When the filename indicates WPS and content is OLE2 or ZIP, trust the filename.
	if ext := extFromFilename(filename, wpsExtensions); ext != "" {
		if IsOLE2(body) || IsZip(body) {
			return ext
		}
	}

	ft := DetectByContent(body)

	// API response content may be misidentified as txt
	if ft == "txt" {
		return ""
	}

	// OLE2 with unrecognizable internal structure — fall back to filename
	if ft == "ole2" {
		if ext := extFromFilename(filename, ole2Extensions); ext != "" {
			return ext
		}
		return ""
	}

	if ft != "" {
		return ft
	}

	// Content unknown — try filename suffix fallback
	return suffixFallback(filename)
}

func extFromFilename(filename string, allowed map[string]bool) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		ext := strings.ToLower(filename[idx+1:])
		if allowed[ext] {
			return ext
		}
	}
	return ""
}

func suffixFallback(filename string) string {
	lower := strings.ToLower(filename)
	for suffix, ext := range suffixToExt {
		if strings.HasSuffix(lower, suffix) {
			return ext
		}
	}
	if strings.HasSuffix(lower, ".xml") {
		return "xml"
	}
	return ""
}
