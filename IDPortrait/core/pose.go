package core

import (
	"image"
	"math"
)

// facePoseReject returns a Chinese reject reason for yaw/pitch only
// (side face / looking down / looking up). Head roll is faceRollReject.
//
// YuNet landmark order: right eye, left eye, nose tip, right mouth, left mouth.
func facePoseReject(f faceHit) string {
	rex, rey := f.kps[0], f.kps[1]
	lex, ley := f.kps[2], f.kps[3]
	nx, ny := f.kps[4], f.kps[5]
	_, rmy := f.kps[6], f.kps[7]
	_, lmy := f.kps[8], f.kps[9]

	eyeDist := math.Hypot(rex-lex, rey-ley)
	if eyeDist < 1 {
		return "人脸关键点无效，已拒绝"
	}

	// Roll is checked separately (faceRollReject) so settings can toggle it apart
	// from yaw/pitch.

	eyeMidX := (rex + lex) / 2
	eyeMidY := (rey + ley) / 2
	// Yaw: nose should sit near the midpoint between the eyes.
	if math.Abs(nx-eyeMidX)/eyeDist > 0.32 {
		return "检测到侧脸，已拒绝"
	}

	mouthMidY := (rmy + lmy) / 2
	vert := mouthMidY - eyeMidY
	if vert < eyeDist*0.35 {
		// Eyes and mouth collapsed vertically — usually a strong head-down pose.
		return "检测到低头，已拒绝"
	}

	// Pitch: nose position between eyes and mouth.
	// Frontal ≈ 0.45–0.70; looking down pushes toward the mouth; looking up toward the eyes.
	t := (ny - eyeMidY) / vert
	if t > 0.78 {
		return "检测到低头，已拒绝"
	}
	if t < 0.28 {
		return "检测到抬头，已拒绝"
	}
	return ""
}

// faceRollReject rejects head tilt from the eye line (independent of yaw/pitch).
func faceRollReject(f faceHit) string {
	rex, rey := f.kps[0], f.kps[1]
	lex, ley := f.kps[2], f.kps[3]
	eyeDist := math.Hypot(rex-lex, rey-ley)
	if eyeDist < 1 {
		return ""
	}
	roll := math.Abs(math.Atan2(ley-rey, lex-rex) * 180 / math.Pi)
	if roll > 10 {
		return "检测到头部倾斜，已拒绝"
	}
	return ""
}

// shoulderTiltReject checks the torso band under the face for:
// 1) roll — left/right shoulder height differ
// 2) yaw  — face looks forward but body/shoulders are turned (asymmetric width)
// Inconclusive / busy backgrounds return OK (empty string).
func shoulderTiltReject(img *image.NRGBA, box []float64) string {
	if img == nil || len(box) < 4 {
		return ""
	}
	x1, y1, x2, y2 := box[0], box[1], box[2], box[3]
	fw, fh := x2-x1, y2-y1
	if fw < 24 || fh < 24 {
		return ""
	}
	_ = y1
	iw := float64(img.Bounds().Dx())
	ih := float64(img.Bounds().Dy())
	bandTop := clampFloat(y2-fh*0.08, 0, ih-1)
	bandBot := clampFloat(y2+fh*0.95, 0, ih-1)
	if bandBot-bandTop < fh*0.25 {
		return ""
	}
	faceCX := (x1 + x2) / 2

	// --- roll: shoulder height ---
	lx := clampFloat(x1-fw*0.22, 0, iw-1)
	rx := clampFloat(x2+fw*0.22, 0, iw-1)
	halfW := fw * 0.18
	ly, lok := shoulderTopY(img, lx, bandTop, bandBot, halfW)
	ry, rok := shoulderTopY(img, rx, bandTop, bandBot, halfW)
	if lok && rok {
		ang := math.Abs(math.Atan2(ry-ly, rx-lx) * 180 / math.Pi)
		if ang > 12 {
			return "检测到肩膀倾斜，已拒绝"
		}
	}

	// --- yaw / twist: face frontal but one shoulder much closer / wider ---
	leftExt, rightExt, ok := shoulderHorizontalExtents(img, faceCX, bandTop, bandBot, fw)
	if ok && leftExt > fw*0.12 && rightExt > fw*0.12 {
		ratio := leftExt / rightExt
		if ratio < 1 {
			ratio = 1 / ratio
		}
		// Frontal shoulders ≈ 1.0; turned torso pushes one side out.
		if ratio > 1.32 {
			return "检测到肩膀扭转，已拒绝"
		}
		offset := math.Abs(leftExt-rightExt) / (leftExt + rightExt)
		if offset > 0.17 {
			return "检测到肩膀扭转，已拒绝"
		}
	}
	return ""
}

// shoulderHorizontalExtents returns mean distance from faceCX to the left/right
// torso silhouette edges inside the shoulder band.
func shoulderHorizontalExtents(img *image.NRGBA, faceCX, y0, y1, faceW float64) (leftExt, rightExt float64, ok bool) {
	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	row0 := int(math.Max(0, math.Floor(y0+(y1-y0)*0.12)))
	row1 := int(math.Min(float64(ih), math.Ceil(y0+(y1-y0)*0.78)))
	if row1-row0 < 6 {
		return 0, 0, false
	}
	searchL := int(math.Max(0, math.Floor(faceCX-faceW*1.55)))
	searchR := int(math.Min(float64(iw), math.Ceil(faceCX+faceW*1.55)))
	if searchR-searchL < 16 {
		return 0, 0, false
	}

	ref, refOK := shoulderBackdropLuma(img, searchL, searchR, row0, row1)
	if !refOK {
		return 0, 0, false
	}
	// Foreground is darker than a light/grey studio backdrop.
	thresh := ref - 18
	if thresh > ref*0.88 {
		thresh = ref * 0.88
	}

	var sumL, sumR float64
	nOK := 0
	cx := int(math.Round(faceCX))
	if cx < searchL {
		cx = searchL
	}
	if cx >= searchR {
		cx = searchR - 1
	}
	for y := row0; y < row1; y++ {
		leftEdge, rightEdge := -1, -1
		for x := searchL; x <= cx; x++ {
			i := img.PixOffset(x, y)
			luma := 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
			if luma < thresh {
				leftEdge = x
				break
			}
		}
		for x := searchR - 1; x >= cx; x-- {
			i := img.PixOffset(x, y)
			luma := 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
			if luma < thresh {
				rightEdge = x
				break
			}
		}
		if leftEdge < 0 || rightEdge < 0 || rightEdge <= leftEdge {
			continue
		}
		sumL += faceCX - float64(leftEdge)
		sumR += float64(rightEdge) - faceCX
		nOK++
	}
	if nOK < 4 {
		return 0, 0, false
	}
	return sumL / float64(nOK), sumR / float64(nOK), true
}

// shoulderBackdropLuma estimates studio backdrop brightness from band-side
// strips and full-image top corners. Grey backdrops (~100–140) are allowed.
func shoulderBackdropLuma(img *image.NRGBA, searchL, searchR, row0, row1 int) (float64, bool) {
	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	refRows := row0 + (row1-row0)/5
	if refRows <= row0 {
		refRows = row0 + 1
	}
	corner := (searchR - searchL) / 10
	if corner < 2 {
		corner = 2
	}
	var bandSum float64
	bandN := 0
	for y := row0; y < refRows && y < ih; y++ {
		for x := searchL; x < searchL+corner && x < iw; x++ {
			i := img.PixOffset(x, y)
			bandSum += 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
			bandN++
		}
		for x := searchR - corner; x < searchR && x < iw; x++ {
			if x < 0 {
				continue
			}
			i := img.PixOffset(x, y)
			bandSum += 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
			bandN++
		}
	}
	var cornerSum float64
	cornerN := 0
	patch := 20
	if patch > iw/4 {
		patch = iw / 4
	}
	if patch > ih/8 {
		patch = ih / 8
	}
	if patch < 4 {
		patch = 4
	}
	for y := 0; y < patch; y++ {
		for x := 0; x < patch; x++ {
			i := img.PixOffset(x, y)
			cornerSum += 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
			cornerN++
		}
		for x := iw - patch; x < iw; x++ {
			i := img.PixOffset(x, y)
			cornerSum += 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
			cornerN++
		}
	}
	ref := 0.0
	if bandN > 0 {
		ref = bandSum / float64(bandN)
	}
	if cornerN > 0 {
		cRef := cornerSum / float64(cornerN)
		// Prefer the brighter of band-side vs frame corners (true backdrop).
		if cRef > ref {
			ref = cRef
		}
		// If band sides are contaminated by hair/shoulder, lean on frame corners.
		if bandN > 0 && math.Abs(bandSum/float64(bandN)-cRef) > 35 && cRef >= 95 {
			ref = cRef
		}
	}
	if ref < 95 {
		return 0, false // too dark / textured — unreliable
	}
	return ref, true
}

// shoulderTopY walks down a vertical strip and returns the first row that looks
// like torso foreground (darker / more saturated than a light studio backdrop).
func shoulderTopY(img *image.NRGBA, cx, y0, y1, halfW float64) (float64, bool) {
	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	xStart := int(math.Max(0, math.Floor(cx-halfW)))
	xEnd := int(math.Min(float64(iw), math.Ceil(cx+halfW)))
	if xEnd-xStart < 4 {
		return 0, false
	}
	row0 := int(math.Max(0, math.Floor(y0)))
	row1 := int(math.Min(float64(ih), math.Ceil(y1)))
	if row1-row0 < 8 {
		return 0, false
	}
	// Backdrop reference: mean luma near the top of this strip.
	var refSum float64
	refN := 0
	refRows := row0 + (row1-row0)/6
	if refRows <= row0 {
		refRows = row0 + 1
	}
	for y := row0; y < refRows; y++ {
		for x := xStart; x < xEnd; x++ {
			i := img.PixOffset(x, y)
			refSum += 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
			refN++
		}
	}
	if refN < 1 {
		return 0, false
	}
	ref := refSum / float64(refN)
	// Grey studio backdrops are often ~100–140 (not pure white).
	if ref < 95 {
		return 0, false
	}
	thresh := ref - 18
	run := 0
	for y := row0; y < row1; y++ {
		dark := 0
		n := 0
		for x := xStart; x < xEnd; x++ {
			i := img.PixOffset(x, y)
			luma := 0.299*float64(img.Pix[i]) + 0.587*float64(img.Pix[i+1]) + 0.114*float64(img.Pix[i+2])
			n++
			if luma < thresh {
				dark++
			}
		}
		if n > 0 && dark*100 >= n*42 {
			run++
			if run >= 3 {
				return float64(y - 2), true
			}
		} else {
			run = 0
		}
	}
	return 0, false
}
