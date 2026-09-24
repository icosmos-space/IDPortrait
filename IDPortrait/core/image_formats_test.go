package core

import (
	"bytes"
	"encoding/base64"
	"image"
	"testing"
)

func TestSniffImageKind(t *testing.T) {
	cases := []struct {
		name string
		head []byte
		want string
	}{
		{"jpeg", []byte{0xff, 0xd8, 0xff, 0xe0}, "jpeg"},
		{"png", []byte{0x89, 'P', 'N', 'G', 0, 0, 0, 0, 0, 0, 0, 0}, "png"},
		{"webp", []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'}, "webp"},
		{"bmp", []byte{'B', 'M', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, "bmp"},
		{"heic", []byte{0, 0, 0, 0x18, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c'}, "heic"},
	}
	for _, tc := range cases {
		if got := sniffImageKind(tc.head); got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestDecodeRegisteredWebP(t *testing.T) {
	// Known-good tiny lossless WebP.
	webp1x1, err := base64.StdEncoding.DecodeString("UklGRhoAAABXRUJQVlA4TA0AAAAvAAAAEAcQERGIiP4HAA==")
	if err != nil {
		t.Fatal(err)
	}
	img, format, err := image.Decode(bytes.NewReader(webp1x1))
	if err != nil {
		t.Fatalf("webp decode: %v", err)
	}
	if format != "webp" {
		t.Fatalf("format %q", format)
	}
	if img.Bounds().Empty() {
		t.Fatalf("empty bounds")
	}
}
