package core

import (
	"image"
	"math"
)

// faceQualityReject inspects the face crop for Gaussian-like blur and mosaic
// (pixelation). Empty string means the face looks sharp enough for an ID photo.
func faceQualityReject(img *image.NRGBA, box []float64) string {
	if img == nil || len(box) < 4 {
		return ""
	}
	crop := cropFaceGray(img, box[0], box[1], box[2], box[3], 0.18)
	if crop == nil || crop.w < 24 || crop.h < 24 {
		return ""
	}

	// Mosaic check must NOT upscale: printed ID-card portraits are small and
	// look blocky after enlargement, which falsely trips pixelation detectors.
	mosaicSrc := crop
	if crop.w > 192 || crop.h > 192 {
		mosaicSrc = resizeGrayCropMax(crop, 192)
	}
	if isMosaicFace(mosaicSrc) {
		return "检测到人脸马赛克，已拒绝"
	}

	norm := resizeGrayCrop(crop, 160)
	sharp := laplacianVariance(norm)
	// Printed ID-card portraits are softer than digital captures; 90 was
	// rejecting them. Intentional heavy blur still lands well below 70.
	if sharp < 70 {
		return "检测到人脸模糊，已拒绝"
	}
	return ""
}

type grayCrop struct {
	w, h int
	pix  []float64 // 0..255
}

func cropFaceGray(img *image.NRGBA, x1, y1, x2, y2, padRatio float64) *grayCrop {
	bw := img.Bounds().Dx()
	bh := img.Bounds().Dy()
	if bw < 2 || bh < 2 {
		return nil
	}
	fw := x2 - x1
	fh := y2 - y1
	if fw < 4 || fh < 4 {
		return nil
	}
	padX := fw * padRatio
	padY := fh * padRatio
	l := int(math.Floor(x1 - padX))
	t := int(math.Floor(y1 - padY))
	r := int(math.Ceil(x2 + padX))
	b := int(math.Ceil(y2 + padY))
	if l < 0 {
		l = 0
	}
	if t < 0 {
		t = 0
	}
	if r > bw {
		r = bw
	}
	if b > bh {
		b = bh
	}
	w := r - l
	h := b - t
	if w < 8 || h < 8 {
		return nil
	}
	pix := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := img.PixOffset(l+x, t+y)
			pix[y*w+x] = 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
		}
	}
	return &grayCrop{w: w, h: h, pix: pix}
}

func resizeGrayCrop(src *grayCrop, side int) *grayCrop {
	if src == nil || side < 8 {
		return src
	}
	return resizeGrayExact(src, side, side)
}

func resizeGrayCropMax(src *grayCrop, maxSide int) *grayCrop {
	if src == nil {
		return nil
	}
	m := src.w
	if src.h > m {
		m = src.h
	}
	if m <= maxSide {
		return src
	}
	nw := src.w * maxSide / m
	nh := src.h * maxSide / m
	if nw < 8 {
		nw = 8
	}
	if nh < 8 {
		nh = 8
	}
	return resizeGrayExact(src, nw, nh)
}

func resizeGrayExact(src *grayCrop, nw, nh int) *grayCrop {
	dst := &grayCrop{w: nw, h: nh, pix: make([]float64, nw*nh)}
	scaleX := float64(src.w) / float64(nw)
	scaleY := float64(src.h) / float64(nh)
	for y := 0; y < nh; y++ {
		y0 := int(float64(y) * scaleY)
		y1 := int(float64(y+1) * scaleY)
		if y1 <= y0 {
			y1 = y0 + 1
		}
		if y1 > src.h {
			y1 = src.h
		}
		for x := 0; x < nw; x++ {
			x0 := int(float64(x) * scaleX)
			x1 := int(float64(x+1) * scaleX)
			if x1 <= x0 {
				x1 = x0 + 1
			}
			if x1 > src.w {
				x1 = src.w
			}
			var sum float64
			n := 0
			for yy := y0; yy < y1; yy++ {
				for xx := x0; xx < x1; xx++ {
					sum += src.pix[yy*src.w+xx]
					n++
				}
			}
			if n > 0 {
				dst.pix[y*nw+x] = sum / float64(n)
			}
		}
	}
	return dst
}

func laplacianVariance(g *grayCrop) float64 {
	if g == nil || g.w < 3 || g.h < 3 {
		return 0
	}
	var sum, sumSq float64
	n := 0
	for y := 1; y < g.h-1; y++ {
		for x := 1; x < g.w-1; x++ {
			c := g.pix[y*g.w+x]
			v := g.pix[(y-1)*g.w+x] + g.pix[(y+1)*g.w+x] + g.pix[y*g.w+x-1] + g.pix[y*g.w+x+1] - 4*c
			sum += v
			sumSq += v * v
			n++
		}
	}
	if n < 1 {
		return 0
	}
	mean := sum / float64(n)
	return sumSq/float64(n) - mean*mean
}

// isMosaicFace requires flat tiles (low within-block variance), strong tile
// jumps, and a near-perfect block reconstruction — typical of intentional
// pixelation, not of small printed portraits on ID cards.
func isMosaicFace(g *grayCrop) bool {
	if g == nil {
		return false
	}
	for _, bs := range []int{8, 10, 12, 16, 20, 24} {
		fit, jump, within := mosaicStatsAt(g, bs)
		if fit >= 0.94 && jump >= 10 && within < 12 {
			return true
		}
		if fit >= 0.97 && jump >= 8 && within < 20 {
			return true
		}
	}
	return false
}

func mosaicStatsAt(g *grayCrop, bs int) (fit, jump, within float64) {
	if g == nil || bs < 2 || g.w < bs*3 || g.h < bs*3 {
		return 0, 0, 0
	}
	cols := g.w / bs
	rows := g.h / bs
	means := make([]float64, rows*cols)

	var sum, sumSq float64
	nPix := 0
	var mse float64
	var withinSum float64
	wn := 0
	for by := 0; by < rows; by++ {
		for bx := 0; bx < cols; bx++ {
			var s, s2 float64
			n := 0
			for y := 0; y < bs; y++ {
				for x := 0; x < bs; x++ {
					v := g.pix[(by*bs+y)*g.w+(bx*bs+x)]
					s += v
					s2 += v * v
					sum += v
					sumSq += v * v
					n++
					nPix++
				}
			}
			m := s / float64(n)
			means[by*cols+bx] = m
			blockVar := s2/float64(n) - m*m
			withinSum += blockVar
			wn++
			for y := 0; y < bs; y++ {
				for x := 0; x < bs; x++ {
					v := g.pix[(by*bs+y)*g.w+(bx*bs+x)]
					d := v - m
					mse += d * d
				}
			}
		}
	}
	if nPix < 1 || wn < 1 {
		return 0, 0, 0
	}
	mean := sum / float64(nPix)
	variance := sumSq/float64(nPix) - mean*mean
	if variance < 1 {
		return 0, 0, 0
	}
	fit = 1 - (mse/float64(nPix))/variance
	if fit < 0 {
		fit = 0
	}
	within = withinSum / float64(wn)

	var between float64
	bn := 0
	for by := 0; by < rows; by++ {
		for bx := 0; bx < cols; bx++ {
			m := means[by*cols+bx]
			if bx+1 < cols {
				between += math.Abs(m - means[by*cols+bx+1])
				bn++
			}
			if by+1 < rows {
				between += math.Abs(m - means[(by+1)*cols+bx])
				bn++
			}
		}
	}
	if bn > 0 {
		jump = between / float64(bn)
	}
	return fit, jump, within
}

func faceQualityMetrics(img *image.NRGBA) (sharp, fit, jump float64) {
	if img == nil {
		return 0, 0, 0
	}
	b := img.Bounds()
	crop := cropFaceGray(img, 0, 0, float64(b.Dx()), float64(b.Dy()), 0)
	norm := resizeGrayCrop(crop, 160)
	sharp = laplacianVariance(norm)
	src := crop
	if crop.w > 192 || crop.h > 192 {
		src = resizeGrayCropMax(crop, 192)
	}
	for _, bs := range []int{8, 10, 12, 16, 20, 24} {
		f, j, _ := mosaicStatsAt(src, bs)
		if f > fit {
			fit, jump = f, j
		}
	}
	return sharp, fit, jump
}
