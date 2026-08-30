package data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nav/config/define"
	"nav/config/model"
)

func TestSeedFallbackOnPermissionDenied(t *testing.T) {
	// 使用一个不可写入的假路径（例如一个指向不存在目录且父目录是只读文件的路径）
	tmpDir := t.TempDir()
	readOnlyFile := filepath.Join(tmpDir, "dummy_file")
	if err := os.WriteFile(readOnlyFile, []byte("content"), 0o444); err != nil {
		t.Fatalf("创建只读文件失败: %v", err)
	}

	// 将 config 设为只读文件下的子文件路径（在任何系统上向文件下创建文件都会失败）
	invalidPath := filepath.Join(readOnlyFile, "sites.yml")
	define.AppFlags = model.Flags{Config: invalidPath}

	// 重置缓存状态
	mu.Lock()
	loaded = false
	cacheMod = nil
	mu.Unlock()

	d, err := Load()
	if err != nil {
		t.Fatalf("期望优雅降级至内置数据，但返回了错误: %v", err)
	}

	if d.Site.Title != "連結收藏" {
		t.Errorf("期望默认标题 %q, 得到 %q", "連結收藏", d.Site.Title)
	}
	if len(d.Items) == 0 {
		t.Errorf("期望降级加载内置示例条目，但为空")
	}
}

func TestLoadValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sites.yml")
	content := `
site:
  title: 自定义测试站点
links:
  - name: 测试链接
    link: https://example.com
`
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}

	define.AppFlags = model.Flags{Config: filePath}

	mu.Lock()
	loaded = false
	cacheMod = nil
	mu.Unlock()

	d, err := Load()
	if err != nil {
		t.Fatalf("读取有效数据文件失败: %v", err)
	}

	if d.Site.Title != "自定义测试站点" {
		t.Errorf("期望标题 %q, 得到 %q", "自定义测试站点", d.Site.Title)
	}
	if len(d.Items) != 1 || !strings.Contains(d.Items[0].Name, "测试链接") {
		t.Errorf("读取条目异常: %+v", d.Items)
	}
}
