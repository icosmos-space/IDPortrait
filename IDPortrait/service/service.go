package service

import (
	"sync"

	"IDPortrait/core"
)

// IDPhotoService is the shared contract for Wails bindings and HTTP API.
type IDPhotoService interface {
	LoadImage(path string) (*core.LoadImageResult, error)
	Generate(params core.GenerateParams) (*core.GenerateResult, error)
	Export(dir string, opt core.ExportOptions) (*core.ExportResult, error)
	Health() map[string]any
	GetPhotoSpecs(query core.SpecQuery) (*core.PhotoSpecCatalog, error)
	GetPaperSpecs(query core.SpecQuery) (*core.PaperSpecCatalog, error)
	GetFaceDetectModels(query core.SpecQuery) (*core.ModelCatalog, error)
	GetMattingModels(query core.SpecQuery) (*core.ModelCatalog, error)
	SetCurrentPhotoSpec(value string) error
	SetCurrentPaperSpec(value string) error
	SetCurrentFaceDetectModel(value string) error
	SetCurrentMattingModel(value string) error
}

// Service is the default implementation backed by core.Engine.
type Service struct {
	engine    *core.Engine
	cfgMu     sync.Mutex
	cfg       userConfig
	cfgLoaded bool
}

func New(engine *core.Engine) *Service {
	if engine == nil {
		engine = core.NewEngine()
	}
	return &Service{engine: engine}
}

func (s *Service) LoadImage(path string) (*core.LoadImageResult, error) {
	return s.engine.LoadImage(path)
}

func (s *Service) Generate(params core.GenerateParams) (*core.GenerateResult, error) {
	return s.engine.Generate(params)
}

func (s *Service) Export(dir string, opt core.ExportOptions) (*core.ExportResult, error) {
	return s.engine.Export(dir, opt)
}

func (s *Service) Health() map[string]any {
	return map[string]any{
		"ok":      true,
		"service": "idportrait",
		"mode":    "shared",
	}
}
