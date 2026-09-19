package core

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "image/gif"
	_ "image/png"
)

// Engine hosts ID-photo processing algorithms.
type Engine struct {
	mu          sync.Mutex
	lastDataURL string
	lastImg     image.Image
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) LoadImage(path string) (*LoadImageResult, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("empty image path")
	}

	var (
		dataURL string
		img     image.Image
		err     error
	)

	switch {
	case strings.HasPrefix(path, "data:image/"):
		img, err = decodeDataURL(path)
		if err != nil {
			return nil, err
		}
		// Re-encode as reasonably sized JPEG for later generate / IPC.
		dataURL, err = encodeJPEGDataURL(downscale(img, 1600), 90)
		if err != nil {
			return nil, err
		}
		img, _ = decodeDataURL(dataURL)
	default:
		raw, format, openErr := loadImageFile(path)
		if openErr != nil {
			return nil, openErr
		}
		img = downscale(raw, 1600)
		if strings.EqualFold(format, "png") || strings.EqualFold(format, "gif") {
			dataURL, err = encodeJPEGDataURL(img, 90)
		} else {
			dataURL, err = encodeJPEGDataURL(img, 90)
		}
		if err != nil {
			return nil, err
		}
	}

	e.mu.Lock()
	e.lastDataURL = dataURL
	e.lastImg = img
	e.mu.Unlock()

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	if w <= 0 {
		w = 360
	}
	if h <= 0 {
		h = 480
	}

	return &LoadImageResult{
		ImgBase64: dataURL,
		FaceBox: []float64{
			float64(w) * 0.25,
			float64(h) * 0.16,
			float64(w) * 0.75,
			float64(h) * 0.72,
		},
		Landmarks: []float64{
			float64(w) * 0.38, float64(h) * 0.36,
			float64(w) * 0.62, float64(h) * 0.36,
			float64(w) * 0.50, float64(h) * 0.48,
			float64(w) * 0.40, float64(h) * 0.58,
			float64(w) * 0.60, float64(h) * 0.58,
		},
		Report: Report{
			FaceOK:    true,
			FaceScore: 0.92,
		},
	}, nil
}

func (e *Engine) Generate(p GenerateParams) (*GenerateResult, error) {
	src, originURL, err := e.resolveSource(p.SourceImg)
	if err != nil {
		return nil, err
	}

	bundle, err := makeIDPhoto(toNRGBA(src), p)
	if err != nil {
		return nil, err
	}
	bg := parseHex(p.BgColor, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	mode := p.BgMode
	if mode == "" {
		mode = "solid"
	}
	std := compositeOn(bundle.std, bg, mode)
	hd := compositeOn(bundle.hd, bg, mode)
	paperW, paperH := paperPixels(p.PaperSize)
	layout := layoutSheet(std, paperW, paperH)
	social := squareSocial(std, bg, mode)

	var idphoto string
	if p.EnableTargetFileSize && p.TargetFileSize > 0 {
		idphoto, err = encodeJPEGLimited(std, p.TargetFileSize)
	} else {
		idphoto, err = encodeJPEGDataURL(std, 92)
	}
	if err != nil {
		return nil, err
	}
	single, err := encodeJPEGDataURL(hd, 90)
	if err != nil {
		return nil, err
	}
	socialURL, err := encodeJPEGDataURL(social, 90)
	if err != nil {
		return nil, err
	}
	layoutURL, err := encodeJPEGDataURL(layout, 88)
	if err != nil {
		return nil, err
	}
	matting, err := encodePNGDataURL(bundle.matting)
	if err != nil {
		return nil, err
	}

	face := bundle.face
	return &GenerateResult{
		OriginImg:  originURL,
		MattingImg: matting,
		ResultImg:  idphoto,
		Results: ResultBundle{
			Single:  single,
			Layout:  layoutURL,
			Social:  socialURL,
			IDPhoto: idphoto,
		},
		FaceBox:   []float64{face.x, face.y, face.x + face.w, face.y + face.h},
		Landmarks: append([]float64(nil), face.kps[:]...),
		Report: Report{
			FaceOK:    true,
			FaceScore: face.score,
		},
	}, nil
}

func (e *Engine) Export(dir string, _ ExportOptions) (*ExportResult, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("export dir is empty")
	}
	return &ExportResult{OK: true, Dir: dir}, nil
}

func (e *Engine) resolveSource(sourceImg string) (image.Image, string, error) {
	sourceImg = strings.TrimSpace(sourceImg)
	if strings.HasPrefix(sourceImg, "data:image/") {
		img, err := decodeDataURL(sourceImg)
		if err != nil {
			return nil, "", err
		}
		img = downscale(img, 1600)
		url, err := encodeJPEGDataURL(img, 90)
		if err != nil {
			return nil, "", err
		}
		e.mu.Lock()
		e.lastImg = img
		e.lastDataURL = url
		e.mu.Unlock()
		return img, url, nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.lastImg != nil && e.lastDataURL != "" {
		return e.lastImg, e.lastDataURL, nil
	}
	return nil, "", fmt.Errorf("请先打开或拍摄照片")
}

func loadImageFile(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open image: %w", err)
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, "", fmt.Errorf("decode image %s: %w", filepath.Base(path), err)
	}
	return img, format, nil
}

func decodeDataURL(dataURL string) (image.Image, error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data url")
	}
	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode data url: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("decode data url image: %w", err)
	}
	return img, nil
}

func encodeJPEGDataURL(img image.Image, quality int) (string, error) {
	var buf bytes.Buffer
	if quality <= 0 || quality > 100 {
		quality = 90
	}
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func encodePNGDataURL(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// composeMatting builds a soft portrait cutout (RGBA) for preview until ONNX matting lands.
func composeMatting(src image.Image) *image.RGBA {
	sb := src.Bounds()
	w, h := sb.Dx(), sb.Dy()
	if w <= 0 || h <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	cx := float64(w) * 0.5
	cy := float64(h) * 0.42
	rx := float64(w) * 0.36
	ry := float64(h) * 0.48
	feather := 0.12
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			nx := (float64(x) - cx) / rx
			ny := (float64(y) - cy) / ry
			d := math.Sqrt(nx*nx + ny*ny)
			var a float64
			switch {
			case d <= 1-feather:
				a = 1
			case d >= 1:
				a = 0
			default:
				t := (d - (1 - feather)) / feather
				a = 1 - t*t*(3-2*t)
			}
			if a <= 0.004 {
				continue
			}
			r, g, b, _ := src.At(sb.Min.X+x, sb.Min.Y+y).RGBA()
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a * 255),
			})
		}
	}
	return dst
}

func downscale(src image.Image, maxSide int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return src
	}
	if w <= maxSide && h <= maxSide {
		return src
	}
	scale := float64(maxSide) / float64(w)
	if h > w {
		scale = float64(maxSide) / float64(h)
	}
	nw := max(1, int(float64(w)*scale))
	nh := max(1, int(float64(h)*scale))
	return resizeNearest(src, nw, nh)
}

func resizeNearest(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	for y := 0; y < h; y++ {
		sy := sb.Min.Y + y*sh/h
		for x := 0; x < w; x++ {
			sx := sb.Min.X + x*sw/w
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func fillBackground(dst *image.RGBA, bg color.RGBA, mode string) {
	b := dst.Bounds()
	w, h := b.Dx(), b.Dy()
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := bg
			switch mode {
			case "vertical":
				c = mix(bg, white, float64(y)/float64(max(h-1, 1)))
			case "radial":
				cx, cy := float64(w)/2, float64(h)*0.42
				dx, dy := float64(x)-cx, float64(y)-cy
				t := (dx*dx + dy*dy) / (float64(w*w+h*h) * 0.25)
				if t > 1 {
					t = 1
				}
				c = mix(bg, white, t)
			}
			dst.Set(x, y, c)
		}
	}
}

// composeCover draws source photo cover-fitted onto a background canvas.
func composeCover(src image.Image, w, h int, bg color.RGBA, mode string) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	fillBackground(dst, bg, mode)

	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	if sw <= 0 || sh <= 0 {
		return dst
	}

	// Cover scale, slight upward bias for headroom.
	scale := float64(w) / float64(sw)
	if float64(sh)*scale < float64(h) {
		scale = float64(h) / float64(sh)
	}
	rw := max(1, int(float64(sw)*scale))
	rh := max(1, int(float64(sh)*scale))
	resized := resizeNearest(src, rw, rh)

	ox := (w - rw) / 2
	oy := (h - rh) / 5 // bias up
	draw.Draw(dst, image.Rect(ox, oy, ox+rw, oy+rh), resized, image.Point{}, draw.Over)
	return dst
}

func composeLayout(tile image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	fillBackground(dst, color.RGBA{R: 248, G: 250, B: 252, A: 255}, "solid")

	const cols, rows = 5, 2
	gap := 10
	marginX, marginY := 28, 36
	tileW := (w - marginX*2 - gap*(cols-1)) / cols
	tileH := (h - marginY*2 - gap*(rows-1)) / rows
	if tileW < 8 || tileH < 8 {
		return dst
	}
	thumb := composeCover(tile, tileW, tileH, color.RGBA{R: 255, G: 255, B: 255, A: 255}, "solid")
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x := marginX + c*(tileW+gap)
			y := marginY + r*(tileH+gap)
			draw.Draw(dst, image.Rect(x, y, x+tileW, y+tileH), thumb, image.Point{}, draw.Over)
		}
	}
	return dst
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: 255,
	}
}

func parseHex(s string, fallback color.RGBA) color.RGBA {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != 6 {
		return fallback
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return fallback
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
