package router

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdminFrontendConfig struct {
	DistDir      string
	DevServerURL string
}

const DefaultAdminFrontendDistDir = "admin/build"
const DefaultAdminFrontendDevServerURL = "http://localhost:5173"

func defaultAdminFrontendConfig(appMode string) AdminFrontendConfig {
	cfg := AdminFrontendConfig{
		DistDir: DefaultAdminFrontendDistDir,
	}
	if appMode != "production" {
		cfg.DevServerURL = DefaultAdminFrontendDevServerURL
	}

	return cfg
}

func mountAdminFrontend(r *gin.Engine, appMode string) {
	mountAdminFrontendWithConfig(r, defaultAdminFrontendConfig(appMode))
}

func mountAdminFrontendWithConfig(r *gin.Engine, cfg AdminFrontendConfig) {
	if proxy := newAdminFrontendProxy(cfg.DevServerURL); proxy != nil {
		r.Any("/admin", gin.WrapH(proxy))
		r.Any("/admin/*path", gin.WrapH(proxy))
		return
	}

	if cfg.DistDir == "" {
		return
	}

	indexPath := filepath.Join(cfg.DistDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[WARN] 检查后台前端构建产物失败: %v", err)
		}
		return
	}

	handler := serveAdminFrontend(cfg.DistDir)
	r.GET("/admin", handler)
	r.GET("/admin/*path", handler)
}

func newAdminFrontendProxy(rawURL string) *httputil.ReverseProxy {
	if rawURL == "" {
		return nil
	}

	target, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("[WARN] 后台前端开发服务器地址无效: %q, err=%v", rawURL, err)
		return nil
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("[WARN] 后台前端开发代理失败: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
		_, _ = rw.Write([]byte("admin frontend dev server unavailable"))
	}

	return proxy
}

func serveAdminFrontend(distDir string) gin.HandlerFunc {
	indexPath := filepath.Join(distDir, "index.html")

	return func(c *gin.Context) {
		adminPath := strings.TrimPrefix(c.Request.URL.Path, "/admin")
		cleanPath := path.Clean("/" + adminPath)
		if cleanPath == "/" {
			c.File(indexPath)
			return
		}

		relativePath := strings.TrimPrefix(cleanPath, "/")
		assetPath := filepath.Join(distDir, filepath.FromSlash(relativePath))
		if fileExists(assetPath) {
			c.File(assetPath)
			return
		}

		if path.Ext(cleanPath) != "" {
			c.Status(http.StatusNotFound)
			return
		}

		c.File(indexPath)
	}
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
