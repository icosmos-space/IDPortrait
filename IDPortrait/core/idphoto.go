package core

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
)

type idBundle struct {
	matting *image.NRGBA
	hd      *image.NRGBA
	std     *image.NRGBA
	face    faceHit
}

func makeIDPhoto(src *image.NRGBA, p GenerateParams) (*idBundle, error) {
	work := src
	faces, err := detectFaces(work)
	if err != nil {
		return nil, err
	}
	if len(faces) == 0 {
		return nil, fmt.Errorf("未检测到人脸，已拒绝")
	}
	if len(faces) > 1 {
		return nil, fmt.Errorf("检测到多张人脸，已拒绝")
	}
	face := faces[0]

	// MODNet is a portrait model. On ID-card scans the face is tiny and the
	// network latches onto the whole card — zoom to a head-shoulders ROI first.
	var cropped bool
	work, face, cropped = cropPortraitROI(work, face)

	if roll := eyeRoll(face); math.Abs(roll) > 2 {
		turned := rotateBound(work, -roll)
		again, err := detectFaces(turned)
		if err == nil && len(again) == 1 && math.Abs(eyeRoll(again[0])) < math.Abs(roll) {
			work = turned
			face = again[0]
		}
	}

	cut, err := portraitMatting(work, p.MattingModel)
	if err != nil {
		return nil, err
	}
	if cropped {
		// Printed faces leave floating card-pattern crumbs; keep only near the face.
		applyFaceGate(cut, face)
	}
	featherAlpha(cut, p.MaskFeather)
	applyBeauty(cut, p.BeautyStrength, p.SkinBright, p.EyeSharp)

	outW, outH := photoPixels(p)
	std := adjustIDPhoto(cut, face, outW, outH, p.FaceRatio, p.HeadTopDistance)
	hd := resizeByMin(std, max(600, outW))
	return &idBundle{matting: cut, hd: hd, std: std, face: face}, nil
}

func eyeRoll(f faceHit) float64 {
	return math.Atan2(f.kps[3]-f.kps[1], f.kps[2]-f.kps[0]) * 180 / math.Pi
}

// cropPortraitROI zooms to a head-and-shoulders window when the face is a
// small fraction of the frame (ID-card prints). Face coords are remapped.
func cropPortraitROI(src *image.NRGBA, face faceHit) (*image.NRGBA, faceHit, bool) {
	if src == nil {
		return src, face, false
	}
	iw, ih := src.Bounds().Dx(), src.Bounds().Dy()
	if iw < 32 || ih < 32 || face.w < 8 || face.h < 8 {
		return src, face, false
	}
	faceFrac := (face.w * face.h) / float64(iw*ih)
	if faceFrac >= 0.08 {
		return src, face, false
	}

	fw, fh := face.w, face.h
	cx := face.x + fw/2
	cy := face.y + fh/2
	top := cy - fh*1.35
	bottom := cy + fh*2.1
	height := bottom - top
	width := height * 3 / 4
	if width < fw*2.4 {
		width = fw * 2.4
	}
	left := cx - width/2
	right := cx + width/2

	x1 := int(math.Floor(left))
	y1 := int(math.Floor(top))
	x2 := int(math.Ceil(right))
	y2 := int(math.Ceil(bottom))
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > iw {
		x2 = iw
	}
	if y2 > ih {
		y2 = ih
	}
	if x2-x1 < 32 || y2-y1 < 32 {
		return src, face, false
	}

	crop := cutPad(src, x1, y1, x2, y2)
	local := face
	local.x = face.x - float64(x1)
	local.y = face.y - float64(y1)
	for i := 0; i < 5; i++ {
		local.kps[i*2] -= float64(x1)
		local.kps[i*2+1] -= float64(y1)
	}
	return crop, local, true
}

func applyFaceGate(img *image.NRGBA, face faceHit) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	cx := face.x + face.w/2
	cy := face.y + face.h*0.55
	rx := math.Max(face.w*1.4, float64(w)*0.45)
	ry := math.Max(face.h*1.9, float64(h)*0.5)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			nx := (float64(x) - cx) / rx
			ny := (float64(y) - cy) / ry
			d := math.Sqrt(nx*nx + ny*ny)
			var g float64
			switch {
			case d <= 0.82:
				g = 1
			case d >= 1.08:
				g = 0
			default:
				t := (d - 0.82) / (1.08 - 0.82)
				g = 1 - t*t*(3-2*t)
			}
			i := img.PixOffset(x, y)
			img.Pix[i+3] = uint8(float64(img.Pix[i+3]) * g)
		}
	}
}

func photoPixels(p GenerateParams) (int, int) {
	if p.Template == "custom" {
		w := mmToPx(p.CustomWidth)
		h := mmToPx(p.CustomHeight)
		if w < 2 {
			w = 295
		}
		if h < 2 {
			h = 413
		}
		return w, h
	}
	for _, spec := range BuiltinPhotoSpecs() {
		if spec.Value != p.Template {
			continue
		}
		if spec.Unit == "px" && spec.WidthPx > 0 && spec.HeightPx > 0 {
			return spec.WidthPx, spec.HeightPx
		}
		return mmToPx(spec.WidthMM), mmToPx(spec.HeightMM)
	}
	return 295, 413
}

func mmToPx(mm float64) int {
	if mm <= 0 {
		return 1
	}
	return max(1, int(math.Round(mm/25.4*300)))
}

func paperPixels(id string) (int, int) {
	for _, spec := range BuiltinPaperSpecs() {
		if spec.Value == id {
			return mmToPx(spec.WidthMM), mmToPx(spec.HeightMM)
		}
	}
	return mmToPx(102), mmToPx(152)
}

// adjustIDPhoto crops a transparent portrait the way HivisionIDPhotos does:
// face area sets the crop, then the head-top gap is nudged into range.
func adjustIDPhoto(src *image.NRGBA, face faceHit, outW, outH int, faceRatio, topDist float64) *image.NRGBA {
	if outW < 2 {
		outW = 295
	}
	if outH < 2 {
		outH = 413
	}
	measure := faceRatio
	if measure <= 0 {
		measure = 0.2
	}
	measure = clampFloat(measure, 0.1, 0.5)
	topMax := topDist
	if topMax <= 0 {
		topMax = 0.12
	}
	topMax = clampFloat(topMax, 0.02, 0.5)
	topMin := topMax - 0.02

	fw, fh := face.w, face.h
	if fw < 2 {
		fw = 2
	}
	if fh < 2 {
		fh = 2
	}
	faceMeasure := fw * fh
	cropMeasure := faceMeasure / measure
	stdH, stdW := float64(outH), float64(outW)
	resizeSingle := math.Sqrt(cropMeasure / (stdH * stdW))
	cropH := max(2, int(stdH*resizeSingle))
	cropW := max(2, int(stdW*resizeSingle))
	cx := face.x + face.w/2
	cy := face.y + face.h/2
	x1 := int(cx - float64(cropW)/2)
	y1 := int(cy - float64(cropH)*0.45)
	x2 := x1 + cropW
	y2 := y1 + cropH

	cut := resizeNRGBA(cutPad(src, x1, y1, x2, y2), cropW, cropH)
	top, _, left, right := alphaMargins(cut)
	statusLR := 0
	cutTop := 0
	if left > 0 || right > 0 {
		statusLR = 1
		cutTop = int(float64(left+right) * stdH / stdW / 2)
	}
	statusTop, move := detectHeadGap(top-cutTop, cropH, topMax, topMin)
	var result *image.NRGBA
	if statusLR == 0 && statusTop == 0 {
		result = cut
	} else {
		result = cutPad(src,
			x1+left,
			y1+cutTop+statusTop*move,
			x2-right,
			y2-cutTop+statusTop*move,
		)
	}
	result = pullToBottom(result)
	return resizeNRGBA(result, outW, outH)
}

func detectHeadGap(gap, cropH int, maxRatio, minRatio float64) (int, int) {
	if cropH < 1 {
		return 0, 0
	}
	value := float64(gap) / float64(cropH)
	if value >= minRatio && value <= maxRatio {
		return 0, 0
	}
	if value > maxRatio {
		return 1, int((value - maxRatio) * float64(cropH))
	}
	return -1, int((minRatio - value) * float64(cropH))
}

func cutPad(src *image.NRGBA, x1, y1, x2, y2 int) *image.NRGBA {
	cw, ch := x2-x1, y2-y1
	if cw < 1 {
		cw = 1
	}
	if ch < 1 {
		ch = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, cw, ch))
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	for y := 0; y < ch; y++ {
		sy := y1 + y
		if sy < 0 || sy >= sh {
			continue
		}
		for x := 0; x < cw; x++ {
			sx := x1 + x
			if sx < 0 || sx >= sw {
				continue
			}
			copy4(dst, x, y, src, sx, sy)
		}
	}
	return dst
}

func alphaMargins(img *image.NRGBA) (top, bottom, left, right int) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	minX, minY, maxX, maxY := w, h, -1, -1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img.Pix[img.PixOffset(x, y)+3] < 20 {
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
		return 0, 0, 0, 0
	}
	return minY, h - 1 - maxY, minX, w - 1 - maxX
}

func pullToBottom(img *image.NRGBA) *image.NRGBA {
	_, bottom, _, _ := alphaMargins(img)
	if bottom <= 0 {
		return img
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h-bottom; y++ {
		copy(dst.Pix[dst.PixOffset(0, y+bottom):dst.PixOffset(0, y+bottom)+w*4], img.Pix[img.PixOffset(0, y):img.PixOffset(0, y)+w*4])
	}
	return dst
}

func resizeByMin(img *image.NRGBA, minSide int) *image.NRGBA {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if min(w, h) >= minSide {
		return img
	}
	var nw, nh int
	if h >= w {
		nw = minSide
		nh = max(1, h*minSide/w)
	} else {
		nh = minSide
		nw = max(1, w*minSide/h)
	}
	return resizeNRGBA(img, nw, nh)
}

func compositeOn(fg *image.NRGBA, bg color.RGBA, mode string) *image.RGBA {
	w, h := fg.Bounds().Dx(), fg.Bounds().Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	fillBackground(dst, bg, mode)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := fg.PixOffset(x, y)
			a := fg.Pix[i+3]
			if a == 0 {
				continue
			}
			dr := dst.Pix[y*dst.Stride+x*4]
			dg := dst.Pix[y*dst.Stride+x*4+1]
			db := dst.Pix[y*dst.Stride+x*4+2]
			af := float32(a) / 255
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(float32(fg.Pix[i])*af + float32(dr)*(1-af)),
				G: uint8(float32(fg.Pix[i+1])*af + float32(dg)*(1-af)),
				B: uint8(float32(fg.Pix[i+2])*af + float32(db)*(1-af)),
				A: 255,
			})
		}
	}
	return dst
}

// Hivision layout canvas reference (6-inch landscape at their DPI).
const (
	hivisionLayoutW = 1795
	hivisionLayoutH = 1205
	hivisionGap     = 30
	hivisionSideW   = 70
	hivisionSideH   = 50
)

type layoutMode struct {
	cols, rows int
	rotate     bool
	blockW     int
	blockH     int
}

// judgeLayout mirrors HivisionIDPhotos layout_calculator.judge_layout.
// photoW/photoH are the upright ID-photo size; interval/limit are in pixels.
func judgeLayout(photoW, photoH, intervalW, intervalH, limitW, limitH int) layoutMode {
	fit := func(cellW, cellH int) (cols, rows, blockW, blockH int) {
		rows = 0
		blockH = cellH
		for i := 1; i <= 3; i++ {
			h := cellH*i + intervalH*(i-1)
			if h < limitH {
				blockH = h
				rows = i
			} else {
				break
			}
		}
		cols = 0
		blockW = cellW
		for j := 1; j <= 8; j++ {
			w := cellW*j + intervalW*(j-1)
			if w < limitW {
				blockW = w
				cols = j
			} else {
				break
			}
		}
		return cols, rows, blockW, blockH
	}

	c1, r1, bw1, bh1 := fit(photoW, photoH)
	c2, r2, bw2, bh2 := fit(photoH, photoW) // transposed cell
	n1, n2 := c1*r1, c2*r2
	if n2 > n1 {
		return layoutMode{cols: c2, rows: r2, rotate: true, blockW: bw2, blockH: bh2}
	}
	return layoutMode{cols: c1, rows: r1, rotate: false, blockW: bw1, blockH: bh1}
}

func layoutSheet(tile *image.RGBA, paperW, paperH int, cropLine bool) *image.RGBA {
	photoW, photoH := tile.Bounds().Dx(), tile.Bounds().Dy()
	if paperW < 2 {
		paperW = hivisionLayoutW
	}
	if paperH < 2 {
		paperH = hivisionLayoutH
	}
	if paperW < photoW+40 {
		paperW = photoW + 40
	}
	if paperH < photoH+40 {
		paperH = photoH + 40
	}

	// Scale Hivision margins/gaps to the current paper size.
	sidesW := max(1, int(math.Round(float64(hivisionSideW)*float64(paperW)/float64(hivisionLayoutW))))
	sidesH := max(1, int(math.Round(float64(hivisionSideH)*float64(paperH)/float64(hivisionLayoutH))))
	gapW := max(1, int(math.Round(float64(hivisionGap)*float64(paperW)/float64(hivisionLayoutW))))
	gapH := max(1, int(math.Round(float64(hivisionGap)*float64(paperH)/float64(hivisionLayoutH))))
	limitW := paperW - 2*sidesW
	limitH := paperH - 2*sidesH
	if limitW < photoW {
		limitW = photoW
	}
	if limitH < photoH {
		limitH = photoH
	}

	mode := judgeLayout(photoW, photoH, gapW, gapH, limitW, limitH)
	if mode.cols < 1 || mode.rows < 1 {
		mode = layoutMode{cols: 1, rows: 1, rotate: false, blockW: photoW, blockH: photoH}
	}

	photo := tile
	cellW, cellH := photoW, photoH
	if mode.rotate {
		photo = transposeFlipVertical(tile)
		cellW, cellH = photo.Bounds().Dx(), photo.Bounds().Dy()
		// Recompute block size for the rotated cell (matches Hivision after swap).
		mode.blockW = cellW*mode.cols + gapW*(mode.cols-1)
		mode.blockH = cellH*mode.rows + gapH*(mode.rows-1)
	}

	dst := image.NewRGBA(image.Rect(0, 0, paperW, paperH))
	fillBackground(dst, color.RGBA{R: 255, G: 255, B: 255, A: 255}, "solid")
	x0 := (paperW - mode.blockW) / 2
	y0 := (paperH - mode.blockH) / 2
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	positions := make([][2]int, 0, mode.rows*mode.cols)
	for r := 0; r < mode.rows; r++ {
		for c := 0; c < mode.cols; c++ {
			x := x0 + c*(cellW+gapW)
			y := y0 + r*(cellH+gapH)
			if x+cellW > paperW || y+cellH > paperH {
				continue
			}
			drawOver(dst, photo, x, y)
			positions = append(positions, [2]int{x, y})
		}
	}
	// Hivision crop_line: light-gray guides along each photo edge, full paper span.
	if cropLine {
		drawLayoutCropLines(dst, positions, cellW, cellH)
	}
	return dst
}

func drawLayoutCropLines(dst *image.RGBA, positions [][2]int, cellW, cellH int) {
	if len(positions) == 0 {
		return
	}
	line := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	paperW, paperH := dst.Bounds().Dx(), dst.Bounds().Dy()
	vert := map[int]struct{}{}
	horiz := map[int]struct{}{}
	for _, p := range positions {
		x, y := p[0], p[1]
		vert[x] = struct{}{}
		vert[x+cellW] = struct{}{}
		horiz[y] = struct{}{}
		horiz[y+cellH] = struct{}{}
	}
	for x := range vert {
		if x < 0 || x >= paperW {
			continue
		}
		for y := 0; y < paperH; y++ {
			dst.SetRGBA(x, y, line)
		}
	}
	for y := range horiz {
		if y < 0 || y >= paperH {
			continue
		}
		for x := 0; x < paperW; x++ {
			dst.SetRGBA(x, y, line)
		}
	}
}

func drawOver(dst *image.RGBA, src *image.RGBA, ox, oy int) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			si := src.PixOffset(x, y)
			di := dst.PixOffset(ox+x, oy+y)
			copy(dst.Pix[di:di+4], src.Pix[si:si+4])
		}
	}
}

// transposeFlipVertical matches Hivision layout rotate:
// cv2.transpose then cv2.flip(..., 0).
func transposeFlipVertical(src *image.RGBA) *image.RGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	// After transpose: width=h, height=w; then vertical flip on that image.
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// transpose: (x,y) -> (y,x); vertical flip: (y,x) -> (y, w-1-x)
			si := src.PixOffset(x, y)
			di := dst.PixOffset(y, w-1-x)
			copy(dst.Pix[di:di+4], src.Pix[si:si+4])
		}
	}
	return dst
}

func squareSocial(tile *image.RGBA, bg color.RGBA, mode string) *image.RGBA {
	side := max(tile.Bounds().Dx(), tile.Bounds().Dy())
	dst := image.NewRGBA(image.Rect(0, 0, side, side))
	fillBackground(dst, bg, mode)
	ox := (side - tile.Bounds().Dx()) / 2
	oy := (side - tile.Bounds().Dy()) / 2
	drawOver(dst, tile, ox, oy)
	return dst
}

func applyBeauty(img *image.NRGBA, strength, bright, sharp float64) {
	if strength < 0 {
		strength = 0
	}
	if strength > 0.4 {
		strength = 0.4
	}
	add := strength*26 + clampFloat(bright, 0, 1)*18
	contrast := 1 + strength*0.12
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := img.PixOffset(x, y)
			if img.Pix[i+3] < 8 {
				continue
			}
			for c := 0; c < 3; c++ {
				v := (float64(img.Pix[i+c])-128)*contrast + 128 + add
				if v < 0 {
					v = 0
				}
				if v > 255 {
					v = 255
				}
				img.Pix[i+c] = uint8(v)
			}
		}
	}
	if sharp > 0.04 {
		sharpen(img, sharp*0.35)
	}
}

func sharpen(img *image.NRGBA, amount float64) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w < 3 || h < 3 {
		return
	}
	orig := append([]uint8(nil), img.Pix...)
	stride := img.Stride
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			i := y*stride + x*4
			if orig[i+3] < 8 {
				continue
			}
			for c := 0; c < 3; c++ {
				center := float64(orig[i+c])
				blur := (float64(orig[i-stride+c]) + float64(orig[i+stride+c]) + float64(orig[i-4+c]) + float64(orig[i+4+c]) + center) / 5
				v := center + (center-blur)*amount
				if v < 0 {
					v = 0
				}
				if v > 255 {
					v = 255
				}
				img.Pix[i+c] = uint8(v)
			}
		}
	}
}

func featherAlpha(img *image.NRGBA, amount float64) {
	radius := int(clampFloat(amount, 0, 0.6) * 4)
	if radius < 1 {
		return
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	next := make([]uint8, len(img.Pix))
	copy(next, img.Pix)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sum, n := 0, 0
			for dy := -radius; dy <= radius; dy++ {
				yy := y + dy
				if yy < 0 || yy >= h {
					continue
				}
				for dx := -radius; dx <= radius; dx++ {
					xx := x + dx
					if xx < 0 || xx >= w {
						continue
					}
					sum += int(img.Pix[img.PixOffset(xx, yy)+3])
					n++
				}
			}
			if n > 0 {
				next[img.PixOffset(x, y)+3] = uint8(sum / n)
			}
		}
	}
	img.Pix = next
}

func applyAlpha(src *image.NRGBA, alpha []float32) *image.NRGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := src.PixOffset(x, y)
			di := dst.PixOffset(x, y)
			copy(dst.Pix[di:di+3], src.Pix[i:i+3])
			a := alpha[y*w+x]
			if a < 0 {
				a = 0
			}
			if a > 1 {
				a = 1
			}
			dst.Pix[di+3] = uint8(a * 255)
		}
	}
	return dst
}

func resizeGray(src []float32, sw, sh, dw, dh int) []float32 {
	dst := make([]float32, dw*dh)
	if sw < 1 || sh < 1 {
		return dst
	}
	for y := 0; y < dh; y++ {
		fy := (float64(y)+0.5)*float64(sh)/float64(dh) - 0.5
		for x := 0; x < dw; x++ {
			fx := (float64(x)+0.5)*float64(sw)/float64(dw) - 0.5
			dst[y*dw+x] = sampleGray(src, sw, sh, fx, fy)
		}
	}
	return dst
}

func sampleGray(src []float32, w, h int, x, y float64) float32 {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > float64(w-1) {
		x = float64(w - 1)
	}
	if y > float64(h-1) {
		y = float64(h - 1)
	}
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1, y1 := x0+1, y0+1
	if x1 >= w {
		x1 = w - 1
	}
	if y1 >= h {
		y1 = h - 1
	}
	dx := float32(x - float64(x0))
	dy := float32(y - float64(y0))
	a := src[y0*w+x0]
	b := src[y0*w+x1]
	c := src[y1*w+x0]
	d := src[y1*w+x1]
	return a + (b-a)*dx + ((c+(d-c)*dx)-(a+(b-a)*dx))*dy
}

func fillMatteHoles(alpha []float32, w, h int) []float32 {
	n := w * h
	solid := make([]byte, n)
	for i := 0; i < n; i++ {
		if alpha[i] >= 0.5 {
			solid[i] = 1
		}
	}
	seen := make([]byte, n)
	best := 0
	bestID := byte(0)
	id := byte(2)
	var stack []int
	for i := 0; i < n; i++ {
		if solid[i] == 0 || seen[i] != 0 {
			continue
		}
		stack = stack[:0]
		stack = append(stack, i)
		seen[i] = id
		count := 0
		for len(stack) > 0 {
			cur := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			count++
			x, y := cur%w, cur/w
			for _, nb := range [4]int{cur - 1, cur + 1, cur - w, cur + w} {
				if nb < 0 || nb >= n || seen[nb] != 0 || solid[nb] == 0 {
					continue
				}
				nx := nb % w
				if (nb == cur-1 && nx != x-1) || (nb == cur+1 && nx != x+1) {
					continue
				}
				if nb == cur-w && y == 0 || nb == cur+w && y == h-1 {
					continue
				}
				seen[nb] = id
				stack = append(stack, nb)
			}
		}
		if count > best {
			best = count
			bestID = id
		}
		if id == 255 {
			break
		}
		id++
	}
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		if seen[i] == bestID {
			out[i] = alpha[i]
		}
	}
	// Fill only small enclosed holes (hair gaps, etc.). Large "holes" usually
	// mean MODNet kept an ID-card / object outline — filling them would paint
	// the whole card opaque. Hivision hollow_out is contour-based; this guard
	// approximates that safety.
	bg := make([]byte, n)
	stack = stack[:0]
	push := func(i int) {
		if i < 0 || i >= n || bg[i] != 0 || out[i] >= 0.5 {
			return
		}
		bg[i] = 1
		stack = append(stack, i)
	}
	for x := 0; x < w; x++ {
		push(x)
		push((h-1)*w + x)
	}
	for y := 0; y < h; y++ {
		push(y * w)
		push(y*w + w - 1)
	}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		x, y := cur%w, cur/w
		if x > 0 {
			push(cur - 1)
		}
		if x+1 < w {
			push(cur + 1)
		}
		if y > 0 {
			push(cur - w)
		}
		if y+1 < h {
			push(cur + w)
		}
	}
	hole := 0
	for i := 0; i < n; i++ {
		if out[i] < 0.5 && bg[i] == 0 {
			hole++
		}
	}
	if best < 1 || hole > best*35/100 {
		return out
	}
	for i := 0; i < n; i++ {
		if out[i] < 0.5 && bg[i] == 0 {
			out[i] = 1
		}
	}
	return out
}

func resizeNRGBA(src *image.NRGBA, dw, dh int) *image.NRGBA {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	if sw < 1 || sh < 1 {
		return dst
	}
	for y := 0; y < dh; y++ {
		fy := (float64(y)+0.5)*float64(sh)/float64(dh) - 0.5
		for x := 0; x < dw; x++ {
			fx := (float64(x)+0.5)*float64(sw)/float64(dw) - 0.5
			r, g, b, a := sampleNRGBA4(src, fx, fy)
			i := dst.PixOffset(x, y)
			dst.Pix[i] = uint8(clampFloat(float64(r), 0, 255))
			dst.Pix[i+1] = uint8(clampFloat(float64(g), 0, 255))
			dst.Pix[i+2] = uint8(clampFloat(float64(b), 0, 255))
			dst.Pix[i+3] = uint8(clampFloat(float64(a), 0, 255))
		}
	}
	return dst
}

func sampleNRGBA4(img *image.NRGBA, x, y float64) (r, g, b, a float32) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > float64(w-1) {
		x = float64(w - 1)
	}
	if y > float64(h-1) {
		y = float64(h - 1)
	}
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1, y1 := x0+1, y0+1
	if x1 >= w {
		x1 = w - 1
	}
	if y1 >= h {
		y1 = h - 1
	}
	dx := float32(x - float64(x0))
	dy := float32(y - float64(y0))
	at := func(px, py int) (float32, float32, float32, float32) {
		i := img.PixOffset(px, py)
		return float32(img.Pix[i]), float32(img.Pix[i+1]), float32(img.Pix[i+2]), float32(img.Pix[i+3])
	}
	r00, g00, b00, a00 := at(x0, y0)
	r10, g10, b10, a10 := at(x1, y0)
	r01, g01, b01, a01 := at(x0, y1)
	r11, g11, b11, a11 := at(x1, y1)
	lerp4 := func(a, b float32) float32 { return a + (b-a)*dx }
	r = lerp4(r00, r10) + (lerp4(r01, r11)-lerp4(r00, r10))*dy
	g = lerp4(g00, g10) + (lerp4(g01, g11)-lerp4(g00, g10))*dy
	b = lerp4(b00, b10) + (lerp4(b01, b11)-lerp4(b00, b10))*dy
	a = lerp4(a00, a10) + (lerp4(a01, a11)-lerp4(a00, a10))*dy
	return r, g, b, a
}

func rotateBound(src *image.NRGBA, deg float64) *image.NRGBA {
	rad := deg * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)
	w, h := float64(src.Bounds().Dx()), float64(src.Bounds().Dy())
	nw := int(math.Ceil(math.Abs(w*cos) + math.Abs(h*sin)))
	nh := int(math.Ceil(math.Abs(w*sin) + math.Abs(h*cos)))
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, nw, nh))
	cx, cy := w/2, h/2
	ncx, ncy := float64(nw)/2, float64(nh)/2
	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			dx := float64(x) - ncx
			dy := float64(y) - ncy
			sx := cos*dx + sin*dy + cx
			sy := -sin*dx + cos*dy + cy
			if sx < -1 || sy < -1 || sx > w || sy > h {
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

func encodeJPEGLimited(img image.Image, targetKB int) (string, error) {
	quality := 92
	if targetKB <= 0 {
		return encodeJPEGDataURL(img, quality)
	}
	var best []byte
	for q := 92; q >= 40; q -= 4 {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
			return "", err
		}
		best = buf.Bytes()
		quality = q
		if len(best) <= targetKB*1024 {
			break
		}
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(best), nil
}
