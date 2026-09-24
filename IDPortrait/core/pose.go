package core

import (
	"image"
	"math"
)

// facePoseReject returns a Chinese reject reason when the five YuNet keypoints
// show a non-frontal pose (side face / looking down / looking up / head tilt).
// Empty string means the pose is acceptable for an ID photo.
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

	// Roll: eye line should be nearly horizontal (neck / head tilt).
	roll := math.Abs(math.Atan2(ley-rey, lex-rex) * 180 / math.Pi)
	if roll > 10 {
		return "检测到头部倾斜，已拒绝"
	}

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

// shoulderTiltReject estimates left/right shoulder height from the torso band
// under the face box. Inconclusive backgrounds return OK (empty string).
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
	bandBot := clampFloat(y2+fh*0.85, 0, ih-1)
	if bandBot-bandTop < fh*0.25 {
		return ""
	}
	lx := clampFloat(x1-fw*0.22, 0, iw-1)
	rx := clampFloat(x2+fw*0.22, 0, iw-1)
	halfW := fw * 0.18
	ly, lok := shoulderTopY(img, lx, bandTop, bandBot, halfW)
	ry, rok := shoulderTopY(img, rx, bandTop, bandBot, halfW)
	if !lok || !rok {
		return ""
	}
	ang := math.Abs(math.Atan2(ry-ly, rx-lx) * 180 / math.Pi)
	if ang > 12 {
		return "检测到肩膀倾斜，已拒绝"
	}
	return ""
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
	// Need a reasonably light backdrop to trust this heuristic.
	if ref < 140 {
		return 0, false
	}
	thresh := ref - 28
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
