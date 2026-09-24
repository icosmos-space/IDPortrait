package core

import (
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIDCardPhotoNotRejectedAsHat(t *testing.T) {
	path := filepath.Join("testdata", "idcard_no_hat.jpg")
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
		t.Fatalf("dark hair on printed ID card must not be hat reject, got %q", res.Reason)
	}
	if strings.Contains(res.Reason, "遮挡") {
		t.Fatalf("printed ID portrait must not be occlusion reject, got %q", res.Reason)
	}
	if !res.OK {
		t.Fatalf("expected pass, got OK=false reason=%q", res.Reason)
	}
}
