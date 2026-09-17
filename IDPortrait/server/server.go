package server

import (
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"sync"

	"IDPortrait/core"
	"IDPortrait/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server wraps Echo HTTP server for remote browser access.
type Server struct {
	mu      sync.Mutex
	echo    *echo.Echo
	svc     service.IDPhotoService
	assets  fs.FS
	running bool
	addr    string
	port    int
}

func New(svc service.IDPhotoService, assets fs.FS) *Server {
	return &Server{svc: svc, assets: assets}
}

// Start listens on 0.0.0.0:port and serves SPA + /api.
func (s *Server) Start(port int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return s.addr, nil
	}
	if port <= 0 || port > 65535 {
		port = 8787
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderContentType},
	}))

	api := e.Group("/api")
	api.GET("/health", s.handleHealth)
	api.POST("/load-image", s.handleLoadImage)
	api.POST("/generate", s.handleGenerate)
	api.POST("/export", s.handleExport)

	if s.assets != nil {
		fileServer := http.FileServer(http.FS(s.assets))
		e.GET("/*", func(c echo.Context) error {
			path := c.Request().URL.Path
			if path == "/" || path == "" {
				return serveIndex(c, s.assets)
			}
			// try static file; fall back to SPA index
			f, err := s.assets.Open(strings.TrimPrefix(path, "/"))
			if err != nil {
				return serveIndex(c, s.assets)
			}
			_ = f.Close()
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		})
	} else {
		e.GET("/", func(c echo.Context) error {
			return c.HTML(http.StatusOK, "<h1>IDPortrait remote API</h1><p>/api/health</p>")
		})
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return "", err
	}
	s.echo = e
	s.port = port
	s.addr = fmt.Sprintf("http://%s:%d", localIP(), port)
	s.running = true

	go func() {
		_ = e.Server.Serve(ln)
	}()

	return s.addr, nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || s.echo == nil {
		s.running = false
		return nil
	}
	err := s.echo.Close()
	s.running = false
	s.echo = nil
	s.addr = ""
	return err
}

func (s *Server) Status() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"running": s.running,
		"port":    s.port,
		"addr":    s.addr,
		"url":     s.addr,
	}
}

func (s *Server) handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, s.svc.Health())
}

type loadImageReq struct {
	Path string `json:"path"`
}

func (s *Server) handleLoadImage(c echo.Context) error {
	var req loadImageReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	res, err := s.svc.LoadImage(req.Path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

func (s *Server) handleGenerate(c echo.Context) error {
	var params core.GenerateParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	res, err := s.svc.Generate(params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

type exportReq struct {
	Dir     string             `json:"dir"`
	Options core.ExportOptions `json:"options"`
}

func (s *Server) handleExport(c echo.Context) error {
	var req exportReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	res, err := s.svc.Export(req.Dir, req.Options)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

func serveIndex(c echo.Context, assets fs.FS) error {
	f, err := assets.Open("index.html")
	if err != nil {
		return c.String(http.StatusNotFound, "index.html not found")
	}
	defer f.Close()
	c.Response().Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Response(), f)
	return err
}

func localIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			return ip.String()
		}
	}
	return "127.0.0.1"
}
