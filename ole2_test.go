package filetype

import (
	"encoding/binary"
	"testing"
)

// buildOLE2 constructs a minimal OLE2 file with the given directory entry stream names.
// It creates: header (512B) + 1 FAT sector (512B) + 1 directory sector (512B).
func buildOLE2(streamNames []string) []byte {
	const sectorSize = 512
	data := make([]byte, sectorSize*3)

	// --- Header ---
	copy(data[0:8], ole2Magic)
	binary.LittleEndian.PutUint16(data[28:30], 0xFFFE)
	binary.LittleEndian.PutUint16(data[30:32], 9)
	binary.LittleEndian.PutUint16(data[32:34], 6)
	binary.LittleEndian.PutUint32(data[44:48], 1)
	binary.LittleEndian.PutUint32(data[48:52], 1)
	binary.LittleEndian.PutUint32(data[76:80], 0)
	for i := 1; i < 109; i++ {
		binary.LittleEndian.PutUint32(data[76+i*4:80+i*4], ole2FreeSect)
	}

	// --- FAT sector (sector 0, offset 512) ---
	fatOffset := sectorSize
	binary.LittleEndian.PutUint32(data[fatOffset:fatOffset+4], 0xFFFFFFFD)
	binary.LittleEndian.PutUint32(data[fatOffset+4:fatOffset+8], ole2EndOfChain)
	for i := 2; i < sectorSize/4; i++ {
		binary.LittleEndian.PutUint32(data[fatOffset+i*4:fatOffset+i*4+4], ole2FreeSect)
	}

	// --- Directory sector (sector 1, offset 1024) ---
	dirOffset := sectorSize * 2
	writeOLE2DirEntry(data[dirOffset:dirOffset+128], "Root Entry", 5)
	for i, name := range streamNames {
		if i >= 3 {
			break
		}
		off := dirOffset + (i+1)*128
		writeOLE2DirEntry(data[off:off+128], name, 2)
	}

	return data
}

func writeOLE2DirEntry(entry []byte, name string, objType byte) {
	for i, c := range name {
		binary.LittleEndian.PutUint16(entry[i*2:i*2+2], uint16(c))
	}
	nameSize := (len(name) + 1) * 2
	binary.LittleEndian.PutUint16(entry[64:66], uint16(nameSize))
	entry[66] = objType
}

func TestDetectOLE2Type(t *testing.T) {
	tests := []struct {
		name    string
		streams []string
		want    string
	}{
		{"DOC file with WordDocument stream", []string{"WordDocument"}, "doc"},
		{"XLS file with Workbook stream", []string{"Workbook"}, "xls"},
		{"XLS file with Book stream (older format)", []string{"Book"}, "xls"},
		{"PPT file with PowerPoint Document stream", []string{"PowerPoint Document"}, "ppt"},
		{"Unknown OLE2 (no recognized streams)", []string{"Contents"}, ""},
		{"Multiple streams, WordDocument present", []string{"SummaryInformation", "WordDocument", "Data"}, "doc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildOLE2(tt.streams)
			if !IsOLE2(data) {
				t.Fatal("IsOLE2 should return true for constructed data")
			}
			got := DetectOLE2Type(data)
			if got != tt.want {
				t.Errorf("DetectOLE2Type() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsOLE2(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"valid magic", ole2Magic, true},
		{"too short", []byte{0xD0, 0xCF}, false},
		{"wrong magic", []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}, false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOLE2(tt.data); got != tt.want {
				t.Errorf("IsOLE2() = %v, want %v", got, tt.want)
			}
		})
	}
}
