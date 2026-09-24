package core

import (
	// Register codecs for image.Decode used by LoadImage / data URLs.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// sniffImageKind returns a short label for common magic headers.
func sniffImageKind(b []byte) string {
	if len(b) < 3 {
		return ""
	}
	switch {
	case b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff:
		return "jpeg"
	case len(b) >= 4 && b[0] == 0x89 && b[1] == 'P' && b[2] == 'N' && b[3] == 'G':
		return "png"
	case b[0] == 'G' && b[1] == 'I' && b[2] == 'F':
		return "gif"
	case len(b) >= 2 && b[0] == 'B' && b[1] == 'M':
		return "bmp"
	case len(b) >= 12 && b[0] == 'R' && b[1] == 'I' && b[2] == 'F' && b[3] == 'F' &&
		b[8] == 'W' && b[9] == 'E' && b[10] == 'B' && b[11] == 'P':
		return "webp"
	case len(b) >= 12 && string(b[4:8]) == "ftyp":
		brand := string(b[8:12])
		switch brand {
		case "heic", "heix", "hevc", "hevx", "mif1", "msf1", "avif":
			return brand
		default:
			return "heif/" + brand
		}
	case len(b) >= 4 && b[0] == 'I' && b[1] == 'I' && b[2] == '*' && b[3] == 0:
		return "tiff"
	case len(b) >= 4 && b[0] == 'M' && b[1] == 'M' && b[2] == 0 && b[3] == '*':
		return "tiff"
	default:
		return ""
	}
}

func decodeFormatHint(b []byte) string {
	kind := sniffImageKind(b)
	switch kind {
	case "":
		return "（无法识别文件头，可能已损坏或不是图片）"
	case "heic", "heix", "hevc", "hevx", "mif1", "msf1", "heif", "avif":
		return "（检测到 " + kind + "，请先转为 JPG/PNG/WebP 后再导入）"
	default:
		return "（检测到 " + kind + "）"
	}
}
