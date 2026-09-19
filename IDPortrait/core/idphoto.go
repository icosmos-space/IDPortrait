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

func layoutSheet(tile *image.RGBA, paperW, paperH int) *image.RGBA {
	if paperW < tile.Bounds().Dx() {
		paperW = tile.Bounds().Dx() + 40
	}
	if paperH < tile.Bounds().Dy() {
		paperH = tile.Bounds().Dy() + 40
	}
	dst := image.NewRGBA(image.Rect(0, 0, paperW, paperH))
	fillBackground(dst, color.RGBA{R: 255, G: 255, B: 255, A: 255}, "solid")
	place := func(photo *image.RGBA, pw, ph int) int {
		gap := mmToPx(2)
		margin := mmToPx(4)
		cols := max(1, (pw-margin*2+gap)/(photo.Bounds().Dx()+gap))
		rows := max(1, (ph-margin*2+gap)/(photo.Bounds().Dy()+gap))
		used := 0
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				x := margin + c*(photo.Bounds().Dx()+gap)
				y := margin + r*(photo.Bounds().Dy()+gap)
				if x+photo.Bounds().Dx() > pw || y+photo.Bounds().Dy() > ph {
					continue
				}
				drawOver(dst, photo, x, y)
				used++
			}
		}
		return used
	}
	upright := place(tile, paperW, paperH)
	turned := rotateRGBA90(tile)
	sideways := countFit(turned, paperW, paperH)
	if sideways > upright {
		fillBackground(dst, color.RGBA{R: 255, G: 255, B: 255, A: 255}, "solid")
		place(turned, paperW, paperH)
	}
	return dst
}

func countFit(photo *image.RGBA, pw, ph int) int {
	gap := mmToPx(2)
	margin := mmToPx(4)
	cols := max(1, (pw-margin*2+gap)/(photo.Bounds().Dx()+gap))
	rows := max(1, (ph-margin*2+gap)/(photo.Bounds().Dy()+gap))
	n := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x := margin + c*(photo.Bounds().Dx()+gap)
			y := margin + r*(photo.Bounds().Dy()+gap)
			if x+photo.Bounds().Dx() <= pw && y+photo.Bounds().Dy() <= ph {
				n++
			}
		}
	}
	return n
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

func rotateRGBA90(src *image.RGBA) *image.RGBA {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si := src.PixOffset(x, y)
			di := dst.PixOffset(h-1-y, x)
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
	// Fill enclosed holes: background is reachable from the border through zeros.
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
