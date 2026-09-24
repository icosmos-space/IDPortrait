package core

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHighBunNotRejectedAsHat(t *testing.T) {
	path := filepath.Join("testdata", "high_bun_no_hat.png")
	f, err := os.Open(path)
	if err != nil {
		t.Skip(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	url, err := encodeJPEGDataURL(toNRGBA(downscale(img, 1600)), 90)
	if err != nil {
		t.Fatal(err)
	}
	res, err := DetectFace(url, FaceCheckOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.Reason, "帽子") {
		t.Fatalf("high bun must not be hat reject, got %q", res.Reason)
	}
	if strings.Contains(res.Reason, "肩膀") {
		t.Fatalf("frontal cream shirt + side locks must not be shoulder reject, got %q", res.Reason)
	}
	if !res.OK {
		t.Fatalf("expected pass, got OK=false reason=%q", res.Reason)
	}
}
