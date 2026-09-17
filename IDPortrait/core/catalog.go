package core

// SpecCategory groups photo specs in the UI.
type SpecCategory struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// PhotoSpec is one ID-photo size template.
type PhotoSpec struct {
	Value      string   `json:"value"`
	Title      string   `json:"title"`
	Desc       string   `json:"desc"`
	Categories []string `json:"categories"`
	Keywords   string   `json:"keywords"`
	WidthMM    float64  `json:"widthMm,omitempty"`
	HeightMM   float64  `json:"heightMm,omitempty"`
	WidthPx    int      `json:"widthPx,omitempty"`
	HeightPx   int      `json:"heightPx,omitempty"`
	Unit       string   `json:"unit,omitempty"` // mm | px
}

// PaperSpec is one print paper size.
type PaperSpec struct {
	Value    string  `json:"value"`
	Title    string  `json:"title"`
	Desc     string  `json:"desc"`
	WidthMM  float64 `json:"widthMm,omitempty"`
	HeightMM float64 `json:"heightMm,omitempty"`
}

// PhotoSpecCatalog is returned by the photo-spec service.
type PhotoSpecCatalog struct {
	List       []PhotoSpec    `json:"list"`
	Default    string         `json:"default"`
	Current    string         `json:"current"`
	Categories []SpecCategory `json:"categories"`
}

// PaperSpecCatalog is returned by the paper-spec service.
type PaperSpecCatalog struct {
	List    []PaperSpec `json:"list"`
	Default string      `json:"default"`
	Current string      `json:"current"`
}

// SpecQuery filters catalog lists.
type SpecQuery struct {
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
}

// DefaultPhotoSpecID is used when config has no current value.
const DefaultPhotoSpecID = "one_inch"

// DefaultPaperSpecID is used when config has no current value.
const DefaultPaperSpecID = "6inch"

// BuiltinSpecCategories returns UI category tabs.
func BuiltinSpecCategories() []SpecCategory {
	return []SpecCategory{
		{Key: "common", Label: "常用"},
		{Key: "visa", Label: "签职"},
		{Key: "id", Label: "证件"},
		{Key: "school", Label: "升学"},
		{Key: "exam", Label: "考试"},
		{Key: "custom", Label: "自定义"},
	}
}

// BuiltinPhotoSpecs returns the built-in ID photo catalog.
func BuiltinPhotoSpecs() []PhotoSpec {
	return []PhotoSpec{
		{Value: "one_inch", Title: "一寸", Desc: "25×35 mm", Categories: []string{"common", "id"}, Keywords: "一寸 常用", WidthMM: 25, HeightMM: 35, Unit: "mm"},
		{Value: "two_inch", Title: "二寸", Desc: "35×49 mm", Categories: []string{"common", "id"}, Keywords: "二寸 常用", WidthMM: 35, HeightMM: 49, Unit: "mm"},
		{Value: "small_two", Title: "小二寸", Desc: "33×48 mm", Categories: []string{"common"}, Keywords: "小二寸", WidthMM: 33, HeightMM: 48, Unit: "mm"},
		{Value: "large_one", Title: "大一寸", Desc: "33×48 mm", Categories: []string{"common"}, Keywords: "大一寸", WidthMM: 33, HeightMM: 48, Unit: "mm"},
		{Value: "passport", Title: "护照", Desc: "33×48 mm", Categories: []string{"common", "visa", "id"}, Keywords: "护照 出国", WidthMM: 33, HeightMM: 48, Unit: "mm"},
		{Value: "visa_us", Title: "美签", Desc: "51×51 mm", Categories: []string{"visa"}, Keywords: "美签 美国签证", WidthMM: 51, HeightMM: 51, Unit: "mm"},
		{Value: "visa_schengen", Title: "申根签", Desc: "35×45 mm", Categories: []string{"visa"}, Keywords: "申根 欧洲签证", WidthMM: 35, HeightMM: 45, Unit: "mm"},
		{Value: "visa_jp", Title: "日签", Desc: "45×45 mm", Categories: []string{"visa"}, Keywords: "日签 日本签证", WidthMM: 45, HeightMM: 45, Unit: "mm"},
		{Value: "id_card", Title: "身份证", Desc: "26×32 mm", Categories: []string{"id"}, Keywords: "身份证", WidthMM: 26, HeightMM: 32, Unit: "mm"},
		{Value: "driver", Title: "驾驶证", Desc: "22×32 mm", Categories: []string{"id"}, Keywords: "驾驶证 驾照", WidthMM: 22, HeightMM: 32, Unit: "mm"},
		{Value: "social", Title: "社保卡", Desc: "26×32 mm", Categories: []string{"id"}, Keywords: "社保卡", WidthMM: 26, HeightMM: 32, Unit: "mm"},
		{Value: "teacher", Title: "教资", Desc: "25×35 mm · 教师资格", Categories: []string{"exam", "visa"}, Keywords: "教资 教师资格", WidthMM: 25, HeightMM: 35, Unit: "mm"},
		{Value: "civil", Title: "公务员", Desc: "25×35 mm", Categories: []string{"exam", "visa"}, Keywords: "公务员 国考", WidthMM: 25, HeightMM: 35, Unit: "mm"},
		{Value: "grad", Title: "毕业证", Desc: "33×48 mm", Categories: []string{"school"}, Keywords: "毕业证 学历", WidthMM: 33, HeightMM: 48, Unit: "mm"},
		{Value: "student", Title: "学生证", Desc: "25×35 mm", Categories: []string{"school"}, Keywords: "学生证", WidthMM: 25, HeightMM: 35, Unit: "mm"},
		{Value: "exam_cet", Title: "四六级", Desc: "144×192 px", Categories: []string{"exam"}, Keywords: "英语 四六级 CET", WidthPx: 144, HeightPx: 192, Unit: "px"},
		{Value: "exam_cs", Title: "计算机等级", Desc: "144×192 px", Categories: []string{"exam"}, Keywords: "计算机等级 NCRE", WidthPx: 144, HeightPx: 192, Unit: "px"},
		{Value: "exam_nurse", Title: "护士资格", Desc: "25×35 mm", Categories: []string{"exam"}, Keywords: "护士资格证", WidthMM: 25, HeightMM: 35, Unit: "mm"},
	}
}

// BuiltinPaperSpecs returns the built-in print paper catalog.
func BuiltinPaperSpecs() []PaperSpec {
	return []PaperSpec{
		{Value: "5inch", Title: "5寸", Desc: "89×127 mm", WidthMM: 89, HeightMM: 127},
		{Value: "6inch", Title: "6寸", Desc: "102×152 mm", WidthMM: 102, HeightMM: 152},
		{Value: "7inch", Title: "7寸", Desc: "127×178 mm", WidthMM: 127, HeightMM: 178},
		{Value: "a6", Title: "A6", Desc: "105×148 mm", WidthMM: 105, HeightMM: 148},
		{Value: "a5", Title: "A5", Desc: "148×210 mm", WidthMM: 148, HeightMM: 210},
		{Value: "a4", Title: "A4", Desc: "210×297 mm", WidthMM: 210, HeightMM: 297},
	}
}

// ModelOption describes a selectable AI model.
type ModelOption struct {
	Value    string `json:"value"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	Keywords string `json:"keywords,omitempty"`
}

// ModelCatalog is returned for face-detect / matting model lists.
type ModelCatalog struct {
	List    []ModelOption `json:"list"`
	Default string        `json:"default"`
	Current string        `json:"current"`
}

const (
	DefaultFaceDetectModelID = "retinaface"
	DefaultMattingModelID    = "modnet"
)

// BuiltinFaceDetectModels returns available face detection models.
func BuiltinFaceDetectModels() []ModelOption {
	return []ModelOption{
		{Value: "retinaface", Title: "RetinaFace", Desc: "高精度人脸框与五点", Keywords: "retinaface 人脸检测"},
		{Value: "scrfd", Title: "SCRFD", Desc: "轻量快速，适合实时", Keywords: "scrfd 轻量"},
		{Value: "yolov8_face", Title: "YOLOv8-Face", Desc: "通用目标检测风格", Keywords: "yolo yolov8"},
		{Value: "mediapipe", Title: "MediaPipe", Desc: "移动端友好", Keywords: "mediapipe google"},
		{Value: "insightface", Title: "InsightFace", Desc: "检测+关键点一体化", Keywords: "insightface 关键点"},
	}
}

// BuiltinMattingModels returns available matting / cutout models.
func BuiltinMattingModels() []ModelOption {
	return []ModelOption{
		{Value: "modnet", Title: "MODNet", Desc: "人像抠图，边缘自然", Keywords: "modnet 抠图"},
		{Value: "u2net", Title: "U²-Net", Desc: "通用显著物体分割", Keywords: "u2net rembg"},
		{Value: "birefnet", Title: "BiRefNet", Desc: "高细节抠图", Keywords: "birefnet 精细"},
		{Value: "isnet", Title: "ISNet", Desc: "人像分割增强", Keywords: "isnet"},
		{Value: "rmbg", Title: "RMBG", Desc: "背景移除专用", Keywords: "rmbg 去背景"},
	}
}
