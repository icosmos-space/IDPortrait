package core

import (
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
	reason := facePoseReject(hit)
	if !strings.Contains(reason, "倾斜") {
		t.Fatalf("expected head-tilt reject, got %q", reason)
	}
}
