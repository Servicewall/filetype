# filetype

A Go library for detecting file types by content (magic bytes) and filename. Zero dependencies outside the standard library.

## Features

- Detects 30+ file formats from binary content alone
- Parses OLE2 (Compound Binary File) internal directory to distinguish `.doc` / `.xls` / `.ppt`
- Identifies ZIP-based subtypes: OOXML (docx/xlsx/pptx), ODF (odt/ods/odp)
- Supports WPS Office formats (wps/et/dps) via filename hint
- Falls back to `net/http.DetectContentType` for common MIME types
- Filename suffix fallback for ambiguous or content-undetectable formats

## Installation

```bash
go get github.com/servicewall/filetype
```

## Usage

```go
import "github.com/servicewall/filetype"

// Detect using both filename and content (recommended)
ext := filetype.Detect("report.docx", fileBytes)
// ext == "docx"

// Detect using content only
ext := filetype.DetectByContent(fileBytes)
// ext == "pdf"

// Check specific format
if filetype.IsZip(fileBytes) {
    subtype := filetype.DetectZipSubtype(fileBytes)
}

if filetype.IsOLE2(fileBytes) {
    subtype := filetype.DetectOLE2Type(fileBytes)
}
```

## Supported Formats

| Category | Extensions |
|----------|-----------|
| Office (OOXML) | docx, xlsx, pptx |
| Office (OLE2) | doc, xls, ppt |
| Office (ODF) | odt, ods, odp |
| WPS Office | wps, et, dps |
| Documents | pdf, rtf |
| Images | jpg, png, gif, bmp, webp, tif, pcx |
| Archives | zip, gz, tar, rar, 7z, xz |
| Markup | html, xml |
| Text | txt |

## API

### `Detect(filename string, body []byte) string`

Primary detection function. Uses content-based detection with filename as fallback for ambiguous formats (WPS Office, unrecognized OLE2). Returns the file extension (e.g. `"pdf"`, `"docx"`) or `""` if unknown.

### `DetectByContent(content []byte) string`

Detects file type purely from content bytes, without any filename hint.

### `IsZip(data []byte) bool`

Checks if data starts with the PK (ZIP) magic bytes.

### `DetectZipSubtype(data []byte) string`

Inspects ZIP content to identify OOXML, ODF, or plain zip.

### `IsOLE2(data []byte) bool`

Checks if data starts with the OLE2 magic signature.

### `DetectOLE2Type(data []byte) string`

Parses OLE2 directory entries to identify doc/xls/ppt.

## Detection Priority

1. `.txt` extension short-circuits to `"txt"`
2. WPS Office filename extensions (wps/et/dps) with matching OLE2/ZIP content
3. ZIP magic bytes -> OOXML / ODF / plain zip
4. OLE2 magic bytes -> parse directory for doc/xls/ppt
5. Custom magic byte table (RTF, 7z, XZ, TIFF, PCX)
6. TAR magic at offset 257
7. `net/http.DetectContentType` MIME-based detection
8. Filename suffix fallback (htm, xhtml, jpeg, tiff, tgz, xml)

## License

MIT
