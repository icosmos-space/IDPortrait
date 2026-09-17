package core

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
)

// Engine hosts ID-photo processing algorithms.
// Current implementation is a scaffold that produces placeholder images
// so HTTP / Wails paths can share the same service contract.
type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) LoadImage(path string) (*LoadImageResult, error) {
	label := path
	if i := strings.LastIndexAny(path, `/\`); i >= 0 {
		label = path[i+1:]
	}
	if label == "" {
		label = "photo"
	}
	img := placeholderPNG(360, 480, color.RGBA{R: 241, G: 245, B: 249, A: 255}, label)
	return &LoadImageResult{
		ImgBase64: img,
		FaceBox:   []float64{90, 80, 270, 300},
		Landmarks: []float64{140, 160, 220, 160, 180, 210, 150, 250, 210, 250},
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
	// simple silhouette block
	for y := h / 5; y < h*4/5; y++ {
		for x := w/3; x < w*2/3; x++ {
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
