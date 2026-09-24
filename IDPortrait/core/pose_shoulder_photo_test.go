package core

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShoulderTwistPhotoRejected(t *testing.T) {
	path := filepath.Join("testdata", "shoulder_twist.png")
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
	if res.OK {
		t.Fatalf("frontal face + twisted shoulders must reject, got OK")
	}
	if !strings.Contains(res.Reason, "肩膀") && !strings.Contains(res.Reason, "倾斜") {
		t.Fatalf("expected shoulder reject, got %q", res.Reason)
	}
}
