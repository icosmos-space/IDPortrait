package core

import (
	"image"
	"image/color"
	"testing"
)

func TestAdjustIDPhotoSize(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 800, 1000))
	for i := 3; i < len(src.Pix); i += 4 {
		src.Pix[i] = 255
	}
	face := faceHit{x: 280, y: 180, w: 240, h: 280, score: 0.9}
	out := adjustIDPhoto(src, face, 295, 413, 0.2, 0.12)
	if out.Bounds().Dx() != 295 || out.Bounds().Dy() != 413 {
		t.Fatalf("size %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}

func TestJudgeLayoutOneInchOnSixInch(t *testing.T) {
	// Hivision defaults: 295×413 on 1795×1205 → 5×2 upright, 10 photos.
	limitW := hivisionLayoutW - 2*hivisionSideW // 1655
	limitH := hivisionLayoutH - 2*hivisionSideH // 1105
	mode := judgeLayout(295, 413, hivisionGap, hivisionGap, limitW, limitH)
	if mode.rotate {
		t.Fatalf("expected upright layout, got rotate")
	}
	if mode.cols != 5 || mode.rows != 2 {
		t.Fatalf("cols×rows = %d×%d, want 5×2", mode.cols, mode.rows)
	}
	if mode.blockW != 1595 || mode.blockH != 856 {
		t.Fatalf("block %dx%d, want 1595x856", mode.blockW, mode.blockH)
	}
}

func TestLayoutSheetOneInchCentered(t *testing.T) {
	const photoW, photoH = 295, 413
	tile := image.NewRGBA(image.Rect(0, 0, photoW, photoH))
	red := color.RGBA{R: 200, G: 40, B: 40, A: 255}
	for y := 0; y < photoH; y++ {
		for x := 0; x < photoW; x++ {
			tile.SetRGBA(x, y, red)
		}
	}

	sheet := layoutSheet(tile, hivisionLayoutW, hivisionLayoutH, true)
	if sheet.Bounds().Dx() != hivisionLayoutW || sheet.Bounds().Dy() != hivisionLayoutH {
		t.Fatalf("sheet size %dx%d", sheet.Bounds().Dx(), sheet.Bounds().Dy())
	}

	mode := judgeLayout(photoW, photoH, hivisionGap, hivisionGap,
		hivisionLayoutW-2*hivisionSideW, hivisionLayoutH-2*hivisionSideH)
	wantN := mode.cols * mode.rows
	if wantN != 10 {
		t.Fatalf("expected 10 tiles, judge gave %d", wantN)
	}

	minX, minY := hivisionLayoutW, hivisionLayoutH
	maxX, maxY := -1, -1
	nInk := 0
	for y := 0; y < hivisionLayoutH; y++ {
		for x := 0; x < hivisionLayoutW; x++ {
			c := sheet.RGBAAt(x, y)
			if c.R > 180 && c.G < 80 && c.B < 80 {
				nInk++
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
	}
	// Crop lines overwrite photo edges; still expect nearly full tile coverage.
	perTile := photoW * photoH
	if nInk < wantN*(perTile-photoW-photoH) {
		t.Fatalf("ink pixels %d too low for %d photos", nInk, wantN)
	}
	blockCX := (minX + maxX + 1) / 2
	blockCY := (minY + maxY + 1) / 2
	paperCX := hivisionLayoutW / 2
	paperCY := hivisionLayoutH / 2
	tol := hivisionGap / 2
	if abs(blockCX-paperCX) > tol || abs(blockCY-paperCY) > tol {
		t.Fatalf("block center (%d,%d) vs paper (%d,%d), tol %d",
			blockCX, blockCY, paperCX, paperCY, tol)
	}
	if minX < 0 || minY < 0 || maxX >= hivisionLayoutW || maxY >= hivisionLayoutH {
		t.Fatalf("photos out of bounds: (%d,%d)-(%d,%d)", minX, minY, maxX, maxY)
	}

	// Crop guides run full height/width at first photo edge.
	x0 := (hivisionLayoutW - mode.blockW) / 2
	guide := sheet.RGBAAt(x0, 0)
	if guide.R != 200 || guide.G != 200 || guide.B != 200 {
		t.Fatalf("missing crop line at x=%d: got %#v", x0, guide)
	}
}

func TestTransposeFlipVertical(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 3, 2))
	// Mark unique pixels: (0,0)=R, (2,0)=G, (0,1)=B, (2,1)=A-marker yellow
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(2, 0, color.RGBA{G: 255, A: 255})
	src.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	src.SetRGBA(2, 1, color.RGBA{R: 255, G: 255, A: 255})

	dst := transposeFlipVertical(src)
	if dst.Bounds().Dx() != 2 || dst.Bounds().Dy() != 3 {
		t.Fatalf("size %dx%d, want 2x3", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
	// out(dx,dy) = src(w-1-dy, dx) with w=3 (cv2.transpose + flip vertical).
	assertRGBA(t, dst, 0, 0, color.RGBA{G: 255, A: 255})
	assertRGBA(t, dst, 1, 0, color.RGBA{R: 255, G: 255, A: 255})
	assertRGBA(t, dst, 0, 2, color.RGBA{R: 255, A: 255})
	assertRGBA(t, dst, 1, 2, color.RGBA{B: 255, A: 255})
}

func assertRGBA(t *testing.T, img *image.RGBA, x, y int, want color.RGBA) {
	t.Helper()
	got := img.RGBAAt(x, y)
	if got != want {
		t.Fatalf("pixel (%d,%d)=%v, want %v", x, y, got, want)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
