package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"IDPortrait/core"
)

// userConfig holds persisted preferences (current selections).
type userConfig struct {
	CurrentPhotoSpec       string                  `json:"currentPhotoSpec"`
	CurrentPaperSpec       string                  `json:"currentPaperSpec"`
	CurrentFaceDetectModel string                  `json:"currentFaceDetectModel"`
	CurrentMattingModel    string                  `json:"currentMattingModel"`
	Watermark              *core.WatermarkSettings `json:"watermark,omitempty"`
}

func (s *Service) configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "IDPortrait", "config.json")
}

func (s *Service) loadConfig() userConfig {
	s.cfgMu.Lock()
	defer s.cfgMu.Unlock()
	if s.cfgLoaded {
		return s.cfg
	}
	s.cfgLoaded = true
	path := s.configPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		return s.cfg
	}
	_ = json.Unmarshal(raw, &s.cfg)
	return s.cfg
}

func (s *Service) saveConfig() error {
	s.cfgMu.Lock()
	defer s.cfgMu.Unlock()
	path := s.configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func containsFold(hay, needle string) bool {
	return strings.Contains(strings.ToLower(hay), strings.ToLower(needle))
}

func hasCategory(cats []string, key string) bool {
	if key == "" || key == "custom" {
		return true
	}
	for _, c := range cats {
		if c == key {
			return true
		}
	}
	return false
}

func photoSpecExists(id string) bool {
	for _, item := range core.BuiltinPhotoSpecs() {
		if item.Value == id {
			return true
		}
	}
	return false
}

func paperSpecExists(id string) bool {
	for _, item := range core.BuiltinPaperSpecs() {
		if item.Value == id {
			return true
		}
	}
	return false
}

// GetPhotoSpecs returns ID-photo specs with search/category filter.
// Current comes from config; Default is the built-in fallback.
func (s *Service) GetPhotoSpecs(query core.SpecQuery) (*core.PhotoSpecCatalog, error) {
	cfg := s.loadConfig()
	keyword := strings.TrimSpace(query.Keyword)
	category := strings.TrimSpace(query.Category)

	all := core.BuiltinPhotoSpecs()
	list := make([]core.PhotoSpec, 0, len(all))
	for _, item := range all {
		if keyword == "" {
			if category != "" && category != "custom" && !hasCategory(item.Categories, category) {
				continue
			}
		} else {
			hay := item.Title + " " + item.Desc + " " + item.Keywords + " " + item.Value
			if !containsFold(hay, keyword) {
				continue
			}
		}
		list = append(list, item)
	}

	current := cfg.CurrentPhotoSpec
	if current != "" && !photoSpecExists(current) {
		current = ""
	}

	return &core.PhotoSpecCatalog{
		List:       list,
		Default:    core.DefaultPhotoSpecID,
		Current:    current,
		Categories: core.BuiltinSpecCategories(),
	}, nil
}

// GetPaperSpecs returns print paper specs with optional keyword search.
func (s *Service) GetPaperSpecs(query core.SpecQuery) (*core.PaperSpecCatalog, error) {
	cfg := s.loadConfig()
	keyword := strings.TrimSpace(query.Keyword)

	all := core.BuiltinPaperSpecs()
	list := make([]core.PaperSpec, 0, len(all))
	for _, item := range all {
		if keyword != "" {
			hay := item.Title + " " + item.Desc + " " + item.Value
			if !containsFold(hay, keyword) {
				continue
			}
		}
		list = append(list, item)
	}

	current := cfg.CurrentPaperSpec
	if current != "" && !paperSpecExists(current) {
		current = ""
	}

	return &core.PaperSpecCatalog{
		List:    list,
		Default: core.DefaultPaperSpecID,
		Current: current,
	}, nil
}

// SetCurrentPhotoSpec persists the selected photo spec into config.
func (s *Service) SetCurrentPhotoSpec(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || value == "custom" {
		s.cfgMu.Lock()
		s.cfg.CurrentPhotoSpec = ""
		s.cfgLoaded = true
		s.cfgMu.Unlock()
		return s.saveConfig()
	}
	if !photoSpecExists(value) {
		return nil
	}
	s.cfgMu.Lock()
	s.cfg.CurrentPhotoSpec = value
	s.cfgLoaded = true
	s.cfgMu.Unlock()
	return s.saveConfig()
}

// SetCurrentPaperSpec persists the selected paper spec into config.
func (s *Service) SetCurrentPaperSpec(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		s.cfgMu.Lock()
		s.cfg.CurrentPaperSpec = ""
		s.cfgLoaded = true
		s.cfgMu.Unlock()
		return s.saveConfig()
	}
	if !paperSpecExists(value) {
		return nil
	}
	s.cfgMu.Lock()
	s.cfg.CurrentPaperSpec = value
	s.cfgLoaded = true
	s.cfgMu.Unlock()
	return s.saveConfig()
}

// PreferSpec returns current if set, otherwise default.
func PreferSpec(current, fallback string) string {
	if strings.TrimSpace(current) != "" {
		return current
	}
	return fallback
}

func faceDetectModelExists(id string) bool {
	for _, item := range core.BuiltinFaceDetectModels() {
		if item.Value == id {
			return true
		}
	}
	return false
}

func mattingModelExists(id string) bool {
	for _, item := range core.BuiltinMattingModels() {
		if item.Value == id {
			return true
		}
	}
	return false
}

func filterModels(all []core.ModelOption, keyword string) []core.ModelOption {
	keyword = strings.TrimSpace(keyword)
	list := make([]core.ModelOption, 0, len(all))
	for _, item := range all {
		if keyword != "" {
			hay := item.Title + " " + item.Desc + " " + item.Keywords + " " + item.Value
			if !containsFold(hay, keyword) {
				continue
			}
		}
		list = append(list, item)
	}
	return list
}

// GetFaceDetectModels returns face detection model options.
func (s *Service) GetFaceDetectModels(query core.SpecQuery) (*core.ModelCatalog, error) {
	cfg := s.loadConfig()
	current := cfg.CurrentFaceDetectModel
	if current != "" && !faceDetectModelExists(current) {
		current = ""
	}
	return &core.ModelCatalog{
		List:    filterModels(core.BuiltinFaceDetectModels(), query.Keyword),
		Default: core.DefaultFaceDetectModelID,
		Current: current,
	}, nil
}

// GetMattingModels returns matting / cutout model options.
func (s *Service) GetMattingModels(query core.SpecQuery) (*core.ModelCatalog, error) {
	cfg := s.loadConfig()
	current := cfg.CurrentMattingModel
	if current != "" && !mattingModelExists(current) {
		current = ""
	}
	return &core.ModelCatalog{
		List:    filterModels(core.BuiltinMattingModels(), query.Keyword),
		Default: core.DefaultMattingModelID,
		Current: current,
	}, nil
}

// SetCurrentFaceDetectModel persists the selected face detection model.
func (s *Service) SetCurrentFaceDetectModel(value string) error {
	value = strings.TrimSpace(value)
	if value != "" && !faceDetectModelExists(value) {
		return nil
	}
	s.cfgMu.Lock()
	s.cfg.CurrentFaceDetectModel = value
	s.cfgLoaded = true
	s.cfgMu.Unlock()
	return s.saveConfig()
}

// SetCurrentMattingModel persists the selected matting model.
func (s *Service) SetCurrentMattingModel(value string) error {
	value = strings.TrimSpace(value)
	if value != "" && !mattingModelExists(value) {
		return nil
	}
	s.cfgMu.Lock()
	s.cfg.CurrentMattingModel = value
	s.cfgLoaded = true
	s.cfgMu.Unlock()
	return s.saveConfig()
}

// GetWatermarkConfig returns default and current watermark settings.
func (s *Service) GetWatermarkConfig() (*core.WatermarkConfigResult, error) {
	cfg := s.loadConfig()
	out := &core.WatermarkConfigResult{
		Default: core.DefaultWatermarkSettings(),
	}
	if cfg.Watermark != nil {
		cp := *cfg.Watermark
		normalizeWatermark(&cp)
		out.Current = &cp
	}
	return out, nil
}

// SetWatermarkConfig persists the current watermark settings.
func (s *Service) SetWatermarkConfig(settings core.WatermarkSettings) error {
	normalizeWatermark(&settings)
	cp := settings
	s.cfgMu.Lock()
	s.cfg.Watermark = &cp
	s.cfgLoaded = true
	s.cfgMu.Unlock()
	return s.saveConfig()
}

func normalizeWatermark(w *core.WatermarkSettings) {
	if w == nil {
		return
	}
	if strings.TrimSpace(w.Text) == "" {
		w.Text = "最美证件照"
	}
	if strings.TrimSpace(w.Color) == "" {
		w.Color = "#FFFFFF"
	}
	if w.FontSize < 8 {
		w.FontSize = 8
	}
	if w.FontSize > 72 {
		w.FontSize = 72
	}
	if w.Opacity < 0 {
		w.Opacity = 0
	}
	if w.Opacity > 1 {
		w.Opacity = 1
	}
	if w.Angle < -90 {
		w.Angle = -90
	}
	if w.Angle > 90 {
		w.Angle = 90
	}
	if w.Spacing < 40 {
		w.Spacing = 40
	}
	if w.Spacing > 400 {
		w.Spacing = 400
	}
}
