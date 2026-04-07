package docs

import (
	"errors"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
)

type Module struct {
	publicDir string
}

func NewModule(publicDir string) *Module {
	return &Module{publicDir: publicDir}
}

func (m *Module) Handle(c *echo.Context) error {
	root := filepath.Join(m.publicDir, "doc")
	target, isAsset, err := resolvePath(root, c.Request().URL.Path)
	if err != nil {
		return echo.ErrNotFound
	}
	if _, err := os.Stat(target); err == nil {
		if isAsset {
			c.Response().Header().Set("Cache-Control", "public, max-age=3600")
		} else {
			c.Response().Header().Set("Cache-Control", "no-cache")
		}
		return serveFile(c, http.StatusOK, target)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if isAsset {
		return echo.ErrNotFound
	}

	notFoundPath := filepath.Join(root, "404.html")
	if _, err := os.Stat(notFoundPath); err == nil {
		return serveFile(c, http.StatusNotFound, notFoundPath)
	}
	return echo.ErrNotFound
}

func serveFile(c *echo.Context, status int, filename string) error {
	body, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename)))
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	return c.Blob(status, contentType, body)
}

func resolvePath(root, requestPath string) (string, bool, error) {
	if root == "" {
		return "", false, errors.New("docs root required")
	}
	if requestPath == "" || (requestPath != "/docs" && !strings.HasPrefix(requestPath, "/docs/")) {
		return "", false, errors.New("invalid docs path")
	}

	rel := strings.TrimPrefix(requestPath, "/docs")
	if rel == "" || rel == "/" {
		return filepath.Join(root, "index.html"), false, nil
	}
	if strings.Contains(rel, "..") {
		return "", false, errors.New("path traversal")
	}

	cleanRel := strings.TrimPrefix(path.Clean("/"+rel), "/")
	if cleanRel == "" || cleanRel == "." {
		return filepath.Join(root, "index.html"), false, nil
	}

	if ext := path.Ext(cleanRel); ext != "" {
		return filepath.Join(root, filepath.FromSlash(cleanRel)), true, nil
	}

	return filepath.Join(root, filepath.FromSlash(cleanRel), "index.html"), false, nil
}
