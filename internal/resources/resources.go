// Package resources 注册模板渲染器与静态资源路由。
// 生产模式下资源来自编译期内嵌的二进制，调试模式下改为读磁盘以便热更新。
package resources

import (
	"bytes"
	"crypto/md5" //nolint:gosec // 仅用于生成资源 ETag，非安全用途
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v5"

	"nav/config/define"
	"nav/embed"
)

var bufPool = sync.Pool{New: func() any { return &bytes.Buffer{} }}

// newTemplate 构造一个空的模板集合作为解析起点。
func newTemplate() *template.Template {
	return template.New("")
}

// Renderer 实现 echo.Renderer，用缓冲池减少每次响应的内存分配。
type Renderer struct {
	templates *templateSet
}

type templateSet struct {
	names map[string]*template.Template
}

func (r *Renderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	t := r.templates.lookup(name)
	if t == nil {
		return fmt.Errorf("找不到模板: %s", name)
	}

	buf, ok := bufPool.Get().(*bytes.Buffer)
	if !ok || buf == nil {
		buf = &bytes.Buffer{}
	}
	buf.Reset()
	defer bufPool.Put(buf)

	if err := t.Execute(buf, data); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

// lookup 兼容 ParseGlob（键为 index.html）与 ParseFS（键为 templates/index.html）两种命名。
func (s *templateSet) lookup(name string) *template.Template {
	for _, cand := range []string{name, "templates/" + name, filepath.Base(name)} {
		if t, ok := s.names[cand]; ok {
			return t
		}
	}
	return nil
}

// Register 解析模板并挂到 echo 上。
func Register(e *echo.Echo) error {
	parsed, err := parseTemplates()
	if err != nil {
		return fmt.Errorf("初始化模板: %w", err)
	}
	e.Renderer = &Renderer{templates: parsed}
	return nil
}

func parseTemplates() (*templateSet, error) {
	set := &templateSet{names: map[string]*template.Template{}}

	// 调试模式：直接读磁盘，改完刷新即可看到效果。
	if define.AppFlags.Debug {
		files, err := filepath.Glob(filepath.Join("embed", "templates", "*.html"))
		if err != nil || len(files) == 0 {
			return nil, fmt.Errorf("调试模式下找不到 embed/templates/*.html: %w", err)
		}
		for _, f := range files {
			// ParseFiles 以文件名为模板名，返回的 t 即该文件本身。
			t, err := template.ParseFiles(f)
			if err != nil {
				return nil, err
			}
			set.names[filepath.Base(f)] = t
			set.names[f] = t
		}
		return set, nil
	}

	entries, err := fs.ReadDir(embedres.FS, "templates")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}
		raw, err := fs.ReadFile(embedres.FS, "templates/"+entry.Name())
		if err != nil {
			return nil, err
		}
		t, err := newTemplate().Parse(string(raw))
		if err != nil {
			return nil, err
		}
		set.names[entry.Name()] = t
	}
	return set, nil
}

// RegisterAssets 提供 /assets/ 下的静态文件，并设置长缓存与 ETag。
func RegisterAssets(e *echo.Echo) {
	etag := resourceETag()
	e.GET("/assets/*", func(c *echo.Context) error {
		req := c.Request()
		rel := strings.TrimPrefix(req.URL.Path, "/assets/")
		if rel == "" || strings.Contains(rel, "..") {
			return echo.NewHTTPError(http.StatusNotFound, "not found")
		}

		data, err := readAsset(rel)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "not found")
		}

		header := c.Response().Header()
		header.Set("Content-Type", contentTypeOf(rel))
		header.Set("Cache-Control", "public, max-age=31536000")
		header.Set("ETag", etag)
		if match := req.Header.Get("If-None-Match"); match != "" && strings.Contains(match, etag) {
			return c.NoContent(http.StatusNotModified)
		}
		return c.Blob(http.StatusOK, contentTypeOf(rel), data)
	})
}

func readAsset(rel string) ([]byte, error) {
	if define.AppFlags.Debug {
		if raw, err := os.ReadFile(filepath.Join("embed", "assets", filepath.FromSlash(rel))); err == nil {
			return raw, nil
		}
	}
	return fs.ReadFile(embedres.FS, "assets/"+filepath.ToSlash(rel))
}

func contentTypeOf(name string) string {
	switch filepath.Ext(name) {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

// resourceETag 进程启动时算一次即可：内嵌资源在运行期不会变。
func resourceETag() string {
	data := []byte(time.Now().String())
	return fmt.Sprintf(`W/"%x"`, md5.Sum(data)) //nolint:gosec // 见上
}
