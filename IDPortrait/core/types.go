package core

// Report is the face / authenticity diagnosis result.
type Report struct {
	FaceOK    bool    `json:"faceOk"`
	FaceScore float64 `json:"faceScore"`
	IsRephoto bool    `json:"isRephoto"`
	IsAiImage bool    `json:"isAiImage"`
}

// GenerateParams mirrors frontend generation options.
type GenerateParams struct {
	Template             string  `json:"template"`
	BgMode               string  `json:"bgMode"`
	BgColor              string  `json:"bgColor"`
	BgPreset             string  `json:"bgPreset"`
	CustomWidth          float64 `json:"customWidth"`
	CustomHeight         float64 `json:"customHeight"`
	BeautyStrength       float64 `json:"beautyStrength"`
	EyeSharp             float64 `json:"eyeSharp"`
	SkinBright           float64 `json:"skinBright"`
	EnableMakeup         bool    `json:"enableMakeup"`
	LipStrength          float64 `json:"lipStrength"`
	RefineBrow           bool    `json:"refineBrow"`
	BrowFill             float64 `json:"browFill"`
	RefineHair           bool    `json:"refineHair"`
	HairColorUniform     float64 `json:"hairColorUniform"`
	HairlineRepair       float64 `json:"hairlineRepair"`
	EnableCloth          bool    `json:"enableCloth"`
	ClothType            string  `json:"clothType"`
	ClothFit             float64 `json:"clothFit"`
	EnableWatermark      bool    `json:"enableWatermark"`
	WatermarkText        string  `json:"watermarkText"`
	WatermarkColor       string  `json:"watermarkColor"`
	WatermarkFontSize    float64 `json:"watermarkFontSize"`
	WatermarkOpacity     float64 `json:"watermarkOpacity"`
	WatermarkAngle       float64 `json:"watermarkAngle"`
	WatermarkSpacing     float64 `json:"watermarkSpacing"`
	GenPrintLayout       bool    `json:"genPrintLayout"`
	LayoutCropLine       bool    `json:"layoutCropLine"`
	PaperSize            string  `json:"paperSize"`
	EnableTargetFileSize bool    `json:"enableTargetFileSize"`
	TargetFileSize       int     `json:"targetFileSize"`
	MaskFeather          float64 `json:"maskFeather"`
	FaceRatio            float64 `json:"faceRatio"`
	HeadTopDistance      float64 `json:"headTopDistance"`
	FaceDetectModel      string  `json:"faceDetectModel"`
	MattingModel         string  `json:"mattingModel"`
	// SourceImg is the current origin photo as a data URL.
	SourceImg string `json:"sourceImg"`
}

// ResultBundle holds multi-type output images (data URLs or base64).
type ResultBundle struct {
	Single  string `json:"single"`
	Layout  string `json:"layout"`
	Social  string `json:"social"`
	Social2 string `json:"social2"`
	IDPhoto string `json:"idphoto"`
}

// FaceCheckResult is the YuNet gate before a photo enters preview.
// Landmarks are five points: right eye, left eye, nose, right mouth, left mouth.
type FaceCheckResult struct {
	OK        bool      `json:"ok"`
	Reason    string    `json:"reason"`
	ImgBase64 string    `json:"imgBase64"`
	FaceBox   []float64 `json:"faceBox"`
	Landmarks []float64 `json:"landmarks"`
	Score     float64   `json:"score"`
}

// FaceCheckOptions toggles optional import prechecks (pose / blur / mosaic / occlusion).
// Missing face or multiple faces are always rejected.
type FaceCheckOptions struct {
	CheckPose   bool `json:"checkPose"`
	CheckBlur   bool `json:"checkBlur"`
	CheckMosaic bool `json:"checkMosaic"`
	CheckParse  bool `json:"checkParse"` // closed eyes / mask / sunglasses / hat
}

// DefaultFaceCheckOptions enables all optional quality gates.
func DefaultFaceCheckOptions() FaceCheckOptions {
	return FaceCheckOptions{
		CheckPose:   true,
		CheckBlur:   true,
		CheckMosaic: true,
		CheckParse:  true,
	}
}

// LoadImageResult is returned after loading a source photo.
type LoadImageResult struct {
	ImgBase64 string    `json:"imgBase64"`
	FaceBox   []float64 `json:"faceBox"`
	Landmarks []float64 `json:"landmarks"`
	Report    Report    `json:"report"`
}

// GenerateResult is returned after finishing an ID photo job.
type GenerateResult struct {
	OriginImg  string       `json:"originImg"`
	MattingImg string       `json:"mattingImg"`
	ResultImg  string       `json:"resultImg"`
	Results    ResultBundle `json:"results"`
	FaceBox    []float64    `json:"faceBox"`
	Landmarks  []float64    `json:"landmarks"`
	Report     Report       `json:"report"`
}

// ExportOptions controls which outputs to write.
type ExportOptions struct {
	Single  bool `json:"single"`
	Layout  bool `json:"layout"`
	Social  bool `json:"social"`
	IDPhoto bool `json:"idphoto"`
}

// ExportResult summarizes an export call.
type ExportResult struct {
	OK  bool   `json:"ok"`
	Dir string `json:"dir"`
}

// PrinterInfo describes an installed system printer.
type PrinterInfo struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

// PrintOptions controls native layout printing.
type PrintOptions struct {
	ImageDataURL string `json:"imageDataUrl"`
	PrinterName  string `json:"printerName"`
	PaperSize    string `json:"paperSize"`
	Copies       int    `json:"copies"`
	Landscape    bool   `json:"landscape"`
}

// PrintResult summarizes a native print job.
type PrintResult struct {
	OK      bool   `json:"ok"`
	Printer string `json:"printer"`
	Copies  int    `json:"copies"`
}
