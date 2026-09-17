package main

import (
	"context"
	"fmt"
	"io/fs"

	"IDPortrait/core"
	"IDPortrait/server"
	"IDPortrait/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application facade. Business logic lives in service/core;
// remote HTTP is served by server package.
type App struct {
	ctx     context.Context
	svc     *service.Service
	remote  *server.Server
	assets  fs.FS
}

// NewApp creates a new App application struct
func NewApp(assets fs.FS) *App {
	engine := core.NewEngine()
	svc := service.New(engine)
	return &App{
		svc:    svc,
		remote: server.New(svc, assets),
		assets: assets,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) domReady(ctx context.Context) {}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.StopRemoteServer()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// --- ID photo APIs (shared with HTTP via service) ---

func (a *App) LoadImage(path string) (*core.LoadImageResult, error) {
	return a.svc.LoadImage(path)
}

func (a *App) Generate(params core.GenerateParams) (*core.GenerateResult, error) {
	return a.svc.Generate(params)
}

func (a *App) Export(dir string, opt core.ExportOptions) (*core.ExportResult, error) {
	return a.svc.Export(dir, opt)
}

func (a *App) GetPhotoSpecs(query core.SpecQuery) (*core.PhotoSpecCatalog, error) {
	return a.svc.GetPhotoSpecs(query)
}

func (a *App) GetPaperSpecs(query core.SpecQuery) (*core.PaperSpecCatalog, error) {
	return a.svc.GetPaperSpecs(query)
}

func (a *App) SetCurrentPhotoSpec(value string) error {
	return a.svc.SetCurrentPhotoSpec(value)
}

func (a *App) SetCurrentPaperSpec(value string) error {
	return a.svc.SetCurrentPaperSpec(value)
}

func (a *App) GetFaceDetectModels(query core.SpecQuery) (*core.ModelCatalog, error) {
	return a.svc.GetFaceDetectModels(query)
}

func (a *App) GetMattingModels(query core.SpecQuery) (*core.ModelCatalog, error) {
	return a.svc.GetMattingModels(query)
}

func (a *App) SetCurrentFaceDetectModel(value string) error {
	return a.svc.SetCurrentFaceDetectModel(value)
}

func (a *App) SetCurrentMattingModel(value string) error {
	return a.svc.SetCurrentMattingModel(value)
}

func (a *App) GetWatermarkConfig() (*core.WatermarkConfigResult, error) {
	return a.svc.GetWatermarkConfig()
}

func (a *App) SetWatermarkConfig(settings core.WatermarkSettings) error {
	return a.svc.SetWatermarkConfig(settings)
}

func (a *App) OpenImageDialog() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not ready")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择照片",
		Filters: []runtime.FileFilter{
			{DisplayName: "Images", Pattern: "*.jpg;*.jpeg;*.png;*.bmp;*.webp"},
		},
	})
}

func (a *App) SelectFolder() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not ready")
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择导出目录",
	})
}

// --- Remote HTTP service ---

// StartRemoteServer starts Echo HTTP server so other devices can open the UI.
func (a *App) StartRemoteServer(port int) (map[string]any, error) {
	addr, err := a.remote.Start(port)
	if err != nil {
		return nil, err
	}
	st := a.remote.Status()
	st["addr"] = addr
	st["url"] = addr
	return st, nil
}

// StopRemoteServer stops the remote HTTP server.
func (a *App) StopRemoteServer() error {
	return a.remote.Stop()
}

// GetRemoteStatus returns remote server running state.
func (a *App) GetRemoteStatus() map[string]any {
	return a.remote.Status()
}
