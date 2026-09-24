package core

import (
	"fmt"
	"image"
	"math"
	"path/filepath"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

// CelebAMask-HQ / face-parsing.PyTorch class ids (yakhyo order).
const (
	parseBG = iota
	parseSkin
	parseLBrow
	parseRBrow
	parseLEye
	parseREye
	parseGlasses
	parseLEar
	parseREar
	parseEarring
	parseNose
	parseMouth
	parseULip
	parseLLip
	parseNeck
	parseNecklace
	parseCloth
	parseHair
	parseHat
	parseClassN
)

const (
	parseModelFile = "bisenet-resnet18.onnx"
	parseSize      = 512
)

type parseSession struct {
	mu   sync.Mutex
	sess *ort.AdvancedSession
	in   *ort.Tensor[float32]
	out  *ort.Tensor[float32]
}

var (
	parseOnce sync.Once
	parseInst *parseSession
	parseErr  error
)

func parseNet() (*parseSession, error) {
	parseOnce.Do(func() {
		root, err := runtimeRoot()
		if err != nil {
			parseErr = err
			return
		}
		prepareRuntimeDir(root)
		if err := initONNX(filepath.Join(root, "onnxruntime.dll")); err != nil {
			parseErr = err
			return
		}
		in, err := ort.NewEmptyTensor[float32](ort.NewShape(1, 3, parseSize, parseSize))
		if err != nil {
			parseErr = err
			return
		}
		out, err := ort.NewEmptyTensor[float32](ort.NewShape(1, parseClassN, parseSize, parseSize))
		if err != nil {
			in.Destroy()
			parseErr = err
			return
		}
		sess, err := ort.NewAdvancedSession(
			filepath.Join(root, "models", parseModelFile),
			[]string{"input"},
			[]string{"output"},
			[]ort.Value{in},
			[]ort.Value{out},
			nil,
		)
		if err != nil {
			in.Destroy()
			out.Destroy()
			parseErr = fmt.Errorf("加载 BiSeNet 失败: %w", err)
			return
		}
		parseInst = &parseSession{sess: sess, in: in, out: out}
	})
	if parseErr != nil {
		return nil, parseErr
	}
	return parseInst, nil
}

// faceParseReject uses BiSeNet face parsing to reject closed eyes and
// common occlusions (mask / sunglasses / hat covering the face).
func faceParseReject(img *image.NRGBA, box []float64) string {
	if img == nil || len(box) < 4 {
		return ""
	}
	// Extra top pad so caps / hats above the forehead stay in the parse crop.
	// Side pad keeps hands on cheeks inside the window.
	// Keep bottom pad small so shirt collars on ID photos are not treated as masks.
	crop := cropFaceNRGBAPad(img, box[0], box[1], box[2], box[3], 0.22, 0.45, 0.22, 0.06)
	if crop == nil {
		return ""
	}
	labels, err := runFaceParse(crop)
	if err != nil || labels == nil {
		// Parsing is advisory; do not block the pipeline on model failure.
		return ""
	}
	return judgeParseLabels(labels)
}

func cropFaceNRGBA(img *image.NRGBA, x1, y1, x2, y2, padRatio float64) *image.NRGBA {
	return cropFaceNRGBAPad(img, x1, y1, x2, y2, padRatio, padRatio, padRatio, padRatio)
}

func cropFaceNRGBAPad(img *image.NRGBA, x1, y1, x2, y2, padL, padT, padR, padB float64) *image.NRGBA {
	bw := img.Bounds().Dx()
	bh := img.Bounds().Dy()
	fw := x2 - x1
	fh := y2 - y1
	if fw < 8 || fh < 8 {
		return nil
	}
	l := int(math.Floor(x1 - fw*padL))
	t := int(math.Floor(y1 - fh*padT))
	r := int(math.Ceil(x2 + fw*padR))
	b := int(math.Ceil(y2 + fh*padB))
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
	if w < 16 || h < 16 {
		return nil
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si := img.PixOffset(l+x, t+y)
			di := dst.PixOffset(x, y)
			copy(dst.Pix[di:di+4], img.Pix[si:si+4])
		}
	}
	return dst
}

func runFaceParse(face *image.NRGBA) ([]uint8, error) {
	net, err := parseNet()
	if err != nil {
		return nil, err
	}
	net.mu.Lock()
	defer net.mu.Unlock()
	fillParseInput(net.in.GetData(), face)
	if err := net.sess.Run(); err != nil {
		return nil, err
	}
	return argmaxParse(net.out.GetData(), parseSize, parseClassN), nil
}

func fillParseInput(dst []float32, src *image.NRGBA) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	plane := parseSize * parseSize
	mean := [3]float32{0.485, 0.456, 0.406}
	std := [3]float32{0.229, 0.224, 0.225}
	for y := 0; y < parseSize; y++ {
		sy := (float64(y)+0.5)*float64(sh)/float64(parseSize) - 0.5
		for x := 0; x < parseSize; x++ {
			sx := (float64(x)+0.5)*float64(sw)/float64(parseSize) - 0.5
			r, g, b := sampleNRGBA(src.Pix, src.Stride, sw, sh, sx, sy)
			i := y*parseSize + x
			dst[i] = (float32(r)/255 - mean[0]) / std[0]
			dst[plane+i] = (float32(g)/255 - mean[1]) / std[1]
			dst[2*plane+i] = (float32(b)/255 - mean[2]) / std[2]
		}
	}
}

func argmaxParse(logits []float32, side, classes int) []uint8 {
	out := make([]uint8, side*side)
	plane := side * side
	for i := 0; i < plane; i++ {
		best := 0
		bestV := logits[i]
		for c := 1; c < classes; c++ {
			v := logits[c*plane+i]
			if v > bestV {
				bestV = v
				best = c
			}
		}
		out[i] = uint8(best)
	}
	return out
}

func judgeParseLabels(labels []uint8) string {
	if len(labels) == 0 {
		return ""
	}
	var counts [parseClassN]int
	for _, id := range labels {
		if int(id) < parseClassN {
			counts[id]++
		}
	}
	eyes := counts[parseLEye] + counts[parseREye]
	brows := counts[parseLBrow] + counts[parseRBrow]
	nose := counts[parseNose]
	mouth := counts[parseMouth] + counts[parseULip] + counts[parseLLip]
	skin := counts[parseSkin]
	glasses := counts[parseGlasses]
	hat := counts[parseHat]
	hair := counts[parseHair]

	faceCore := skin + brows + eyes + nose + mouth + counts[parseLEar] + counts[parseREar]
	if faceCore < 800 {
		// Parsing failed or crop almost empty — skip hard reject.
		return ""
	}
	fc := float64(faceCore)

	// Closed eyes: almost no eye pixels while brows / face are present.
	if eyes < int(fc*0.0025) || (brows > 200 && eyes < brows/8) {
		return "检测到闭眼，已拒绝"
	}

	// Surgical mask / hand / scarf: nose or mouth largely missing.
	if nose < int(fc*0.012) {
		return "检测到人脸遮挡，已拒绝"
	}
	if mouth < int(fc*0.01) {
		return "检测到人脸遮挡，已拒绝"
	}

	// Sunglasses: glasses region large but eyes barely visible.
	if glasses > int(fc*0.035) && eyes < int(fc*0.004) {
		return "检测到墨镜遮挡，已拒绝"
	}

	// Cloth only counts as occlusion when it covers the mid-face band
	// (where a mask sits) AND mouth/nose are weak. Shirt collars on ID
	// photos produce lots of cloth at the bottom and must not trip this.
	if midFaceClothRatio(labels) > 0.45 && (nose < int(fc*0.02) || mouth < int(fc*0.015)) {
		return "检测到人脸遮挡，已拒绝"
	}

	// Hats before the generic hair rule — caps often get hair labels.
	if hat >= 600 || hat > int(fc*0.02) {
		return "检测到帽子遮挡，已拒绝"
	}
	if topHatRatio(labels) >= 0.06 {
		return "检测到帽子遮挡，已拒绝"
	}
	if topCapLikeCover(labels) {
		return "检测到帽子遮挡，已拒绝"
	}

	if hair > int(fc*0.7) && skin < int(fc*0.15) {
		return "检测到人脸遮挡，已拒绝"
	}

	// Hand / object covering one half of the face: BiSeNet often paints the
	// palm as skin or hair, but eye/brow/ear on that side disappear together.
	if hemifaceOcclusion(counts) {
		return "检测到人脸遮挡，已拒绝"
	}

	// Hand on cheek / temple: palm as skin + ear/cheek zone cues.
	if handOnCheek(labels, counts) {
		return "检测到人脸遮挡，已拒绝"
	}
	return ""
}

// hemifaceOcclusion is true when ≥2 of {eye, brow, ear} pairs are present on
// one side and nearly absent on the other — typical of a hand covering a cheek.
func hemifaceOcclusion(counts [parseClassN]int) bool {
	type pair struct{ a, b int }
	pairs := []pair{
		{counts[parseLEye], counts[parseREye]},
		{counts[parseLBrow], counts[parseRBrow]},
		{counts[parseLEar], counts[parseREar]},
	}
	leftHeavy, rightHeavy := 0, 0
	for _, p := range pairs {
		if p.a >= 150 && p.b < 40 {
			leftHeavy++ // subject-left features only
		}
		if p.b >= 150 && p.a < 40 {
			rightHeavy++ // subject-right features only
		}
	}
	// All deficits on the same side of the face.
	return leftHeavy >= 2 || rightHeavy >= 2
}

// handOnCheek detects a hand covering one cheek/ear. Hands are labeled as skin,
// so we look for ear asymmetry plus skin (not hair) filling the missing-ear zone.
func handOnCheek(labels []uint8, counts [parseClassN]int) bool {
	lear, rear := counts[parseLEar], counts[parseREar]
	// CelebAMask: LEar = subject's left (image-right), REar = subject's right (image-left).
	if lear >= 200 && rear < 50 {
		// Subject's right ear gone → image-left ear/cheek zone.
		if earZoneCoveredBySkin(labels, true) || outerSkinIntrusion(labels, true) {
			return true
		}
		// Near-total one-sided ear loss is abnormal for a clean frontal ID crop.
		if rear < 15 && lear >= 400 && !earZoneCoveredByHair(labels, true) {
			return true
		}
	}
	if rear >= 200 && lear < 50 {
		if earZoneCoveredBySkin(labels, false) || outerSkinIntrusion(labels, false) {
			return true
		}
		if lear < 15 && rear >= 400 && !earZoneCoveredByHair(labels, false) {
			return true
		}
	}
	return asymmetricSideSkin(labels)
}

// earZoneCounts inspects the approximate ear/cheek strip on one image side.
func earZoneCounts(labels []uint8, imageLeft bool) (skin, hair, ear, total int) {
	side := parseSide(labels)
	if side < 16 {
		return 0, 0, 0, 0
	}
	// Subject's right ear sits on the image-left flank for a frontal face.
	x0, x1 := side*4/100, side*32/100
	if !imageLeft {
		x0, x1 = side*68/100, side*96/100
	}
	y0, y1 := side*28/100, side*62/100
	for y := y0; y < y1; y++ {
		row := y * side
		for x := x0; x < x1; x++ {
			total++
			switch labels[row+x] {
			case parseSkin:
				skin++
			case parseHair:
				hair++
			case parseLEar, parseREar:
				ear++
			}
		}
	}
	return skin, hair, ear, total
}

func earZoneCoveredBySkin(labels []uint8, imageLeft bool) bool {
	skin, hair, ear, total := earZoneCounts(labels, imageLeft)
	if total < 80 {
		return false
	}
	// Hand on cheek: ear gone, skin dominates (hand), hair is not the cover.
	return ear < total/25 && skin >= total*35/100 && skin > hair+total/10
}

func earZoneCoveredByHair(labels []uint8, imageLeft bool) bool {
	skin, hair, ear, total := earZoneCounts(labels, imageLeft)
	if total < 80 {
		return false
	}
	return hair >= total*45/100 && hair > skin && ear < total/15
}

// outerSkinIntrusion reports a thick skin slab on one image side that also
// touches the bottom edge (typical of a hand/arm entering the frame).
// imageLeft=true checks the left columns of the parse map.
func outerSkinIntrusion(labels []uint8, imageLeft bool) bool {
	side := parseSide(labels)
	if side < 16 {
		return false
	}
	x0, x1 := 0, side*20/100
	if !imageLeft {
		x0, x1 = side*80/100, side
	}
	y0 := side * 25 / 100
	y1 := side * 90 / 100
	skin, total, bottomSkin := 0, 0, 0
	bottomY0 := side * 78 / 100
	for y := y0; y < y1; y++ {
		row := y * side
		for x := x0; x < x1; x++ {
			total++
			if labels[row+x] != parseSkin {
				continue
			}
			skin++
			if y >= bottomY0 {
				bottomSkin++
			}
		}
	}
	if total < 1 {
		return false
	}
	skinR := float64(skin) / float64(total)
	bottomBand := (x1 - x0) * (y1 - bottomY0)
	if bottomBand < 1 {
		bottomBand = 1
	}
	bottomR := float64(bottomSkin) / float64(bottomBand)
	return skinR >= 0.38 && bottomR >= 0.12
}

// asymmetricSideSkin: one outer flank is mostly skin (hand), the other is not.
func asymmetricSideSkin(labels []uint8) bool {
	left := outerSkinRatio(labels, true)
	right := outerSkinRatio(labels, false)
	leftBottom := outerTouchesBottom(labels, true)
	rightBottom := outerTouchesBottom(labels, false)
	if left >= 0.40 && leftBottom && right < 0.26 {
		return true
	}
	if right >= 0.40 && rightBottom && left < 0.26 {
		return true
	}
	return false
}

func outerSkinRatio(labels []uint8, imageLeft bool) float64 {
	side := parseSide(labels)
	if side < 16 {
		return 0
	}
	x0, x1 := 0, side*18/100
	if !imageLeft {
		x0, x1 = side*82/100, side
	}
	y0, y1 := side*28/100, side*85/100
	skin, total := 0, 0
	for y := y0; y < y1; y++ {
		row := y * side
		for x := x0; x < x1; x++ {
			total++
			if labels[row+x] == parseSkin {
				skin++
			}
		}
	}
	if total < 1 {
		return 0
	}
	return float64(skin) / float64(total)
}

func outerTouchesBottom(labels []uint8, imageLeft bool) bool {
	side := parseSide(labels)
	if side < 16 {
		return false
	}
	x0, x1 := 0, side*18/100
	if !imageLeft {
		x0, x1 = side*82/100, side
	}
	y0 := side * 82 / 100
	skin, total := 0, 0
	for y := y0; y < side; y++ {
		row := y * side
		for x := x0; x < x1; x++ {
			total++
			if labels[row+x] == parseSkin {
				skin++
			}
		}
	}
	if total < 1 {
		return false
	}
	return float64(skin)/float64(total) >= 0.10
}

// topHatRatio is the share of hat pixels in the upper 30% of the parse map.
func topHatRatio(labels []uint8) float64 {
	side := parseSide(labels)
	if side < 8 {
		return 0
	}
	y1 := side * 30 / 100
	hat, total := 0, 0
	for y := 0; y < y1; y++ {
		row := y * side
		for x := 0; x < side; x++ {
			total++
			if labels[row+x] == parseHat {
				hat++
			}
		}
	}
	if total < 1 {
		return 0
	}
	return float64(hat) / float64(total)
}

// topCapLikeCover detects a hard forehead cover (cap) when BiSeNet labels the
// brim as hair: upper band almost no skin, mostly hair/hat/bg, while the
// mid-face still shows a normal skin mass.
func topCapLikeCover(labels []uint8) bool {
	side := parseSide(labels)
	if side < 8 {
		return false
	}
	yTop := side * 28 / 100
	yMid0 := side * 30 / 100
	yMid1 := side * 55 / 100
	topSkin, topHairHat, topTotal := 0, 0, 0
	midSkin, midTotal := 0, 0
	for y := 0; y < yTop; y++ {
		row := y * side
		for x := 0; x < side; x++ {
			topTotal++
			switch labels[row+x] {
			case parseSkin:
				topSkin++
			case parseHair, parseHat:
				topHairHat++
			}
		}
	}
	for y := yMid0; y < yMid1; y++ {
		row := y * side
		for x := 0; x < side; x++ {
			midTotal++
			if labels[row+x] == parseSkin {
				midSkin++
			}
		}
	}
	if topTotal < 1 || midTotal < 1 {
		return false
	}
	topSkinR := float64(topSkin) / float64(topTotal)
	topCoverR := float64(topHairHat) / float64(topTotal)
	midSkinR := float64(midSkin) / float64(midTotal)
	return midSkinR > 0.22 && topSkinR < 0.04 && topCoverR > 0.55
}

func parseSide(labels []uint8) int {
	side := parseSize
	if len(labels) != side*side {
		side = int(math.Sqrt(float64(len(labels))))
		if side*side != len(labels) {
			return 0
		}
	}
	return side
}

// midFaceClothRatio is the share of cloth pixels in the vertical band
// where surgical masks usually cover (roughly nose to chin).
func midFaceClothRatio(labels []uint8) float64 {
	side := parseSide(labels)
	if side < 8 {
		return 0
	}
	y0 := side * 35 / 100
	y1 := side * 80 / 100
	cloth, total := 0, 0
	for y := y0; y < y1; y++ {
		row := y * side
		for x := 0; x < side; x++ {
			total++
			if labels[row+x] == parseCloth {
				cloth++
			}
		}
	}
	if total < 1 {
		return 0
	}
	return float64(cloth) / float64(total)
}
