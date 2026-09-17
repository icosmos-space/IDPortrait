//go:build windows

package printer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"syscall"
	"unsafe"

	"IDPortrait/core"

	"golang.org/x/sys/windows"
)

var (
	modWinspool = windows.NewLazySystemDLL("winspool.drv")
	modGdi32    = windows.NewLazySystemDLL("gdi32.dll")

	procEnumPrintersW       = modWinspool.NewProc("EnumPrintersW")
	procGetDefaultPrinterW  = modWinspool.NewProc("GetDefaultPrinterW")
	procOpenPrinterW        = modWinspool.NewProc("OpenPrinterW")
	procClosePrinter        = modWinspool.NewProc("ClosePrinter")
	procDocumentPropertiesW = modWinspool.NewProc("DocumentPropertiesW")
	procCreateDCW           = modGdi32.NewProc("CreateDCW")
	procDeleteDC            = modGdi32.NewProc("DeleteDC")
	procStartDocW           = modGdi32.NewProc("StartDocW")
	procEndDoc              = modGdi32.NewProc("EndDoc")
	procStartPage           = modGdi32.NewProc("StartPage")
	procEndPage             = modGdi32.NewProc("EndPage")
	procGetDeviceCaps       = modGdi32.NewProc("GetDeviceCaps")
	procStretchDIBits       = modGdi32.NewProc("StretchDIBits")
)

const (
	PRINTER_ENUM_LOCAL   = 0x00000002
	PRINTER_ENUM_CONNECTIONS = 0x00000004

	DM_OUT_BUFFER = 2
	DM_IN_BUFFER  = 8

	DM_ORIENTATION = 0x00000001
	DM_PAPERSIZE   = 0x00000002
	DM_PAPERLENGTH = 0x00000004
	DM_PAPERWIDTH  = 0x00000008
	DM_COPIES      = 0x00000100

	DMORIENT_PORTRAIT  = 1
	DMORIENT_LANDSCAPE = 2

	DMPAPER_A4   = 9
	DMPAPER_A5   = 11
	DMPAPER_A6   = 70
	DMPAPER_USER = 256

	HORZRES = 8
	VERTRES = 10

	BI_RGB         = 0
	DIB_RGB_COLORS = 0
	SRCCOPY        = 0x00CC0020
)

type printerInfo4 struct {
	pPrinterName *uint16
	pServerName  *uint16
	Attributes   uint32
}

type docInfoW struct {
	cbSize       int32
	lpszDocName  *uint16
	lpszOutput   *uint16
	lpszDatatype *uint16
	fwType       uint32
}

// DEVMODEW truncated to the fields we need; driver extra follows dmSize.
type devModeW struct {
	dmDeviceName     [32]uint16
	dmSpecVersion    uint16
	dmDriverVersion  uint16
	dmSize           uint16
	dmDriverExtra    uint16
	dmFields         uint32
	dmOrientation    int16
	dmPaperSize      int16
	dmPaperLength    int16
	dmPaperWidth     int16
	dmScale          int16
	dmCopies         int16
	dmDefaultSource  int16
	dmPrintQuality   int16
	dmColor          int16
	dmDuplex         int16
	dmYResolution    int16
	dmTTOption       int16
	dmCollate        int16
	dmFormName       [32]uint16
	dmLogPixels      uint16
	dmBitsPerPel     uint32
	dmPelsWidth      uint32
	dmPelsHeight     uint32
	dmDisplayFlags   uint32
	dmDisplayFreq    uint32
	dmICMMethod      uint32
	dmICMIntent      uint32
	dmMediaType      uint32
	dmDitherType     uint32
	dmReserved1      uint32
	dmReserved2      uint32
	dmPanningWidth   uint32
	dmPanningHeight  uint32
}

type bitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

func ListPrinters() ([]core.PrinterInfo, error) {
	flags := uint32(PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS)
	var needed, returned uint32
	r1, _, err := procEnumPrintersW.Call(
		uintptr(flags),
		0,
		4,
		0,
		0,
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)),
	)
	if r1 == 0 && needed == 0 {
		if err != windows.ERROR_INSUFFICIENT_BUFFER && err != syscall.Errno(122) {
			return nil, fmt.Errorf("枚举打印机失败: %v", err)
		}
	}
	if needed == 0 {
		return []core.PrinterInfo{}, nil
	}
	buf := make([]byte, needed)
	r1, _, err = procEnumPrintersW.Call(
		uintptr(flags),
		0,
		4,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(needed),
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)),
	)
	if r1 == 0 {
		return nil, fmt.Errorf("枚举打印机失败: %v", err)
	}

	defName, _ := defaultPrinterName()
	out := make([]core.PrinterInfo, 0, returned)
	infos := unsafe.Slice((*printerInfo4)(unsafe.Pointer(&buf[0])), int(returned))
	for _, info := range infos {
		name := windows.UTF16PtrToString(info.pPrinterName)
		if strings.TrimSpace(name) == "" {
			continue
		}
		out = append(out, core.PrinterInfo{
			Name:      name,
			IsDefault: name == defName,
		})
	}
	return out, nil
}

func defaultPrinterName() (string, error) {
	var needed uint32 = 0
	r1, _, err := procGetDefaultPrinterW.Call(0, uintptr(unsafe.Pointer(&needed)))
	if r1 == 0 && needed == 0 {
		return "", err
	}
	buf := make([]uint16, needed)
	r1, _, err = procGetDefaultPrinterW.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&needed)))
	if r1 == 0 {
		return "", err
	}
	return windows.UTF16ToString(buf), nil
}

func PrintLayout(opt core.PrintOptions) (*core.PrintResult, error) {
	img, err := decodeDataURLImage(opt.ImageDataURL)
	if err != nil {
		return nil, err
	}
	printerName := strings.TrimSpace(opt.PrinterName)
	if printerName == "" {
		printerName, err = defaultPrinterName()
		if err != nil || printerName == "" {
			return nil, fmt.Errorf("未找到可用打印机")
		}
	}
	copies := opt.Copies
	if copies < 1 {
		copies = 1
	}
	if copies > 99 {
		copies = 99
	}

	dib, hdr, err := imageToBGRDIB(img)
	if err != nil {
		return nil, err
	}

	namePtr, err := windows.UTF16PtrFromString(printerName)
	if err != nil {
		return nil, err
	}

	var hPrinter uintptr
	r1, _, err := procOpenPrinterW.Call(uintptr(unsafe.Pointer(namePtr)), uintptr(unsafe.Pointer(&hPrinter)), 0)
	if r1 == 0 {
		return nil, fmt.Errorf("打开打印机失败: %v", err)
	}
	defer procClosePrinter.Call(hPrinter)

	devMode, err := buildDevMode(hPrinter, namePtr, opt.PaperSize, copies, opt.Landscape)
	if err != nil {
		return nil, err
	}

	driver, err := windows.UTF16PtrFromString("WINSPOOL")
	if err != nil {
		return nil, err
	}
	hdc, _, err := procCreateDCW.Call(
		uintptr(unsafe.Pointer(driver)),
		uintptr(unsafe.Pointer(namePtr)),
		0,
		uintptr(unsafe.Pointer(&devMode[0])),
	)
	if hdc == 0 {
		return nil, fmt.Errorf("创建打印设备失败: %v", err)
	}
	defer procDeleteDC.Call(hdc)

	docName, _ := windows.UTF16PtrFromString("最美证件照-排版照")
	di := docInfoW{
		cbSize:      int32(unsafe.Sizeof(docInfoW{})),
		lpszDocName: docName,
	}
	job, _, err := procStartDocW.Call(hdc, uintptr(unsafe.Pointer(&di)))
	if int32(job) <= 0 {
		return nil, fmt.Errorf("开始打印任务失败: %v", err)
	}
	defer procEndDoc.Call(hdc)

	page, _, err := procStartPage.Call(hdc)
	if int32(page) <= 0 {
		return nil, fmt.Errorf("开始打印页失败: %v", err)
	}

	pageW, _, _ := procGetDeviceCaps.Call(hdc, HORZRES)
	pageH, _, _ := procGetDeviceCaps.Call(hdc, VERTRES)
	dstW, dstH, ox, oy := fitRect(int(hdr.biWidth), int(hdr.biHeight), int(pageW), int(pageH))

	r1, _, err = procStretchDIBits.Call(
		hdc,
		uintptr(ox),
		uintptr(oy),
		uintptr(dstW),
		uintptr(dstH),
		0,
		0,
		uintptr(hdr.biWidth),
		uintptr(hdr.biHeight),
		uintptr(unsafe.Pointer(&dib[0])),
		uintptr(unsafe.Pointer(&hdr)),
		DIB_RGB_COLORS,
		SRCCOPY,
	)
	if int32(r1) <= 0 {
		_ = err
		procEndPage.Call(hdc)
		return nil, fmt.Errorf("绘制排版图失败")
	}
	procEndPage.Call(hdc)

	return &core.PrintResult{
		OK:      true,
		Printer: printerName,
		Copies:  copies,
	}, nil
}

func buildDevMode(hPrinter uintptr, namePtr *uint16, paperSize string, copies int, landscape bool) ([]byte, error) {
	hwnd := uintptr(0)
	size, _, _ := procDocumentPropertiesW.Call(hwnd, hPrinter, uintptr(unsafe.Pointer(namePtr)), 0, 0, 0)
	if int32(size) <= 0 {
		return nil, fmt.Errorf("读取打印机配置失败")
	}
	buf := make([]byte, size)
	r1, _, err := procDocumentPropertiesW.Call(
		hwnd,
		hPrinter,
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(&buf[0])),
		0,
		DM_OUT_BUFFER,
	)
	if int32(r1) < 0 {
		return nil, fmt.Errorf("读取打印机默认设置失败: %v", err)
	}

	dm := (*devModeW)(unsafe.Pointer(&buf[0]))
	if dm.dmSize == 0 {
		dm.dmSize = uint16(unsafe.Sizeof(devModeW{}))
	}
	dm.dmCopies = int16(copies)
	dm.dmFields |= DM_COPIES
	if landscape {
		dm.dmOrientation = DMORIENT_LANDSCAPE
	} else {
		dm.dmOrientation = DMORIENT_PORTRAIT
	}
	dm.dmFields |= DM_ORIENTATION
	applyPaper(dm, paperSize)

	r1, _, err = procDocumentPropertiesW.Call(
		hwnd,
		hPrinter,
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&buf[0])),
		DM_IN_BUFFER|DM_OUT_BUFFER,
	)
	if int32(r1) < 0 {
		// Keep local modifications even if driver rejects merge.
		_ = err
	}
	return buf, nil
}

func applyPaper(dm *devModeW, paperSize string) {
	switch strings.ToLower(strings.TrimSpace(paperSize)) {
	case "a4":
		dm.dmPaperSize = DMPAPER_A4
		dm.dmFields |= DM_PAPERSIZE
		dm.dmFields &^= DM_PAPERWIDTH | DM_PAPERLENGTH
	case "a5":
		dm.dmPaperSize = DMPAPER_A5
		dm.dmFields |= DM_PAPERSIZE
		dm.dmFields &^= DM_PAPERWIDTH | DM_PAPERLENGTH
	case "a6":
		dm.dmPaperSize = DMPAPER_A6
		dm.dmFields |= DM_PAPERSIZE
		dm.dmFields &^= DM_PAPERWIDTH | DM_PAPERLENGTH
	case "5inch":
		setCustomPaper(dm, 89, 127)
	case "6inch":
		setCustomPaper(dm, 102, 152)
	case "7inch":
		setCustomPaper(dm, 127, 178)
	default:
		// Prefer matching catalog dimensions when known.
		for _, p := range core.BuiltinPaperSpecs() {
			if p.Value == paperSize && p.WidthMM > 0 && p.HeightMM > 0 {
				setCustomPaper(dm, p.WidthMM, p.HeightMM)
				return
			}
		}
		dm.dmPaperSize = DMPAPER_A4
		dm.dmFields |= DM_PAPERSIZE
	}
}

func setCustomPaper(dm *devModeW, widthMM, heightMM float64) {
	dm.dmPaperSize = DMPAPER_USER
	dm.dmPaperWidth = int16(widthMM * 10)   // 0.1 mm
	dm.dmPaperLength = int16(heightMM * 10) // 0.1 mm
	dm.dmFields |= DM_PAPERSIZE | DM_PAPERWIDTH | DM_PAPERLENGTH
}

func fitRect(srcW, srcH, pageW, pageH int) (dstW, dstH, ox, oy int) {
	if srcW <= 0 || srcH <= 0 || pageW <= 0 || pageH <= 0 {
		return pageW, pageH, 0, 0
	}
	scale := float64(pageW) / float64(srcW)
	if float64(srcH)*scale > float64(pageH) {
		scale = float64(pageH) / float64(srcH)
	}
	dstW = max(1, int(float64(srcW)*scale))
	dstH = max(1, int(float64(srcH)*scale))
	ox = (pageW - dstW) / 2
	oy = (pageH - dstH) / 2
	return
}

func decodeDataURLImage(dataURL string) (image.Image, error) {
	dataURL = strings.TrimSpace(dataURL)
	if dataURL == "" {
		return nil, fmt.Errorf("排版图为空")
	}
	parts := strings.SplitN(dataURL, ",", 2)
	raw := dataURL
	if len(parts) == 2 {
		raw = parts[1]
	}
	bin, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("解码排版图失败: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(bin))
	if err != nil {
		return nil, fmt.Errorf("解析排版图失败: %w", err)
	}
	return img, nil
}

func imageToBGRDIB(src image.Image) ([]byte, bitmapInfoHeader, error) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, bitmapInfoHeader{}, fmt.Errorf("无效图片尺寸")
	}
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(rgba, rgba.Bounds(), src, b.Min, draw.Src)

	rowSize := (w*3 + 3) &^ 3
	dib := make([]byte, rowSize*h)
	for y := 0; y < h; y++ {
		srcY := h - 1 - y // bottom-up DIB
		row := dib[y*rowSize : (y+1)*rowSize]
		for x := 0; x < w; x++ {
			i := (srcY*w + x) * 4
			off := x * 3
			row[off+0] = rgba.Pix[i+2] // B
			row[off+1] = rgba.Pix[i+1] // G
			row[off+2] = rgba.Pix[i+0] // R
		}
	}
	hdr := bitmapInfoHeader{
		biSize:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		biWidth:       int32(w),
		biHeight:      int32(h),
		biPlanes:      1,
		biBitCount:    24,
		biCompression: BI_RGB,
		biSizeImage:   uint32(len(dib)),
	}
	return dib, hdr, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
