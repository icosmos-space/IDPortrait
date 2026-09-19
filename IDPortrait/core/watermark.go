package core

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	fontOnce sync.Once
	fontData []byte
	fontColl bool
	fontErr  error
)

func cjkFont() (*opentype.Font, error) {
	fontOnce.Do(func() {
		paths := []string{
			`C:\Windows\Fonts\msyh.ttc`,
			`C:\Windows\Fonts\msyh.ttf`,
			`C:\Windows\Fonts\simhei.ttf`,
			`C:\Windows\Fonts\simsun.ttc`,
			`C:\Windows\Fonts\Deng.ttf`,
		}
		var path string
		for _, p := range paths {
			b, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			fontData = b
			path = p
			break
		}
		if fontData == nil {
			fontErr = fmt.Errorf("未找到可用的中文字体")
			return
		}
		fontColl = strings.HasSuffix(strings.ToLower(path), ".ttc")
	})
	if fontErr != nil || fontData == nil {
		return nil, fontErr
	}
	if fontColl {
		col, err := opentype.ParseCollection(fontData)
		if err != nil {
			return nil, err
		}
		return col.Font(0)
	}
	return opentype.Parse(fontData)
}

// paintWatermark tiles the text on a diagonal, matching Hivision's striped watermark.
func paintWatermark(src *image.RGBA, p GenerateParams) (*image.RGBA, error) {
	text := strings.TrimSpace(p.WatermarkText)
	if text == "" || !p.EnableWatermark {
		return src, nil
	}
	size := p.WatermarkFontSize
	if size < 8 {
		size = 18
	}
	faceFont, err := cjkFont()
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(faceFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}
	defer face.Close()

	mark := drawTextMark(face, text, parseHex(p.WatermarkColor, color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	mark = cropAlpha(mark)
	setOpacity(mark, p.WatermarkOpacity)

	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	c := int(math.Sqrt(float64(w*w + h*h)))
	if c < 1 {
		c = 1
	}
	sheet := image.NewNRGBA(image.Rect(0, 0, c, c))
	space := int(p.WatermarkSpacing)
	if space < 8 {
		space = 75
	}
	mw, mh := mark.Bounds().Dx(), mark.Bounds().Dy()
	stepX := mw + space
	stepY := mh + space
	if stepX < 1 {
		stepX = 1
	}
	if stepY < 1 {
		stepY = 1
	}
	row := 0
	for y := 0; y < c; y += stepY {
		x := 0
		if row%2 == 1 {
			x = -stepX / 2
		}
		for ; x < c; x += stepX {
			pasteNRGBA(sheet, mark, x, y)
		}
		row++
	}
	sheet = rotateCanvas(sheet, p.WatermarkAngle)

	out := image.NewRGBA(src.Bounds())
	copy(out.Pix, src.Pix)
	ox := (w - sheet.Bounds().Dx()) / 2
	oy := (h - sheet.Bounds().Dy()) / 2
	blendOver(out, sheet, ox, oy)
	return out, nil
}

func drawTextMark(face font.Face, text string, col color.RGBA) *image.NRGBA {
	b, _ := font.BoundString(face, text)
	w := (b.Max.X - b.Min.X).Ceil() + 4
	h := face.Metrics().Height.Ceil() + 4
	if w < 2 {
		w = 2
	}
	if h < 2 {
		h = 2
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.NRGBA{R: col.R, G: col.G, B: col.B, A: 255}),
		Face: face,
		Dot:  fixed.P(2-b.Min.X.Floor(), face.Metrics().Ascent.Ceil()),
	}
	d.DrawString(text)
	return dst
}

func cropAlpha(src *image.NRGBA) *image.NRGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	minX, minY, maxX, maxY := w, h, -1, -1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if src.Pix[src.PixOffset(x, y)+3] == 0 {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if maxX < 0 {
		return src
	}
	dst := image.NewNRGBA(image.Rect(0, 0, maxX-minX+1, maxY-minY+1))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			copy4(dst, x-minX, y-minY, src, x, y)
		}
	}
	return dst
}

func setOpacity(img *image.NRGBA, opacity float64) {
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = uint8(float64(img.Pix[i]) * opacity)
	}
}

func pasteNRGBA(dst, src *image.NRGBA, ox, oy int) {
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
			a := src.Pix[si+3]
			if a == 0 {
				continue
			}
			di := dst.PixOffset(dx, dy)
			if a == 255 || dst.Pix[di+3] == 0 {
				copy(dst.Pix[di:di+4], src.Pix[si:si+4])
				continue
			}
			af := float32(a) / 255
			bf := float32(dst.Pix[di+3]) / 255
			outA := af + bf*(1-af)
			if outA <= 0 {
				continue
			}
			for c := 0; c < 3; c++ {
				dst.Pix[di+c] = uint8((float32(src.Pix[si+c])*af + float32(dst.Pix[di+c])*bf*(1-af)) / outA)
			}
			dst.Pix[di+3] = uint8(outA * 255)
		}
	}
}

func blendOver(dst *image.RGBA, src *image.NRGBA, ox, oy int) {
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
			a := src.Pix[si+3]
			if a == 0 {
				continue
			}
			di := dst.PixOffset(dx, dy)
			af := float32(a) / 255
			dst.Pix[di] = uint8(float32(src.Pix[si])*af + float32(dst.Pix[di])*(1-af))
			dst.Pix[di+1] = uint8(float32(src.Pix[si+1])*af + float32(dst.Pix[di+1])*(1-af))
			dst.Pix[di+2] = uint8(float32(src.Pix[si+2])*af + float32(dst.Pix[di+2])*(1-af))
			dst.Pix[di+3] = 255
		}
	}
}

// rotateCanvas rotates around the center and keeps the canvas size, like PIL rotate(expand=False).
func rotateCanvas(src *image.NRGBA, deg float64) *image.NRGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	rad := deg * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)
	cx := float64(w-1) / 2
	cy := float64(h-1) / 2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			sx := cos*dx + sin*dy + cx
			sy := -sin*dx + cos*dy + cy
			if sx < 0 || sy < 0 || sx > float64(w-1) || sy > float64(h-1) {
				continue
			}
			r, g, b, a := sampleNRGBA4(src, sx, sy)
			i := dst.PixOffset(x, y)
			dst.Pix[i] = uint8(clampFloat(float64(r), 0, 255))
			dst.Pix[i+1] = uint8(clampFloat(float64(g), 0, 255))
			dst.Pix[i+2] = uint8(clampFloat(float64(b), 0, 255))
			dst.Pix[i+3] = uint8(clampFloat(float64(a), 0, 255))
		}
	}
	return dst
}
