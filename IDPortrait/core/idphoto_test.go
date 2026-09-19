package core

import (
	"image"
	"testing"
)

func TestAdjustIDPhotoSize(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 800, 1000))
	for i := 3; i < len(src.Pix); i += 4 {
		src.Pix[i] = 255
	}
	face := faceHit{x: 280, y: 180, w: 240, h: 280, score: 0.9}
	out := adjustIDPhoto(src, face, 295, 413, 0.7, 0.07)
	if out.Bounds().Dx() != 295 || out.Bounds().Dy() != 413 {
		t.Fatalf("size %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}
