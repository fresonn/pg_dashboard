package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyByteSize(t *testing.T) {

	t.Parallel()

	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero", 0, "0 Byte"},
		{"negative", -1, "0 Byte"},
		{"1 byte", 1, "1 Byte"},
		{"999 bytes", 999, "999 Byte"},
		{"1023 bytes", 1023, "1023 Byte"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"just below 1 MB", 1048575, "1024.0 KB"},
		{"exactly 1 MB", 1048576, "1.0 MB"},
		{"2.3 MB", 2*1048576 + 300*1024, "2.3 MB"},
		{"just below 1 GB", 1073741823, "1024.0 MB"},
		{"exactly 1 GB", 1073741824, "1.0 GB"},
		{"just below 1 TB", 1099511627775, "1024.0 GB"},
		{"exactly 1 TB", 1099511627776, "1.0 TB"},
		{"big value", 15*1099511627776 + 600*1073741824, "15.6 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrettyByteSize(tt.bytes)
			assert.Equal(t, tt.expected, result, "PrettyByteSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
		})
	}
}
