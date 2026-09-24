package core

import (
	"strings"
	"testing"
)

func TestJudgeParseRejectsClosedEyes(t *testing.T) {
	labels := synthParseMap(func(x, y, side int) uint8 {
		// skin everywhere, brows but almost no eyes
		if y > side/5 && y < side/3 && (x < side*2/5 || x > side*3/5) {
			if x%20 < 8 {
				return parseLBrow
			}
			return parseRBrow
		}
		if y > side/3 && y < side/2 && x > side*2/5 && x < side*3/5 {
			return parseNose
		}
		if y > side*3/5 {
			return parseULip
		}
		return parseSkin
	})
	reason := judgeParseLabels(labels)
	if !strings.Contains(reason, "闭眼") {
		t.Fatalf("expected closed-eye reject, got %q", reason)
	}
}

func TestJudgeParseRejectsMaskOcclusion(t *testing.T) {
	labels := synthParseMap(func(x, y, side int) uint8 {
		if y < side/3 {
			if x < side/2 {
				return parseLEye
			}
			return parseREye
		}
		if y < side/2 {
			return parseSkin
		}
		// lower face covered by cloth (mask)
		return parseCloth
	})
	reason := judgeParseLabels(labels)
	if !strings.Contains(reason, "遮挡") {
		t.Fatalf("expected occlusion reject, got %q", reason)
	}
}

func TestJudgeParseRejectsSunglasses(t *testing.T) {
	labels := synthParseMap(func(x, y, side int) uint8 {
		if y > side/5 && y < side*2/5 {
			return parseGlasses
		}
		if y > side*2/5 && y < side/2 && x > side*2/5 && x < side*3/5 {
			return parseNose
		}
		if y > side*3/5 {
			return parseULip
		}
		return parseSkin
	})
	reason := judgeParseLabels(labels)
	if !strings.Contains(reason, "墨镜") && !strings.Contains(reason, "闭眼") {
		t.Fatalf("expected sunglasses/closed-eye reject, got %q", reason)
	}
}

func TestJudgeParseAllowsFrontal(t *testing.T) {
	labels := synthParseMap(func(x, y, side int) uint8 {
		if y > side/5 && y < side/3 {
			if x > side/4 && x < side*2/5 {
				return parseLEye
			}
			if x > side*3/5 && x < side*3/4 {
				return parseREye
			}
			if x < side/2 {
				return parseLBrow
			}
			return parseRBrow
		}
		if y > side/3 && y < side/2 && x > side*2/5 && x < side*3/5 {
			return parseNose
		}
		if y > side*11/20 && y < side*13/20 {
			return parseULip
		}
		if y > side*13/20 && y < side*3/4 {
			return parseLLip
		}
		return parseSkin
	})
	if reason := judgeParseLabels(labels); reason != "" {
		t.Fatalf("expected frontal ok, got %q", reason)
	}
}

func synthParseMap(fn func(x, y, side int) uint8) []uint8 {
	side := parseSize
	out := make([]uint8, side*side)
	for y := 0; y < side; y++ {
		for x := 0; x < side; x++ {
			out[y*side+x] = fn(x, y, side)
		}
	}
	return out
}
