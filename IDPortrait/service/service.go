package service

import "IDPortrait/core"

// IDPhotoService is the shared contract for Wails bindings and HTTP API.
type IDPhotoService interface {
	LoadImage(path string) (*core.LoadImageResult, error)
	Generate(params core.GenerateParams) (*core.GenerateResult, error)
	Export(dir string, opt core.ExportOptions) (*core.ExportResult, error)
	Health() map[string]any
}

// Service is the default implementation backed by core.Engine.
type Service struct {
	engine *core.Engine
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
