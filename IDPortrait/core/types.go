package core

// Report is the face / authenticity diagnosis result.
type Report struct {
	FaceOK     bool    `json:"faceOk"`
	FaceScore  float64 `json:"faceScore"`
	IsRephoto  bool    `json:"isRephoto"`
	IsAiImage  bool    `json:"isAiImage"`
}

// GenerateParams mirrors frontend generation options.
type GenerateParams struct {
	Template       string  `json:"template"`
	BgMode         string  `json:"bgMode"`
	BgColor        string  `json:"bgColor"`
	BgPreset       string  `json:"bgPreset"`
	CustomWidth    float64 `json:"customWidth"`
	CustomHeight   float64 `json:"customHeight"`
	BeautyStrength float64 `json:"beautyStrength"`
	EyeSharp       float64 `json:"eyeSharp"`
	SkinBright     float64 `json:"skinBright"`
	EnableMakeup   bool    `json:"enableMakeup"`
	LipStrength    float64 `json:"lipStrength"`
	RefineBrow     bool    `json:"refineBrow"`
	BrowFill       float64 `json:"browFill"`
	RefineHair     bool    `json:"refineHair"`
	HairColorUniform float64 `json:"hairColorUniform"`
	HairlineRepair float64 `json:"hairlineRepair"`
	EnableCloth    bool    `json:"enableCloth"`
	ClothType      string  `json:"clothType"`
	ClothFit       float64 `json:"clothFit"`
	AddWatermark   bool    `json:"addWatermark"`
	GenPrintLayout bool    `json:"genPrintLayout"`
	PaperSize      string  `json:"paperSize"`
	EnableTargetFileSize bool `json:"enableTargetFileSize"`
	TargetFileSize int     `json:"targetFileSize"`
	MaskFeather    float64 `json:"maskFeather"`
	// SourceImg is the current origin photo as a data URL.
	SourceImg string `json:"sourceImg"`
}

// ResultBundle holds multi-type output images (data URLs or base64).
type ResultBundle struct {
	Single  string `json:"single"`
	Layout  string `json:"layout"`
	Social  string `json:"social"`
	IDPhoto string `json:"idphoto"`
}

// LoadImageResult is returned after loading a source photo.
type LoadImageResult struct {
	ImgBase64  string    `json:"imgBase64"`
	FaceBox    []float64 `json:"faceBox"`
	Landmarks  []float64 `json:"landmarks"`
	Report     Report    `json:"report"`
}

// GenerateResult is returned after finishing an ID photo job.
type GenerateResult struct {
	OriginImg  string       `json:"originImg"`
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
