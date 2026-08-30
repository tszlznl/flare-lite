package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"nav/config/data"
	"nav/config/define"
	"nav/config/model"
)

// newTestHandler 把进程切到临时目录，让首次启动自动生成示例 sites.yml。
func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取工作目录失败: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("切换临时目录失败: %v", err)
	}
	t.Cleanup(func() {
		data.ResetCache()
		_ = os.Chdir(origWd)
	})

	data.ResetCache()
	define.AppFlags = model.Flags{Config: "sites.yml"}

	handler, err := NewRouter()
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return handler
}

// newTestHandlerWithYAML 用给定的数据文件启动，用于测试分组等非默认形态。
func newTestHandlerWithYAML(t *testing.T, yaml string) http.Handler {
	t.Helper()
	h := newTestHandler(t)
	if err := os.WriteFile("sites.yml", []byte(yaml), 0o644); err != nil {
		t.Fatalf("写入测试数据失败: %v", err)
	}
	return h
}

func get(t *testing.T, h http.Handler, path string) (int, string, http.Header) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code, rec.Body.String(), rec.Header()
}

func TestIndexRendersReferenceTable(t *testing.T) {
	code, body, header := get(t, newTestHandler(t), "/")

	if code != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 200", code)
	}

	for _, want := range []string{
		"<th>名稱</th>", "<th>URL</th>", "<th>第一印象</th>",
		"The Lunduke Journal", "ManateeLazyCat", "聆音播放室",
		"共 16 條連結",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("页面缺少 %q", want)
		}
	}

	// 缺少协议的条目应被补全后才写进 href。
	if !strings.Contains(body, `href="https://www.zaqizaba.xyz"`) {
		t.Error("zaqizaba 的 href 未补全协议")
	}
	if !strings.Contains(body, `>www.zaqizaba.xyz<`) {
		t.Error("zaqizaba 的展示文本应保持原样")
	}

	// 页面设计为零脚本，CSP 必须落地。
	if strings.Contains(body, "<script") {
		t.Error("页面不应包含 <script>")
	}
	csp := header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'none'") {
		t.Errorf("CSP 缺少 script-src 'none': %q", csp)
	}
}

func TestSearchFiltersRows(t *testing.T) {
	h := newTestHandler(t)

	code, body, _ := get(t, h, "/?q=deepin")
	if code != http.StatusOK {
		t.Fatalf("状态码 = %d", code)
	}
	if !strings.Contains(body, "ManateeLazyCat") {
		t.Error("搜索 deepin 应命中 ManateeLazyCat")
	}
	if strings.Contains(body, "The Lunduke Journal") {
		t.Error("搜索 deepin 不应命中无关条目")
	}
	if !strings.Contains(body, "命中 1 條") {
		t.Error("页脚应显示命中数量")
	}

	_, body, _ = get(t, h, "/?q=不存在的站点xyz")
	if strings.Contains(body, "nav-table") {
		t.Error("无命中时不应渲染表格")
	}
	if !strings.Contains(body, "沒有符合條件的連結") {
		t.Error("无命中时应显示空状态")
	}
}

func TestUnknownRouteIs404(t *testing.T) {
	code, _, _ := get(t, newTestHandler(t), "/nope")
	if code != http.StatusNotFound {
		t.Errorf("未知路径状态码 = %d, 期望 404", code)
	}
}

func TestAssetsServedAndTraversalBlocked(t *testing.T) {
	h := newTestHandler(t)

	code, body, header := get(t, h, "/assets/css/style.css")
	if code != http.StatusOK {
		t.Fatalf("样式表状态码 = %d", code)
	}
	if !strings.Contains(body, ".nav-table") {
		t.Error("样式表内容不完整")
	}
	if ct := header.Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("Content-Type = %q", ct)
	}

	if code, _, _ := get(t, h, "/assets/../../go.mod"); code == http.StatusOK {
		t.Errorf("目录穿越请求未被拦截, 状态码 = %d", code)
	}
}

func TestHealthz(t *testing.T) {
	code, body, _ := get(t, newTestHandler(t), "/healthz")
	if code != http.StatusOK || body != "OK" {
		t.Errorf("healthz = %d/%q", code, body)
	}
}

// TestGroupsSplitIntoMultipleTables 验证声明 groups 后按分组拆表，
// 未标注 group 的条目落到第一个分组。
func TestGroupsSplitIntoMultipleTables(t *testing.T) {
	const data = `
site:
  title: 分组测试
  show_search: false
groups:
  - id: tech
    title: 技術
  - id: life
    title: 生活
links:
  - name: 没标分组
    link: https://a.com
  - name: 技术站
    link: https://b.com
    group: tech
  - name: 生活站
    link: https://c.com
    group: life
  - name: 野站
    link: https://d.com
    group: not-exist
`
	code, body, _ := get(t, newTestHandlerWithYAML(t, data), "/")
	if code != http.StatusOK {
		t.Fatalf("状态码 = %d", code)
	}

	for _, want := range []string{"技術", "生活", "未分類"} {
		if !strings.Contains(body, want) {
			t.Errorf("缺少分组标题 %q", want)
		}
	}
	// tech（含未标注分组的条目）+ life + 未分類 兜底
	if got := strings.Count(body, `<table class="nav-table">`); got != 3 {
		t.Errorf("表格数量 = %d, 期望 3", got)
	}
	// 归入第一个分组的条目应与 tech 同表
	if strings.Count(body, "没标分组") != 1 {
		t.Error("未标注 group 的条目应恰好出现一次")
	}
	// show_search: false 时不渲染搜索框
	if strings.Contains(body, `<form class="search"`) {
		t.Error("show_search=false 时不应渲染搜索框")
	}
	// 自定义 footer 生效前，标题应正常显示
	if !strings.Contains(body, "<title>分组测试</title>") {
		t.Error("标题未生效")
	}
}

// TestHotReloadWithoutRestart 验证改完 sites.yml 下一次请求即生效。
func TestHotReloadWithoutRestart(t *testing.T) {
	h := newTestHandler(t)

	_, body, _ := get(t, h, "/")
	if !strings.Contains(body, "The Lunduke Journal") {
		t.Fatal("初始数据未加载")
	}

	const replaced = `
site:
  title: 改过了
links:
  - name: 唯一一条
    link: https://only.example.com
`
	// 内容长度和 mtime 都变了，缓存必须失效
	if err := os.WriteFile("sites.yml", []byte(replaced), 0o644); err != nil {
		t.Fatalf("改写数据文件失败: %v", err)
	}

	_, body, _ = get(t, h, "/")
	if !strings.Contains(body, "唯一一条") {
		t.Error("改写后未热更新")
	}
	if strings.Contains(body, "The Lunduke Journal") {
		t.Error("仍在使用旧缓存")
	}
	if !strings.Contains(body, "共 1 條連結") {
		t.Errorf("页脚未随之更新:\n%s", body)
	}
}

// TestBrokenYAMLFailsGracefully 数据文件写坏时应返回错误而不是 panic。
func TestBrokenYAMLFailsGracefully(t *testing.T) {
	h := newTestHandlerWithYAML(t, "site:\n  title: [broken\n   bad indent\n")

	code, body, _ := get(t, h, "/")
	if code != http.StatusInternalServerError {
		t.Errorf("状态码 = %d, 期望 500", code)
	}
	if !strings.Contains(body, "解析") {
		t.Errorf("错误信息应说明解析失败: %q", body)
	}
}
