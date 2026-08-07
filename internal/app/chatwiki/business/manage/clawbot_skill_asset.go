// Copyright © 2016- 2025 Wuhan Sesame Small Customer Service Network Technology Co., Ltd.

package manage

import (
	"chatwiki/internal/app/chatwiki/define"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var clawbotSkillAssetContentTypes = map[string]string{
	`.apng`:  `image/apng`,
	`.avif`:  `image/avif`,
	`.bmp`:   `image/bmp`,
	`.cur`:   `image/x-icon`,
	`.dib`:   `image/bmp`,
	`.gif`:   `image/gif`,
	`.heic`:  `image/heic`,
	`.heif`:  `image/heif`,
	`.ico`:   `image/x-icon`,
	`.jfi`:   `image/jpeg`,
	`.jfif`:  `image/jpeg`,
	`.jif`:   `image/jpeg`,
	`.jpe`:   `image/jpeg`,
	`.jpeg`:  `image/jpeg`,
	`.jpg`:   `image/jpeg`,
	`.jxl`:   `image/jxl`,
	`.pjp`:   `image/jpeg`,
	`.pjpeg`: `image/jpeg`,
	`.png`:   `image/png`,
	`.svg`:   `image/svg+xml`,
	`.svgz`:  `image/svg+xml`,
	`.tif`:   `image/tiff`,
	`.tiff`:  `image/tiff`,
	`.webp`:  `image/webp`,
}

// GetClawbotSkillAsset serves image files below the clawbot/skills_robot directory.
func GetClawbotSkillAsset(c *gin.Context) {
	privateSkillsDir := filepath.Clean(filepath.FromSlash(define.PrivateSkillsDir))
	skillsRootDir := filepath.Dir(privateSkillsDir)
	assetPath := strings.TrimPrefix(c.Param(`asset_path`), `/`)
	localAssetPath := filepath.Clean(filepath.FromSlash(assetPath))
	if !filepath.IsLocal(localAssetPath) {
		c.Status(http.StatusNotFound)
		return
	}
	contentType, ok := clawbotSkillAssetContentTypes[strings.ToLower(filepath.Ext(localAssetPath))]
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	skillsRoot, err := os.OpenRoot(skillsRootDir)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer skillsRoot.Close()

	file, err := skillsRoot.Open(localAssetPath)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header(`Content-Type`, contentType)
	c.Header(`X-Content-Type-Options`, `nosniff`)
	c.Header(`Cache-Control`, `public, max-age=300`)
	if strings.EqualFold(filepath.Ext(localAssetPath), `.svgz`) {
		c.Header(`Content-Encoding`, `gzip`)
	}
	if contentType == `image/svg+xml` {
		c.Header(`Content-Security-Policy`, `sandbox; default-src 'none'; style-src 'unsafe-inline'; img-src data:`)
	}
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
}
