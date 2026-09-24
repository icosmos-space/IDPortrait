package core

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

func TestHandPhotoRejectedByDetectFace(t *testing.T) {
	path := `C:\Users\tacey\.cursor\projects\d-gospace-src-github-com-icosmos-space-IDPortrait\assets\c__Users_tacey_AppData_Roaming_Cursor_User_workspaceStorage_ccffda6d87fcd2bd7d20207414d6276d_images_image-9c7cd379-aaf9-4b02-80f9-ca63e851c692.png`
	f, err := os.Open(path)
	if err != nil {
		t.Skip(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	url, err := encodeJPEGDataURL(toNRGBA(downscale(img, 1600)), 90)
	if err != nil {
		t.Fatal(err)
	}
	// Empty opts (= all Skip false) must still reject — mirrors broken Wails binding.
	res, err := DetectFace(url, FaceCheckOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Fatalf("expected reject, got OK score=%v", res.Score)
	}
	if res.Reason == "" || res.Reason == "未检测到人脸，已拒绝" {
		t.Fatalf("unexpected reason %q", res.Reason)
	}
}
