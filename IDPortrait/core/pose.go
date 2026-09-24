package core

import "math"

// facePoseReject returns a Chinese reject reason when the five YuNet keypoints
// show a non-frontal pose (side face / looking down / looking up).
// Empty string means the pose is acceptable for an ID photo.
//
// YuNet landmark order: right eye, left eye, nose tip, right mouth, left mouth.
func facePoseReject(f faceHit) string {
	rex, rey := f.kps[0], f.kps[1]
	lex, ley := f.kps[2], f.kps[3]
	nx, ny := f.kps[4], f.kps[5]
	_, rmy := f.kps[6], f.kps[7]
	_, lmy := f.kps[8], f.kps[9]

	eyeDist := math.Hypot(rex-lex, rey-ley)
	if eyeDist < 1 {
		return "人脸关键点无效，已拒绝"
	}

	eyeMidX := (rex + lex) / 2
	eyeMidY := (rey + ley) / 2
	// Yaw: nose should sit near the midpoint between the eyes.
	if math.Abs(nx-eyeMidX)/eyeDist > 0.32 {
		return "检测到侧脸，已拒绝"
	}

	mouthMidY := (rmy + lmy) / 2
	vert := mouthMidY - eyeMidY
	if vert < eyeDist*0.35 {
		// Eyes and mouth collapsed vertically — usually a strong head-down pose.
		return "检测到低头，已拒绝"
	}

	// Pitch: nose position between eyes and mouth.
	// Frontal ≈ 0.45–0.70; looking down pushes toward the mouth; looking up toward the eyes.
	t := (ny - eyeMidY) / vert
	if t > 0.78 {
		return "检测到低头，已拒绝"
	}
	if t < 0.28 {
		return "检测到抬头，已拒绝"
	}
	return ""
}
