package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"IDPortrait/core"
	"IDPortrait/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server wraps Echo HTTPS server for remote browser access.
// HTTPS is required so browsers allow camera (getUserMedia) on LAN IPs.
type Server struct {
	mu          sync.Mutex
	router      *echo.Echo
	svc         service.IDPhotoService
	assets      fs.FS
	devFrontend string // e.g. http://127.0.0.1:5173 during wails dev
	running     bool
	addr        string
	port        int
}

func New(svc service.IDPhotoService, assets fs.FS, devFrontend string) *Server {
	return &Server{svc: svc, assets: assets, devFrontend: strings.TrimRight(strings.TrimSpace(devFrontend), "/")}
}

// SetDevFrontend updates the Vite proxy target (useful when Vite starts after the app).
func (s *Server) SetDevFrontend(devFrontend string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.devFrontend = strings.TrimRight(strings.TrimSpace(devFrontend), "/")
}


// Start listens on 0.0.0.0:port with a self-signed TLS cert and serves SPA + /api.
func (s *Server) Start(port int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return s.addr, nil
	}
	if port <= 0 || port > 65535 {
		port = 8787
	}

	cert, err := selfSignedCert()
	if err != nil {
		return "", fmt.Errorf("generate tls cert: %w", err)
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
	api.POST("/detect-face", s.handleDetectFace)
	api.POST("/generate", s.handleGenerate)
	api.POST("/export", s.handleExport)
	api.GET("/photo-specs", s.handlePhotoSpecs)
	api.GET("/paper-specs", s.handlePaperSpecs)
	api.GET("/face-detect-models", s.handleFaceDetectModels)
	api.GET("/matting-models", s.handleMattingModels)
	api.GET("/watermark-config", s.handleWatermarkConfig)
	api.POST("/photo-specs/current", s.handleSetPhotoSpec)
	api.POST("/paper-specs/current", s.handleSetPaperSpec)
	api.POST("/face-detect-models/current", s.handleSetFaceDetectModel)
	api.POST("/matting-models/current", s.handleSetMattingModel)
	api.POST("/watermark-config", s.handleSetWatermarkConfig)

	s.mountFrontend(e)

	tlsCfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}
	ln, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port), tlsCfg)
	if err != nil {
		return "", err
	}

	s.router = e
	s.port = port
	s.addr = fmt.Sprintf("https://%s:%d", localIP(), port)
	s.running = true

	go func() {
		_ = e.Server.Serve(ln)
	}()

	return s.addr, nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || s.router == nil {
		s.running = false
		return nil
	}
	err := s.router.Close()
	s.running = false
	s.router = nil
	s.addr = ""
	return err
}

func (s *Server) Status() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	frontend := "embed"
	if s.devFrontend != "" {
		frontend = "dev-proxy:" + s.devFrontend
	} else if s.assets != nil {
		frontend = "assets"
	}
	return map[string]any{
		"running":  s.running,
		"port":     s.port,
		"addr":     s.addr,
		"url":      s.addr,
		"tls":      true,
		"frontend": frontend,
	}
}

func (s *Server) mountFrontend(e *echo.Echo) {
	if s.devFrontend != "" {
		target, err := url.Parse(s.devFrontend)
		if err == nil {
			proxy := httputil.NewSingleHostReverseProxy(target)
			// Stream websockets / HMR without buffering.
			proxy.FlushInterval = -1
			defaultDirector := proxy.Director
			proxy.Director = func(req *http.Request) {
				origHost := req.Host
				if fwd := req.Header.Get("X-Forwarded-Host"); fwd != "" {
					origHost = fwd
				}
				defaultDirector(req)
				req.Host = target.Host
				req.Header.Set("X-Forwarded-Proto", "https")
				if origHost != "" {
					req.Header.Set("X-Forwarded-Host", origHost)
				}
			}
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, "frontend dev server unavailable: "+err.Error(), http.StatusBadGateway)
			}
			e.Any("/*", func(c echo.Context) error {
				proxy.ServeHTTP(c.Response(), c.Request())
				return nil
			})
			return
		}
	}

	if s.assets != nil {
		fileServer := http.FileServer(http.FS(s.assets))
		e.GET("/*", func(c echo.Context) error {
			path := c.Request().URL.Path
			if path == "/" || path == "" {
				return serveIndex(c, s.assets)
			}
			f, err := s.assets.Open(strings.TrimPrefix(path, "/"))
			if err != nil {
				return serveIndex(c, s.assets)
			}
			_ = f.Close()
			fileServer.ServeHTTP(c.Response(), c.Request())
			return nil
		})
		return
	}

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "<h1>IDPortrait remote API</h1><p>/api/health</p>")
	})
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

type detectFaceReq struct {
	Path       string                 `json:"path"`
	Image      string                 `json:"image"`
	SkipPose   *bool                  `json:"skipPose"`
	SkipBlur   *bool                  `json:"skipBlur"`
	SkipMosaic *bool                  `json:"skipMosaic"`
	SkipParse  *bool                  `json:"skipParse"`
	Options    *core.FaceCheckOptions `json:"options"`
}

func (s *Server) handleDetectFace(c echo.Context) error {
	var req detectFaceReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	src := strings.TrimSpace(req.Path)
	if src == "" {
		src = strings.TrimSpace(req.Image)
	}
	opts := core.DefaultFaceCheckOptions()
	if req.Options != nil {
		opts = *req.Options
	} else {
		if req.SkipPose != nil {
			opts.SkipPose = *req.SkipPose
		}
		if req.SkipBlur != nil {
			opts.SkipBlur = *req.SkipBlur
		}
		if req.SkipMosaic != nil {
			opts.SkipMosaic = *req.SkipMosaic
		}
		if req.SkipParse != nil {
			opts.SkipParse = *req.SkipParse
		}
	}
	res, err := s.svc.DetectFace(src, opts)
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

func (s *Server) handlePhotoSpecs(c echo.Context) error {
	q := core.SpecQuery{
		Keyword:  c.QueryParam("keyword"),
		Category: c.QueryParam("category"),
	}
	res, err := s.svc.GetPhotoSpecs(q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

func (s *Server) handlePaperSpecs(c echo.Context) error {
	q := core.SpecQuery{Keyword: c.QueryParam("keyword")}
	res, err := s.svc.GetPaperSpecs(q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

type currentSpecReq struct {
	Value string `json:"value"`
}

func (s *Server) handleSetPhotoSpec(c echo.Context) error {
	var req currentSpecReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.svc.SetCurrentPhotoSpec(req.Value); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"ok": true, "value": req.Value})
}

func (s *Server) handleSetPaperSpec(c echo.Context) error {
	var req currentSpecReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.svc.SetCurrentPaperSpec(req.Value); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"ok": true, "value": req.Value})
}

func (s *Server) handleFaceDetectModels(c echo.Context) error {
	q := core.SpecQuery{Keyword: c.QueryParam("keyword")}
	res, err := s.svc.GetFaceDetectModels(q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

func (s *Server) handleMattingModels(c echo.Context) error {
	q := core.SpecQuery{Keyword: c.QueryParam("keyword")}
	res, err := s.svc.GetMattingModels(q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

func (s *Server) handleSetFaceDetectModel(c echo.Context) error {
	var req currentSpecReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.svc.SetCurrentFaceDetectModel(req.Value); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"ok": true, "value": req.Value})
}

func (s *Server) handleSetMattingModel(c echo.Context) error {
	var req currentSpecReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.svc.SetCurrentMattingModel(req.Value); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"ok": true, "value": req.Value})
}

func (s *Server) handleWatermarkConfig(c echo.Context) error {
	res, err := s.svc.GetWatermarkConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, res)
}

func (s *Server) handleSetWatermarkConfig(c echo.Context) error {
	var req core.WatermarkSettings
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	if err := s.svc.SetWatermarkConfig(req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"ok": true})
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

func selfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	ip := net.ParseIP(localIP())
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"IDPortrait"},
			CommonName:   "IDPortrait Remote",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}
	if ip != nil {
		tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return tls.X509KeyPair(certPEM, keyPEM)
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
