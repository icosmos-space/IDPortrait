package core

import (
	"image"
	"image/color"
	"math"
	"strings"
	"testing"
)

func TestFaceQualityRejectsBlur(t *testing.T) {
	sharp := synthFaceTexture(160, false, 0)
	blur := boxBlurNRGBA(sharp, 10)
	box := []float64{20, 20, 140, 140}
	if reason := faceQualityReject(sharp, box); reason != "" {
		t.Fatalf("sharp should pass, got %q", reason)
	}
	reason := faceQualityReject(blur, box)
	if !strings.Contains(reason, "模糊") {
		t.Fatalf("blur should reject, got %q (sharp=%v blur=%v)", reason, mustSharp(sharp), mustSharp(blur))
	}
}

func TestFaceQualityRejectsMosaic(t *testing.T) {
	sharp := synthFaceTexture(160, false, 0)
	mosaic := pixelateNRGBA(sharp, 12)
	box := []float64{20, 20, 140, 140}
	reason := faceQualityReject(mosaic, box)
	if !strings.Contains(reason, "马赛克") && !strings.Contains(reason, "模糊") {
		fit, jump := mustMosaic(mosaic)
		t.Fatalf("mosaic should reject, got %q (sharp=%v fit=%v jump=%v)", reason, mustSharp(mosaic), fit, jump)
	}
}

func mustMosaic(img *image.NRGBA) (fit, jump float64) {
	_, fit, jump = faceQualityMetrics(img)
	return fit, jump
}

func mustSharp(img *image.NRGBA) float64 {
	s, _, _ := faceQualityMetrics(img)
	return s
}

func synthFaceTexture(side int, flat bool, seed int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, side, side))
	for y := 0; y < side; y++ {
		for x := 0; x < side; x++ {
			v := 110.0
			if !flat {
				v += 40 * math.Sin(float64(x)*0.35+float64(seed)) * math.Cos(float64(y)*0.28)
				v += 25 * math.Sin(float64(x+y)*0.55)
				if (x+y+seed)%7 == 0 {
					v += 30
				}
			}
			if v < 0 {
				v = 0
			}
			if v > 255 {
				v = 255
			}
			c := uint8(v)
			img.SetNRGBA(x, y, color.NRGBA{R: c, G: c, B: uint8(math.Min(255, v+8)), A: 255})
		}
	}
	return img
}

func boxBlurNRGBA(src *image.NRGBA, radius int) *image.NRGBA {
	if radius < 1 {
		return src
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	tmp := image.NewNRGBA(src.Bounds())
	copy(tmp.Pix, src.Pix)
	for pass := 0; pass < radius; pass++ {
		next := image.NewNRGBA(src.Bounds())
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				var r, g, b, n float64
				for dy := -1; dy <= 1; dy++ {
					yy := y + dy
					if yy < 0 || yy >= h {
						continue
					}
					for dx := -1; dx <= 1; dx++ {
						xx := x + dx
						if xx < 0 || xx >= w {
							continue
						}
						i := tmp.PixOffset(xx, yy)
						r += float64(tmp.Pix[i])
						g += float64(tmp.Pix[i+1])
						b += float64(tmp.Pix[i+2])
						n++
					}
				}
				i := next.PixOffset(x, y)
				next.Pix[i] = uint8(r / n)
				next.Pix[i+1] = uint8(g / n)
				next.Pix[i+2] = uint8(b / n)
				next.Pix[i+3] = 255
			}
		}
		tmp = next
	}
	return tmp
}

func pixelateNRGBA(src *image.NRGBA, block int) *image.NRGBA {
	if block < 2 {
		return src
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(src.Bounds())
	for y := 0; y < h; y += block {
		for x := 0; x < w; x += block {
			var r, g, b, n float64
			for yy := y; yy < y+block && yy < h; yy++ {
				for xx := x; xx < x+block && xx < w; xx++ {
					i := src.PixOffset(xx, yy)
					r += float64(src.Pix[i])
					g += float64(src.Pix[i+1])
					b += float64(src.Pix[i+2])
					n++
				}
			}
			cr, cg, cb := uint8(r/n), uint8(g/n), uint8(b/n)
			for yy := y; yy < y+block && yy < h; yy++ {
				for xx := x; xx < x+block && xx < w; xx++ {
					i := dst.PixOffset(xx, yy)
					dst.Pix[i] = cr
					dst.Pix[i+1] = cg
					dst.Pix[i+2] = cb
					dst.Pix[i+3] = 255
				}
			}
		}
	}
	return dst
}
