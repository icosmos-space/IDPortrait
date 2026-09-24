package core

import (
	"image"
	"strings"
	"testing"
)

func TestFacePoseFrontalOK(t *testing.T) {
	// YuNet order: right eye, left eye, nose, right mouth, left mouth
	hit := faceHit{kps: [10]float64{
		120, 100,
		180, 100,
		150, 130,
		125, 160,
		175, 160,
	}}
	if reason := facePoseReject(hit); reason != "" {
		t.Fatalf("expected frontal ok, got %q", reason)
	}
}

func TestFacePoseRejectsYaw(t *testing.T) {
	hit := faceHit{kps: [10]float64{
		120, 100,
		180, 100,
		175, 130,
		125, 160,
		175, 160,
	}}
	reason := facePoseReject(hit)
	if !strings.Contains(reason, "侧脸") {
		t.Fatalf("expected side-face reject, got %q", reason)
	}
}

func TestFacePoseRejectsPitchDown(t *testing.T) {
	hit := faceHit{kps: [10]float64{
		120, 100,
		180, 100,
		150, 155,
		125, 160,
		175, 160,
	}}
	reason := facePoseReject(hit)
	if !strings.Contains(reason, "低头") {
		t.Fatalf("expected looking-down reject, got %q", reason)
	}
}

func TestFacePoseRejectsPitchUp(t *testing.T) {
	hit := faceHit{kps: [10]float64{
		120, 100,
		180, 100,
		150, 108,
		125, 160,
		175, 160,
	}}
	reason := facePoseReject(hit)
	if !strings.Contains(reason, "抬头") {
		t.Fatalf("expected looking-up reject, got %q", reason)
	}
}

func TestFacePoseRejectsRoll(t *testing.T) {
	// Eyes differ by ~15° — head tilt.
	hit := faceHit{kps: [10]float64{
		120, 100,
		180, 116,
		150, 130,
		125, 160,
		175, 165,
	}}
	if reason := facePoseReject(hit); reason != "" {
		t.Fatalf("yaw/pitch gate must ignore roll, got %q", reason)
	}
	reason := faceRollReject(hit)
	if !strings.Contains(reason, "倾斜") {
		t.Fatalf("expected head-tilt reject, got %q", reason)
	}
}

func TestShoulderYawRejectsTwistedTorso(t *testing.T) {
	// White studio bg; face box centered; dark torso shifted to the right
	// (body turned while head would still look frontal).
	img := image.NewNRGBA(image.Rect(0, 0, 400, 560))
	for y := 0; y < 560; y++ {
		for x := 0; x < 400; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i] = 245
			img.Pix[i+1] = 245
			img.Pix[i+2] = 245
			img.Pix[i+3] = 255
		}
	}
	// Face oval (approx) centered.
	for y := 80; y < 220; y++ {
		for x := 140; x < 260; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i] = 90
			img.Pix[i+1] = 70
			img.Pix[i+2] = 60
		}
	}
	// Torso heavily biased to the right of the face.
	for y := 210; y < 420; y++ {
		for x := 170; x < 340; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i] = 40
			img.Pix[i+1] = 40
			img.Pix[i+2] = 50
		}
	}
	box := []float64{140, 80, 260, 220}
	reason := shoulderTiltReject(img, box)
	if !strings.Contains(reason, "扭转") && !strings.Contains(reason, "倾斜") {
		t.Fatalf("expected shoulder twist reject, got %q", reason)
	}
}

func TestShoulderYawAllowsFrontalTorso(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 400, 560))
	for y := 0; y < 560; y++ {
		for x := 0; x < 400; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i] = 245
			img.Pix[i+1] = 245
			img.Pix[i+2] = 245
			img.Pix[i+3] = 255
		}
	}
	for y := 80; y < 220; y++ {
		for x := 140; x < 260; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i] = 90
			img.Pix[i+1] = 70
			img.Pix[i+2] = 60
		}
	}
	// Symmetric shoulders under the face.
	for y := 210; y < 420; y++ {
		for x := 110; x < 290; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i] = 40
			img.Pix[i+1] = 40
			img.Pix[i+2] = 50
		}
	}
	box := []float64{140, 80, 260, 220}
	if reason := shoulderTiltReject(img, box); reason != "" {
		t.Fatalf("expected frontal torso ok, got %q", reason)
	}
}
