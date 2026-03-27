package app

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

func docsHandler(publicDir string) echo.HandlerFunc {
	root := filepath.Join(publicDir, "doc")

	return func(c *echo.Context) error {
		target, isAsset, err := resolveDocsPath(root, c.Request().URL.Path)
		if err != nil {
			return echo.ErrNotFound
		}
		if _, err := os.Stat(target); err == nil {
			if isAsset {
				c.Response().Header().Set("Cache-Control", "public, max-age=3600")
			} else {
				c.Response().Header().Set("Cache-Control", "no-cache")
			}
			return serveDocsFile(c, http.StatusOK, target)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if isAsset {
			return echo.ErrNotFound
		}

		notFoundPath := filepath.Join(root, "404.html")
		if _, err := os.Stat(notFoundPath); err == nil {
			return serveDocsFile(c, http.StatusNotFound, notFoundPath)
		}
		return echo.ErrNotFound
	}
}

func serveDocsFile(c *echo.Context, status int, filename string) error {
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

func resolveDocsPath(root, requestPath string) (string, bool, error) {
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
