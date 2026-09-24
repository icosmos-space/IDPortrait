package core

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

const (
	yunetModelFile = "face_detection_yunet_2023mar.onnx"
	yunetInput     = 640
	yunetScoreMin  = 0.6
	yunetNMS       = 0.3
)

type faceHit struct {
	x, y, w, h float64
	kps        [10]float64
	score      float64
}

type yunetDetector struct {
	mu      sync.Mutex
	session *ort.AdvancedSession
	input   *ort.Tensor[float32]
	outs    []*ort.Tensor[float32]
}

var (
	yunetOnce sync.Mutex
	yunetInst *yunetDetector
)

// DetectFace runs YuNet on a new photo before it enters preview.
// Exactly one frontal face is accepted. Sideways photos are rotated upright.
// Optional prechecks (pose / blur / mosaic / closed eyes / occlusion) follow opts.
// The live camera viewfinder does not use this path.
func DetectFace(src string, opts FaceCheckOptions) (*FaceCheckResult, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, fmt.Errorf("empty image")
	}
	img, originalURL, err := openPortrait(src)
	if err != nil {
		return nil, err
	}
	origW, origH := img.Bounds().Dx(), img.Bounds().Dy()
	img = downscale(img, 1600)
	base := toNRGBA(img)

	det, err := yunet()
	if err != nil {
		return nil, err
	}

	angles := []int{0, 90, 270, 180}
	var best *FaceCheckResult
	var bestView *image.NRGBA
	sawMultiple := false
	for _, angle := range angles {
		view := base
		if angle != 0 {
			view = rotateClockwise(base, angle)
		}
		hits, err := det.infer(view)
		if err != nil {
			return nil, err
		}
		if len(hits) > 1 {
			sawMultiple = true
			continue
		}
		if len(hits) != 1 {
			continue
		}
		hit := hits[0]
		if best != nil && hit.score <= best.Score {
			continue
		}
		url := originalURL
		if angle != 0 || view.Bounds().Dx() != origW || view.Bounds().Dy() != origH {
			url, err = encodeJPEGDataURL(view, 92)
			if err != nil {
				return nil, err
			}
		}
		lm := make([]float64, 10)
		copy(lm, hit.kps[:])
		best = &FaceCheckResult{
			OK:        true,
			ImgBase64: url,
			FaceBox:   []float64{hit.x, hit.y, hit.x + hit.w, hit.y + hit.h},
			Landmarks: lm,
			Score:     hit.score,
		}
		bestView = view
	}
	if best != nil {
		poseHit := faceHit{
			x: best.FaceBox[0], y: best.FaceBox[1],
			w: best.FaceBox[2] - best.FaceBox[0], h: best.FaceBox[3] - best.FaceBox[1],
			score: best.Score,
		}
		copy(poseHit.kps[:], best.Landmarks)
		if !opts.SkipPose {
			if reason := facePoseReject(poseHit); reason != "" {
				return &FaceCheckResult{OK: false, Reason: reason, ImgBase64: best.ImgBase64}, nil
			}
		}
		if !opts.SkipBlur || !opts.SkipMosaic {
			if reason := faceQualityReject(bestView, best.FaceBox, !opts.SkipBlur, !opts.SkipMosaic); reason != "" {
				return &FaceCheckResult{OK: false, Reason: reason, ImgBase64: best.ImgBase64}, nil
			}
		}
		if !opts.SkipParse {
			if reason := faceParseReject(bestView, best.FaceBox); reason != "" {
				return &FaceCheckResult{OK: false, Reason: reason, ImgBase64: best.ImgBase64}, nil
			}
		}
		return best, nil
	}
	if sawMultiple {
		return &FaceCheckResult{OK: false, Reason: "检测到多张人脸，已拒绝"}, nil
	}
	return &FaceCheckResult{OK: false, Reason: "未检测到人脸，已拒绝"}, nil
}

func openPortrait(src string) (image.Image, string, error) {
	if strings.HasPrefix(src, "data:image/") {
		img, err := decodeDataURL(src)
		if err != nil {
			return nil, "", err
		}
		return img, src, nil
	}
	img, _, err := loadImageFile(src)
	if err != nil {
		return nil, "", err
	}
	url, err := encodeJPEGDataURL(img, 92)
	if err != nil {
		return nil, "", err
	}
	return img, url, nil
}

func yunet() (*yunetDetector, error) {
	yunetOnce.Lock()
	defer yunetOnce.Unlock()
	if yunetInst != nil {
		return yunetInst, nil
	}
	d, err := newYunet()
	if err != nil {
		return nil, err
	}
	yunetInst = d
	return yunetInst, nil
}

func newYunet() (*yunetDetector, error) {
	root, err := runtimeRoot()
	if err != nil {
		return nil, err
	}
	prepareRuntimeDir(root)
	if err := initONNX(filepath.Join(root, "onnxruntime.dll")); err != nil {
		return nil, err
	}

	input, err := ort.NewEmptyTensor[float32](ort.NewShape(1, 3, yunetInput, yunetInput))
	if err != nil {
		return nil, fmt.Errorf("yunet input: %w", err)
	}
	strides := []int{8, 16, 32}
	channels := []int{1, 1, 4, 10}
	outs := make([]*ort.Tensor[float32], 0, 12)
	values := make([]ort.Value, 0, 12)
	for _, ch := range channels {
		for _, stride := range strides {
			n := (yunetInput / stride) * (yunetInput / stride)
			t, terr := ort.NewEmptyTensor[float32](ort.NewShape(1, int64(n), int64(ch)))
			if terr != nil {
				input.Destroy()
				for _, old := range outs {
					old.Destroy()
				}
				return nil, fmt.Errorf("yunet output: %w", terr)
			}
			outs = append(outs, t)
			values = append(values, t)
		}
	}

	names := []string{
		"cls_8", "cls_16", "cls_32",
		"obj_8", "obj_16", "obj_32",
		"bbox_8", "bbox_16", "bbox_32",
		"kps_8", "kps_16", "kps_32",
	}
	session, err := ort.NewAdvancedSession(
		filepath.Join(root, "models", yunetModelFile),
		[]string{"input"},
		names,
		[]ort.Value{input},
		values,
		nil,
	)
	if err != nil {
		input.Destroy()
		for _, old := range outs {
			old.Destroy()
		}
		return nil, fmt.Errorf("加载 YuNet 失败: %w", err)
	}
	return &yunetDetector{session: session, input: input, outs: outs}, nil
}

var (
	onnxOnce sync.Once
	onnxErr  error
)

func initONNX(dll string) error {
	onnxOnce.Do(func() {
		ort.SetSharedLibraryPath(dll)
		onnxErr = ort.InitializeEnvironment()
	})
	if onnxErr != nil {
		return fmt.Errorf("初始化 ONNX Runtime 失败: %w", onnxErr)
	}
	return nil
}

func runtimeRoot() (string, error) {
	var candidates []string
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "core", "runtime"),
			filepath.Join(dir, "runtime"),
		)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "runtime"),
			filepath.Join(cwd, "core", "runtime"),
			filepath.Join(cwd, "IDPortrait", "core", "runtime"),
		)
	}
	for _, dir := range candidates {
		dll := filepath.Join(dir, "onnxruntime.dll")
		model := filepath.Join(dir, "models", yunetModelFile)
		if fileExists(dll) && fileExists(model) {
			return dir, nil
		}
	}
	return "", fmt.Errorf("找不到 YuNet 模型 %s", yunetModelFile)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (d *yunetDetector) infer(img *image.NRGBA) ([]faceHit, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	nw, nh := fillYuNetInput(d.input.GetData(), img)
	if err := d.session.Run(); err != nil {
		return nil, fmt.Errorf("YuNet 推理失败: %w", err)
	}
	return decodeYuNet(d.outs, img.Bounds().Dx(), img.Bounds().Dy(), nw, nh), nil
}

func fillYuNetInput(dst []float32, img *image.NRGBA) (nw, nh int) {
	for i := range dst {
		dst[i] = 0
	}
	sw, sh := img.Bounds().Dx(), img.Bounds().Dy()
	if sw < 1 || sh < 1 {
		return 0, 0
	}
	scale := math.Min(float64(yunetInput)/float64(sw), float64(yunetInput)/float64(sh))
	nw = int(math.Round(float64(sw) * scale))
	nh = int(math.Round(float64(sh) * scale))
	if nw > yunetInput {
		nw = yunetInput
	}
	if nh > yunetInput {
		nh = yunetInput
	}
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	plane := yunetInput * yunetInput
	pix := img.Pix
	stride := img.Stride
	for y := 0; y < nh; y++ {
		sy := (float64(y)+0.5)*float64(sh)/float64(nh) - 0.5
		for x := 0; x < nw; x++ {
			sx := (float64(x)+0.5)*float64(sw)/float64(nw) - 0.5
			r, g, b := sampleNRGBA(pix, stride, sw, sh, sx, sy)
			idx := y*yunetInput + x
			dst[idx] = b
			dst[plane+idx] = g
			dst[2*plane+idx] = r
		}
	}
	return nw, nh
}

func sampleNRGBA(pix []uint8, stride, w, h int, x, y float64) (r, g, b float32) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	maxX := float64(w - 1)
	maxY := float64(h - 1)
	if x > maxX {
		x = maxX
	}
	if y > maxY {
		y = maxY
	}
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1
	if x1 >= w {
		x1 = w - 1
	}
	if y1 >= h {
		y1 = h - 1
	}
	dx := float32(x - float64(x0))
	dy := float32(y - float64(y0))
	r00, g00, b00 := pixRGB(pix, stride, x0, y0)
	r10, g10, b10 := pixRGB(pix, stride, x1, y0)
	r01, g01, b01 := pixRGB(pix, stride, x0, y1)
	r11, g11, b11 := pixRGB(pix, stride, x1, y1)
	r = lerp(lerp(r00, r10, dx), lerp(r01, r11, dx), dy)
	g = lerp(lerp(g00, g10, dx), lerp(g01, g11, dx), dy)
	b = lerp(lerp(b00, b10, dx), lerp(b01, b11, dx), dy)
	return r, g, b
}

func pixRGB(pix []uint8, stride, x, y int) (r, g, b float32) {
	i := y*stride + x*4
	return float32(pix[i]), float32(pix[i+1]), float32(pix[i+2])
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

func decodeYuNet(outs []*ort.Tensor[float32], sw, sh, nw, nh int) []faceHit {
	if nw < 1 || nh < 1 || sw < 1 || sh < 1 {
		return nil
	}
	scaleX := float64(sw) / float64(nw)
	scaleY := float64(sh) / float64(nh)
	strides := []int{8, 16, 32}
	cls := []*ort.Tensor[float32]{outs[0], outs[1], outs[2]}
	obj := []*ort.Tensor[float32]{outs[3], outs[4], outs[5]}
	bbox := []*ort.Tensor[float32]{outs[6], outs[7], outs[8]}
	kps := []*ort.Tensor[float32]{outs[9], outs[10], outs[11]}

	hits := make([]faceHit, 0, 8)
	for i, stride := range strides {
		cols := yunetInput / stride
		rows := yunetInput / stride
		clsV := cls[i].GetData()
		objV := obj[i].GetData()
		bboxV := bbox[i].GetData()
		kpsV := kps[i].GetData()
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				idx := r*cols + c
				clsScore := clamp01(clsV[idx])
				objScore := clamp01(objV[idx])
				score := math.Sqrt(float64(clsScore * objScore))
				if score < yunetScoreMin {
					continue
				}
				cx := (float64(c) + float64(bboxV[idx*4])) * float64(stride)
				cy := (float64(r) + float64(bboxV[idx*4+1])) * float64(stride)
				bw := math.Exp(float64(bboxV[idx*4+2])) * float64(stride)
				bh := math.Exp(float64(bboxV[idx*4+3])) * float64(stride)
				x1 := cx - bw/2
				y1 := cy - bh/2
				if cx < 0 || cy < 0 || cx >= float64(nw) || cy >= float64(nh) {
					continue
				}
				var pts [10]float64
				for n := 0; n < 5; n++ {
					pts[n*2] = (float64(kpsV[idx*10+n*2]) + float64(c)) * float64(stride) * scaleX
					pts[n*2+1] = (float64(kpsV[idx*10+n*2+1]) + float64(r)) * float64(stride) * scaleY
				}
				hits = append(hits, faceHit{
					x:     x1 * scaleX,
					y:     y1 * scaleY,
					w:     bw * scaleX,
					h:     bh * scaleY,
					kps:   pts,
					score: score,
				})
			}
		}
	}
	kept := nmsFaces(hits, yunetNMS, 5000)
	for i := range kept {
		kept[i] = clampFace(kept[i], float64(sw), float64(sh))
	}
	return kept
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func nmsFaces(dets []faceHit, iouThr float64, topK int) []faceHit {
	if len(dets) <= 1 {
		return dets
	}
	sort.Slice(dets, func(i, j int) bool { return dets[i].score > dets[j].score })
	if len(dets) > 5000 {
		dets = dets[:5000]
	}
	dead := make([]bool, len(dets))
	kept := make([]faceHit, 0, len(dets))
	for i := range dets {
		if dead[i] {
			continue
		}
		kept = append(kept, dets[i])
		if topK > 0 && len(kept) >= topK {
			break
		}
		for j := i + 1; j < len(dets); j++ {
			if dead[j] {
				continue
			}
			if faceIoU(dets[i], dets[j]) >= iouThr {
				dead[j] = true
			}
		}
	}
	return kept
}

func faceIoU(a, b faceHit) float64 {
	ax2, ay2 := a.x+a.w, a.y+a.h
	bx2, by2 := b.x+b.w, b.y+b.h
	ix1 := math.Max(a.x, b.x)
	iy1 := math.Max(a.y, b.y)
	ix2 := math.Min(ax2, bx2)
	iy2 := math.Min(ay2, by2)
	iw := ix2 - ix1
	ih := iy2 - iy1
	if iw <= 0 || ih <= 0 {
		return 0
	}
	inter := iw * ih
	union := a.w*a.h + b.w*b.h - inter
	if union <= 0 {
		return 0
	}
	return inter / union
}

func clampFace(f faceHit, w, h float64) faceHit {
	if f.x < 0 {
		f.w += f.x
		f.x = 0
	}
	if f.y < 0 {
		f.h += f.y
		f.y = 0
	}
	if f.x+f.w > w {
		f.w = w - f.x
	}
	if f.y+f.h > h {
		f.h = h - f.y
	}
	if f.w < 1 {
		f.w = 1
	}
	if f.h < 1 {
		f.h = 1
	}
	for i := 0; i < 10; i += 2 {
		f.kps[i] = clampFloat(f.kps[i], 0, w-1)
		f.kps[i+1] = clampFloat(f.kps[i+1], 0, h-1)
	}
	return f
}

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func toNRGBA(src image.Image) *image.NRGBA {
	if n, ok := src.(*image.NRGBA); ok && n.Rect.Min.X == 0 && n.Rect.Min.Y == 0 {
		return n
	}
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func rotateClockwise(src *image.NRGBA, deg int) *image.NRGBA {
	switch ((deg % 360) + 360) % 360 {
	case 90:
		return rotate90(src)
	case 180:
		return rotate180(src)
	case 270:
		return rotate270(src)
	default:
		return src
	}
}

func rotate90(src *image.NRGBA) *image.NRGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			copy4(dst, h-1-y, x, src, x, y)
		}
	}
	return dst
}

func rotate180(src *image.NRGBA) *image.NRGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			copy4(dst, w-1-x, h-1-y, src, x, y)
		}
	}
	return dst
}

func rotate270(src *image.NRGBA) *image.NRGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			copy4(dst, y, w-1-x, src, x, y)
		}
	}
	return dst
}

func copy4(dst *image.NRGBA, dx, dy int, src *image.NRGBA, sx, sy int) {
	di := dst.PixOffset(dx, dy)
	si := src.PixOffset(sx, sy)
	copy(dst.Pix[di:di+4], src.Pix[si:si+4])
}

// detectFaces runs YuNet once on an already upright photo.
func detectFaces(img image.Image) ([]faceHit, error) {
	det, err := yunet()
	if err != nil {
		return nil, err
	}
	return det.infer(toNRGBA(img))
}
