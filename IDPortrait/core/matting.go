package core

import (
	"fmt"
	"image"
	"math"
	"path/filepath"
	"strings"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

type matteKind int

const (
	matteMODNet matteKind = iota
	matteRMBG
	matteBiRef
)

type matteSpec struct {
	file   string
	inName string
	out    string
	size   int
	kind   matteKind
}

type matteSession struct {
	mu   sync.Mutex
	sess *ort.AdvancedSession
	in   *ort.Tensor[float32]
	out  *ort.Tensor[float32]
	spec matteSpec
}

var matteCache sync.Map

func mattingModel(id string) matteSpec {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "modnet", "modnet_photographic_portrait_matting":
		return matteSpec{file: "modnet_photographic_portrait_matting.onnx", inName: "input", out: "output", size: 512, kind: matteMODNet}
	case "rmbg", "rmbg-1.4":
		return matteSpec{file: "rmbg-1.4.onnx", inName: "input", out: "output", size: 1024, kind: matteRMBG}
	case "birefnet", "birefnet-v1-lite":
		return matteSpec{file: "birefnet-v1-lite.onnx", inName: "input_image", out: "output_image", size: 1024, kind: matteBiRef}
	default:
		return matteSpec{file: "hivision_modnet.onnx", inName: "input1", out: "output1", size: 512, kind: matteMODNet}
	}
}

func mattingSession(spec matteSpec) (*matteSession, error) {
	if v, ok := matteCache.Load(spec.file); ok {
		return v.(*matteSession), nil
	}
	root, err := runtimeRoot()
	if err != nil {
		return nil, err
	}
	prepareRuntimeDir(root)
	if err := initONNX(filepath.Join(root, "onnxruntime.dll")); err != nil {
		return nil, err
	}
	in, err := ort.NewEmptyTensor[float32](ort.NewShape(1, 3, int64(spec.size), int64(spec.size)))
	if err != nil {
		return nil, err
	}
	out, err := ort.NewEmptyTensor[float32](ort.NewShape(1, 1, int64(spec.size), int64(spec.size)))
	if err != nil {
		in.Destroy()
		return nil, err
	}
	sess, err := ort.NewAdvancedSession(
		filepath.Join(root, "models", spec.file),
		[]string{spec.inName},
		[]string{spec.out},
		[]ort.Value{in},
		[]ort.Value{out},
		nil,
	)
	if err != nil {
		in.Destroy()
		out.Destroy()
		return nil, fmt.Errorf("加载抠图模型 %s 失败: %w", spec.file, err)
	}
	ms := &matteSession{sess: sess, in: in, out: out, spec: spec}
	actual, loaded := matteCache.LoadOrStore(spec.file, ms)
	if loaded {
		sess.Destroy()
		in.Destroy()
		out.Destroy()
	}
	return actual.(*matteSession), nil
}

// portraitMatting returns an RGBA cutout. Alpha is the person.
func portraitMatting(src *image.NRGBA, modelID string) (*image.NRGBA, error) {
	spec := mattingModel(modelID)
	sess, err := mattingSession(spec)
	if err != nil {
		return nil, err
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()

	fillMatteInput(sess.in.GetData(), src, spec)
	if err := sess.sess.Run(); err != nil {
		return nil, fmt.Errorf("抠图推理失败: %w", err)
	}
	mask := matteToMask(sess.out.GetData(), spec.size, spec.kind)
	alpha := resizeGray(mask, spec.size, spec.size, src.Bounds().Dx(), src.Bounds().Dy())
	if spec.file == "hivision_modnet.onnx" {
		alpha = fillMatteHoles(alpha, src.Bounds().Dx(), src.Bounds().Dy())
	}
	return applyAlpha(src, alpha), nil
}

func fillMatteInput(dst []float32, src *image.NRGBA, spec matteSpec) {
	n := spec.size
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	plane := n * n
	pix := src.Pix
	stride := src.Stride
	for y := 0; y < n; y++ {
		sy := (float64(y)+0.5)*float64(sh)/float64(n) - 0.5
		for x := 0; x < n; x++ {
			sx := (float64(x)+0.5)*float64(sw)/float64(n) - 0.5
			r, g, b := sampleNRGBA(pix, stride, sw, sh, sx, sy)
			rf, gf, bf := r/255, g/255, b/255
			idx := y*n + x
			switch spec.kind {
			case matteBiRef:
				dst[idx] = (rf - 0.485) / 0.229
				dst[plane+idx] = (gf - 0.456) / 0.224
				dst[2*plane+idx] = (bf - 0.406) / 0.225
			default:
				dst[idx] = (rf - 0.5) / 0.5
				dst[plane+idx] = (gf - 0.5) / 0.5
				dst[2*plane+idx] = (bf - 0.5) / 0.5
			}
		}
	}
}

func matteToMask(raw []float32, size int, kind matteKind) []float32 {
	n := size * size
	out := make([]float32, n)
	switch kind {
	case matteBiRef:
		for i := 0; i < n; i++ {
			out[i] = sigmoid(raw[i])
		}
	case matteRMBG:
		minV, maxV := raw[0], raw[0]
		for i := 1; i < n; i++ {
			if raw[i] < minV {
				minV = raw[i]
			}
			if raw[i] > maxV {
				maxV = raw[i]
			}
		}
		span := maxV - minV
		if span < 1e-6 {
			span = 1
		}
		for i := 0; i < n; i++ {
			out[i] = (raw[i] - minV) / span
		}
	default:
		scale := float32(1)
		maxV := float32(0)
		for i := 0; i < n; i++ {
			if raw[i] > maxV {
				maxV = raw[i]
			}
		}
		if maxV > 1.5 {
			scale = 1 / 255
		}
		for i := 0; i < n; i++ {
			v := raw[i] * scale
			if v < 0 {
				v = 0
			}
			if v > 1 {
				v = 1
			}
			out[i] = v
		}
	}
	return out
}

func sigmoid(v float32) float32 {
	return float32(1 / (1 + math.Exp(float64(-v))))
}
