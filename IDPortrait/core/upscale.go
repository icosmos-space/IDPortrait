package core

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	upscaylModelName = "upscayl-standard-4x"
	upscaylScale     = 4
)

// upscalePortrait runs Upscayl (ncnn) 4× on a cutout, preserving alpha by
// upscaling RGB via the model and resizing the alpha channel to match.
func upscalePortrait(src *image.NRGBA) (*image.NRGBA, error) {
	if src == nil {
		return nil, fmt.Errorf("empty image")
	}
	bin, models, err := resolveUpscaylPaths()
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "idportrait-upscayl-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.png")
	outPath := filepath.Join(tmpDir, "out.png")
	if err := writePNGNRGBA(inPath, flattenRGBKeepSize(src)); err != nil {
		return nil, err
	}

	cmd := exec.Command(bin,
		"-i", inPath,
		"-o", outPath,
		"-m", models,
		"-n", upscaylModelName,
		"-s", fmt.Sprintf("%d", upscaylScale),
		"-z", fmt.Sprintf("%d", upscaylScale),
		"-f", "png",
		"-t", "0",
		"-c", "0",
	)
	cmd.Dir = filepath.Dir(bin)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		if len(msg) > 400 {
			msg = msg[len(msg)-400:]
		}
		return nil, fmt.Errorf("upscayl: %s", msg)
	}
	rgb, err := readPNGNRGBA(outPath)
	if err != nil {
		return nil, fmt.Errorf("读取扩图结果失败: %w", err)
	}
	return mergeScaledAlpha(rgb, src), nil
}

func resolveUpscaylPaths() (binPath, modelsPath string, err error) {
	root, err := runtimeRoot()
	if err != nil {
		return "", "", err
	}
	base := filepath.Join(root, "upscayl")
	binPath = filepath.Join(base, "bin", "upscayl-bin.exe")
	modelsPath = filepath.Join(base, "models")
	if !fileExists(binPath) {
		return "", "", fmt.Errorf("找不到 upscayl-bin.exe（请放到 %s）", filepath.Join(base, "bin"))
	}
	param := filepath.Join(modelsPath, upscaylModelName+".param")
	binModel := filepath.Join(modelsPath, upscaylModelName+".bin")
	if !fileExists(param) || !fileExists(binModel) {
		return "", "", fmt.Errorf("找不到模型 %s（请放到 %s）", upscaylModelName, modelsPath)
	}
	return binPath, modelsPath, nil
}

// flattenRGBKeepSize copies RGB; transparent pixels keep underlying RGB (matting).
func flattenRGBKeepSize(src *image.NRGBA) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	copy(dst.Pix, src.Pix)
	for i := 3; i < len(dst.Pix); i += 4 {
		dst.Pix[i] = 255
	}
	return dst
}

func mergeScaledAlpha(rgb, srcAlpha *image.NRGBA) *image.NRGBA {
	dw, dh := rgb.Bounds().Dx(), rgb.Bounds().Dy()
	alpha := resizeNRGBA(srcAlpha, dw, dh)
	out := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			si := rgb.PixOffset(x, y)
			ai := alpha.PixOffset(x, y)
			di := out.PixOffset(x, y)
			out.Pix[di] = rgb.Pix[si]
			out.Pix[di+1] = rgb.Pix[si+1]
			out.Pix[di+2] = rgb.Pix[si+2]
			out.Pix[di+3] = alpha.Pix[ai+3]
		}
	}
	return out
}

func writePNGNRGBA(path string, img *image.NRGBA) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func readPNGNRGBA(path string) (*image.NRGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	return toNRGBA(img), nil
}
