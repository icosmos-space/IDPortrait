package core

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClearGlassesNotRejectedAsSunglasses(t *testing.T) {
	path := filepath.Join("testdata", "clear_glasses.png")
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
	if strings.Contains(res.Reason, "墨镜") || strings.Contains(res.Reason, "闭眼") {
		t.Fatalf("clear eyeglasses must not be sunglasses/closed-eye, got %q", res.Reason)
	}
	if !res.OK {
		t.Fatalf("expected pass, got OK=false reason=%q", res.Reason)
	}
}
