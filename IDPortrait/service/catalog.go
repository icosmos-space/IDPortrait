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
	CurrentPhotoSpec string `json:"currentPhotoSpec"`
	CurrentPaperSpec string `json:"currentPaperSpec"`
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
