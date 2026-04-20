package filetype

import (
	"encoding/binary"
	"unicode/utf16"
)

// OLE2 Compound Binary File Format
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cfb
//
// Header (512 bytes):
//   offset 0:  magic (8 bytes)  D0 CF 11 E0 A1 B1 1A E1
//   offset 30: sector size power (2 bytes), v3=9(512B) v4=12(4096B)
//   offset 44: number of FAT sectors (4 bytes)
//   offset 48: first directory sector SID (4 bytes)
//   offset 76: DIFAT array, 109 entries × 4 bytes
//
// Directory Entry (128 bytes):
//   offset 0:  name in UTF-16LE (64 bytes)
//   offset 64: name size in bytes including null terminator (2 bytes)
//   offset 66: object type (1 byte)
//
// Detection strategy:
//   Scan directory entries for well-known stream names to identify the file type.
//   - "WordDocument"           → doc
//   - "Workbook" / "Book"      → xls
//   - "PowerPoint Document"    → ppt
//   - "HwpSummaryInformation"  → hwp (Hangul)
//   If none matched, return "" so the caller can fall back to filename extension
//   (covers WPS-specific formats: wps, et, dps).

const (
	ole2HeaderSize       = 512
	ole2DirEntrySize     = 128
	ole2EndOfChain       = 0xFFFFFFFE
	ole2FreeSect         = 0xFFFFFFFF
	ole2MaxNameBytes     = 64
	ole2MaxDirSectors    = 16
	ole2HeaderDIFATCount = 109
)

var ole2Magic = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}

type ole2Header struct {
	sectorSize  uint32
	dirFirstSID uint32
	fatSectors  []uint32
}

// IsOLE2 checks the magic bytes at the start of data.
func IsOLE2(data []byte) bool {
	if len(data) < len(ole2Magic) {
		return false
	}
	for i, b := range ole2Magic {
		if data[i] != b {
			return false
		}
	}
	return true
}

// DetectOLE2Type parses the OLE2 directory to identify the file type.
// Returns "doc", "xls", "ppt", etc. or "" if unknown.
func DetectOLE2Type(data []byte) string {
	h, ok := parseOLE2Header(data)
	if !ok {
		return ""
	}

	sid := h.dirFirstSID
	for i := 0; i < ole2MaxDirSectors && sid != ole2EndOfChain && sid != ole2FreeSect; i++ {
		if ft := scanDirSector(data, h, sid); ft != "" {
			return ft
		}
		next, ok := readFATEntry(data, h, sid)
		if !ok {
			break
		}
		sid = next
	}
	return ""
}

func parseOLE2Header(data []byte) (ole2Header, bool) {
	var h ole2Header
	if len(data) < ole2HeaderSize || !IsOLE2(data) {
		return h, false
	}

	ssz := binary.LittleEndian.Uint16(data[30:32])
	if ssz < 7 || ssz > 15 {
		return h, false
	}
	h.sectorSize = 1 << ssz
	h.dirFirstSID = binary.LittleEndian.Uint32(data[48:52])

	numFAT := binary.LittleEndian.Uint32(data[44:48])
	count := min(numFAT, ole2HeaderDIFATCount)
	h.fatSectors = make([]uint32, 0, count)
	for i := range count {
		sid := binary.LittleEndian.Uint32(data[76+i*4 : 80+i*4])
		if sid != ole2FreeSect && sid != ole2EndOfChain {
			h.fatSectors = append(h.fatSectors, sid)
		}
	}
	return h, true
}

func sectorOffset(sid, sectorSize uint32) int64 {
	return int64(sid+1) * int64(sectorSize)
}

func readFATEntry(data []byte, h ole2Header, sid uint32) (uint32, bool) {
	entriesPerSector := h.sectorSize / 4
	fatIndex := sid / entriesPerSector
	fatEntry := sid % entriesPerSector

	if int(fatIndex) >= len(h.fatSectors) {
		return 0, false
	}

	offset := sectorOffset(h.fatSectors[fatIndex], h.sectorSize) + int64(fatEntry)*4
	if offset+4 > int64(len(data)) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(data[offset : offset+4]), true
}

func scanDirSector(data []byte, h ole2Header, sid uint32) string {
	offset := sectorOffset(sid, h.sectorSize)
	end := offset + int64(h.sectorSize)
	if offset < 0 || end > int64(len(data)) {
		return ""
	}

	sector := data[offset:end]
	entriesPerSector := h.sectorSize / ole2DirEntrySize

	for i := range entriesPerSector {
		entry := sector[i*ole2DirEntrySize : (i+1)*ole2DirEntrySize]
		nameSize := binary.LittleEndian.Uint16(entry[64:66])
		if nameSize == 0 || nameSize > ole2MaxNameBytes {
			continue
		}
		name := decodeUTF16LE(entry[:nameSize])
		if ft := matchOLE2Stream(name); ft != "" {
			return ft
		}
	}
	return ""
}

func matchOLE2Stream(name string) string {
	switch name {
	case "WordDocument":
		return "doc"
	case "Workbook", "Book":
		return "xls"
	case "PowerPoint Document":
		return "ppt"
	default:
		return ""
	}
}

func decodeUTF16LE(b []byte) string {
	u16 := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		c := binary.LittleEndian.Uint16(b[i : i+2])
		if c == 0 {
			break
		}
		u16 = append(u16, c)
	}
	return string(utf16.Decode(u16))
}
