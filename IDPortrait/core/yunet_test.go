package core

import (
	"image"
	"testing"
)

func TestDetectFaceRejectsBlank(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 480, 640))
	url, err := encodeJPEGDataURL(img, 80)
	if err != nil {
		t.Fatal(err)
	}
	res, err := DetectFace(url)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Fatalf("blank image should be rejected, got %+v", res)
	}
	if res.Reason == "" {
		t.Fatal("expected a reject reason")
	}
}
