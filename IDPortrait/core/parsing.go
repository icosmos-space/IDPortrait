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
	// Keep bottom pad small so shirt collars on ID photos are not treated as masks.
	crop := cropFaceNRGBAPad(img, box[0], box[1], box[2], box[3], 0.12, 0.12, 0.12, 0.04)
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

	if hair > int(fc*0.7) && skin < int(fc*0.15) {
		return "检测到人脸遮挡，已拒绝"
	}
	if hat > int(fc*0.35) && eyes < int(fc*0.008) {
		return "检测到帽子遮挡，已拒绝"
	}
	return ""
}

// midFaceClothRatio is the share of cloth pixels in the vertical band
// where surgical masks usually cover (roughly nose to chin).
func midFaceClothRatio(labels []uint8) float64 {
	side := parseSize
	if len(labels) != side*side {
		side = int(math.Sqrt(float64(len(labels))))
		if side*side != len(labels) || side < 8 {
			return 0
		}
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
