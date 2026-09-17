package core

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"
)

// Engine hosts ID-photo processing algorithms.
// Current implementation loads real source images and returns placeholder
// generation outputs so HTTP / Wails paths share the same service contract.
type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) LoadImage(path string) (*LoadImageResult, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("empty image path")
	}

	var dataURL string
	var bounds image.Rectangle

	switch {
	case strings.HasPrefix(path, "data:image/"):
		dataURL = path
		img, err := decodeDataURL(path)
		if err != nil {
			return nil, err
		}
		bounds = img.Bounds()
	default:
		img, format, err := loadImageFile(path)
		if err != nil {
			return nil, err
		}
		dataURL, err = encodeDataURL(img, format)
		if err != nil {
			return nil, err
		}
		bounds = img.Bounds()
	}

	w := bounds.Dx()
	h := bounds.Dy()
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
	bg := parseHex(p.BgColor, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	mode := p.BgMode
	if mode == "" {
		mode = "solid"
	}
	paper := p.PaperSize
	if paper == "" {
		paper = "6inch"
	}

	origin := placeholderPNG(360, 480, color.RGBA{R: 241, G: 245, B: 249, A: 255}, "原图")
	single := placeholderPNGMode(360, 480, bg, mode, "单张照片")
	idphoto := placeholderPNGMode(295, 413, bg, mode, "证件照")
	social := placeholderPNGMode(400, 400, bg, mode, "社交照")
	layout := placeholderPNGMode(600, 400, color.RGBA{R: 248, G: 250, B: 252, A: 255}, "solid", paper+"排版照")

	return &GenerateResult{
		OriginImg: origin,
		ResultImg: idphoto,
		Results: ResultBundle{
			Single:  single,
			Layout:  layout,
			Social:  social,
			IDPhoto: idphoto,
		},
		FaceBox:   []float64{90, 80, 270, 300},
		Landmarks: []float64{140, 160, 220, 160, 180, 210, 150, 250, 210, 250},
		Report: Report{
			FaceOK:    true,
			FaceScore: 0.94,
		},
	}, nil
}

func (e *Engine) Export(dir string, _ ExportOptions) (*ExportResult, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("export dir is empty")
	}
	return &ExportResult{OK: true, Dir: dir}, nil
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

func encodeDataURL(img image.Image, format string) (string, error) {
	var buf bytes.Buffer
	mime := "image/png"
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		mime = "image/jpeg"
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 92}); err != nil {
			return "", err
		}
	default:
		if err := png.Encode(&buf, img); err != nil {
			return "", err
		}
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func placeholderPNG(w, h int, bg color.RGBA, label string) string {
	return placeholderPNGMode(w, h, bg, "solid", label)
}

func placeholderPNGMode(w, h int, bg color.RGBA, mode, label string) string {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := bg
			switch mode {
			case "vertical":
				t := float64(y) / float64(max(h-1, 1))
				c = mix(bg, white, t)
			case "radial":
				cx, cy := float64(w)/2, float64(h)*0.42
				dx, dy := float64(x)-cx, float64(y)-cy
				t := (dx*dx + dy*dy) / (float64(w*w+h*h) * 0.25)
				if t > 1 {
					t = 1
				}
				c = mix(bg, white, t)
			}
			img.Set(x, y, c)
		}
	}
	for y := h / 5; y < h*4/5; y++ {
		for x := w / 3; x < w*2/3; x++ {
			img.Set(x, y, color.RGBA{R: 71, G: 85, B: 105, A: 255})
		}
	}
	_ = label
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
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
