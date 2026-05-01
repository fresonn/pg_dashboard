package helper

import (
	"fmt"
)

// Units of measurement in bytes
const (
	_        = iota
	KB int64 = 1 << (10 * iota) // 1024
	MB                          // 1024²   = 1 048 576
	GB                          // 1024³   = 1 073 741 824
	TB                          // 1024⁴   = 1 099 511 627 776
)

func PrettyByteSize(bytes int64) string {
	if bytes <= 0 {
		return "0 Byte"
	}

	var value float64
	var unit string

	switch {
	case bytes >= TB:
		value = float64(bytes) / float64(TB)
		unit = "TB"
	case bytes >= GB:
		value = float64(bytes) / float64(GB)
		unit = "GB"
	case bytes >= MB:
		value = float64(bytes) / float64(MB)
		unit = "MB"
	case bytes >= KB:
		value = float64(bytes) / float64(KB)
		unit = "KB"
	default:
		return fmt.Sprintf("%d Byte", bytes)
	}

	return fmt.Sprintf("%.1f %s", value, unit)
}
