package core

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"sync"
)

type templateAnchor struct {
	LeftTop     [2]float64 `json:"left_top"`
	RightTop    [2]float64 `json:"right_top"`
	LeftBottom  [2]float64 `json:"left_bottom"`
	RightBottom [2]float64 `json:"right_bottom"`
	Rotation    float64    `json:"rotation"`
}

type templateSpec struct {
	Width        int            `json:"width"`
	Height       int            `json:"height"`
	AnchorPoints templateAnchor `json:"anchor_points"`
}

var templatePNG sync.Map

func socialTemplates(photo *image.RGBA) (first, second *image.RGBA, err error) {
	first, err = renderTemplate("template_1", photo)
	if err != nil {
		return nil, nil, err
	}
	second, err = renderTemplate("template_2", photo)
	if err != nil {
		return nil, nil, err
	}
	return first, second, nil
}

func renderTemplate(name string, photo *image.RGBA) (*image.RGBA, error) {
	root, err := runtimeRoot()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, "templates")
	raw, err := os.ReadFile(filepath.Join(dir, "template_config.json"))
	if err != nil {
		return nil, fmt.Errorf("读取社交模板配置失败: %w", err)
	}
	var all map[string]templateSpec
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, err
	}
	spec, ok := all[name]
	if !ok {
		return nil, fmt.Errorf("缺少社交模板 %s", name)
	}
	tpl, err := loadTemplatePNG(filepath.Join(dir, name+".png"))
	if err != nil {
		return nil, err
	}

	anchor := spec.AnchorPoints
	var holeW, holeH float64
	if anchor.Rotation < 0 {
		holeH = anchor.RightBottom[1] - anchor.LeftTop[1]
		holeW = anchor.RightTop[0] - anchor.LeftBottom[0]
	} else {
		holeH = anchor.LeftTop[1] - anchor.RightBottom[1]
		holeW = anchor.LeftBottom[0] - anchor.RightTop[0]
	}
	if holeW < 1 {
		holeW = 1
	}
	if holeH < 1 {
		holeH = 1
	}

	turned := rotateBound(rgbaToNRGBA(photo), -anchor.Rotation)
	rw, rh := turned.Bounds().Dx(), turned.Bounds().Dy()
	scale := math.Max(holeW/float64(rw), holeH/float64(rh))
	nw := max(1, int(math.Round(float64(rw)*scale)))
	nh := max(1, int(math.Round(float64(rh)*scale)))
	resized := resizeNRGBA(turned, nw, nh)

	canvas := image.NewRGBA(image.Rect(0, 0, spec.Width, spec.Height))
	for i := 0; i < len(canvas.Pix); i += 4 {
		canvas.Pix[i] = 255
		canvas.Pix[i+1] = 255
		canvas.Pix[i+2] = 255
		canvas.Pix[i+3] = 255
	}
	pasteX := int(math.Round(anchor.LeftBottom[0]))
	pasteY := int(math.Round(anchor.LeftTop[1]))
	pasteClipped(canvas, resized, pasteX, pasteY)
	overlayTemplate(canvas, tpl)
	return canvas, nil
}

func loadTemplatePNG(path string) (*image.NRGBA, error) {
	if v, ok := templatePNG.Load(path); ok {
		return v.(*image.NRGBA), nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("读取社交模板失败: %w", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	n := toNRGBA(img)
	templatePNG.Store(path, n)
	return n, nil
}

func rgbaToNRGBA(src *image.RGBA) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	copy(dst.Pix, src.Pix)
	return dst
}

func pasteClipped(dst *image.RGBA, src *image.NRGBA, ox, oy int) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	dw, dh := dst.Bounds().Dx(), dst.Bounds().Dy()
	for y := 0; y < sh; y++ {
		dy := oy + y
		if dy < 0 || dy >= dh {
			continue
		}
		for x := 0; x < sw; x++ {
			dx := ox + x
			if dx < 0 || dx >= dw {
				continue
			}
			si := src.PixOffset(x, y)
			if src.Pix[si+3] < 8 {
				continue
			}
			di := dst.PixOffset(dx, dy)
			copy(dst.Pix[di:di+3], src.Pix[si:si+3])
			dst.Pix[di+3] = 255
		}
	}
}

func overlayTemplate(dst *image.RGBA, tpl *image.NRGBA) {
	w := min(dst.Bounds().Dx(), tpl.Bounds().Dx())
	h := min(dst.Bounds().Dy(), tpl.Bounds().Dy())
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si := tpl.PixOffset(x, y)
			a := tpl.Pix[si+3]
			if a == 0 {
				continue
			}
			di := dst.PixOffset(x, y)
			af := float32(a) / 255
			dst.Pix[di] = uint8(float32(tpl.Pix[si])*af + float32(dst.Pix[di])*(1-af))
			dst.Pix[di+1] = uint8(float32(tpl.Pix[si+1])*af + float32(dst.Pix[di+1])*(1-af))
			dst.Pix[di+2] = uint8(float32(tpl.Pix[si+2])*af + float32(dst.Pix[di+2])*(1-af))
			dst.Pix[di+3] = 255
		}
	}
}
