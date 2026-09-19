package core

import (
	"image"
	"testing"
)

func TestSocialTemplatesSize(t *testing.T) {
	photo := image.NewRGBA(image.Rect(0, 0, 295, 413))
	for i := 0; i < len(photo.Pix); i += 4 {
		photo.Pix[i] = 40
		photo.Pix[i+1] = 90
		photo.Pix[i+2] = 160
		photo.Pix[i+3] = 255
	}
	a, b, err := socialTemplates(photo)
	if err != nil {
		t.Fatal(err)
	}
	if a.Bounds().Dx() != 1080 || a.Bounds().Dy() != 1400 {
		t.Fatalf("template1 %dx%d", a.Bounds().Dx(), a.Bounds().Dy())
	}
	if b.Bounds().Dx() != 1080 || b.Bounds().Dy() != 1440 {
		t.Fatalf("template2 %dx%d", b.Bounds().Dx(), b.Bounds().Dy())
	}
}

func TestStripedWatermark(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 320, 240))
	for i := 0; i < len(src.Pix); i += 4 {
		src.Pix[i] = 20
		src.Pix[i+1] = 20
		src.Pix[i+2] = 20
		src.Pix[i+3] = 255
	}
	out, err := paintWatermark(src, GenerateParams{
		EnableWatermark:   true,
		WatermarkText:     "最美证件照",
		WatermarkColor:    "#FFFFFF",
		WatermarkFontSize: 28,
		WatermarkOpacity:  0.8,
		WatermarkAngle:    -30,
		WatermarkSpacing:  80,
	})
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for i := 0; i < len(src.Pix); i += 4 {
		if out.Pix[i] != src.Pix[i] || out.Pix[i+1] != src.Pix[i+1] {
			changed++
		}
	}
	if changed < 20 {
		t.Fatalf("watermark barely visible, changed pixels %d", changed)
	}
}
