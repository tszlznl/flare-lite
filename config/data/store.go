// Package data 负责读写书签数据。没有数据库，一个 YAML 文件就是全部存储。
package data

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v2"

	"nav/config/define"
	"nav/config/model"
)

var (
	mu       sync.Mutex
	cache    model.Data
	cacheMod os.FileInfo
	loaded   bool
)

// Load 读取并解析数据文件。文件不存在时自动写入一份示例数据，
// 通过对比 mtime + size 做进程内缓存，避免每次请求都读盘。
func Load() (model.Data, error) {
	mu.Lock()
	defer mu.Unlock()

	path := define.ConfigPath()
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return seed(path)
		}
		return cache, fmt.Errorf("检查数据文件 %s 失败: %w", path, err)
	}

	if loaded && cacheMod != nil && info.ModTime().Equal(cacheMod.ModTime()) && info.Size() == cacheMod.Size() {
		return cache, nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return cache, fmt.Errorf("读取数据文件 %s 失败: %w", path, err)
	}

	var result model.Data
	if err := yaml.Unmarshal(raw, &result); err != nil {
		return cache, fmt.Errorf("解析数据文件 %s 失败，请检查 YAML 格式: %w", path, err)
	}

	cache, cacheMod, loaded = result, info, true
	applyDefaults(&cache)
	return cache, nil
}

// applyDefaults 补齐缺省的展示配置，让 sites.yml 可以只写必要字段。
// footer 故意留空，由页面根据是否处于搜索结果动态生成。
func applyDefaults(d *model.Data) {
	if d.Site.Title == "" {
		d.Site.Title = "連結收藏"
	}
}

// Save 覆盖写回数据文件，并让缓存立即失效。
func Save(d model.Data) error {
	out, err := yaml.Marshal(d)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}
	if err := os.WriteFile(define.ConfigPath(), out, 0o644); err != nil {
		return fmt.Errorf("写入数据文件失败: %w", err)
	}
	mu.Lock()
	loaded = false
	cacheMod = nil
	mu.Unlock()
	return nil
}

// seed 在首次运行时生成示例数据，随后照常返回。
func seed(path string) (model.Data, error) {
	if err := os.WriteFile(path, []byte(_exampleYAML), 0o644); err != nil {
		return cache, fmt.Errorf("创建数据文件 %s 失败: %w", path, err)
	}
	log.Printf("未找到数据文件，已生成示例数据：%s", path)

	var result model.Data
	if err := yaml.Unmarshal([]byte(_exampleYAML), &result); err != nil {
		return cache, fmt.Errorf("解析示例数据失败: %w", err)
	}
	info, statErr := os.Stat(path)
	if statErr == nil {
		cache, cacheMod, loaded = result, info, true
	} else {
		cache, loaded = result, false
	}
	applyDefaults(&cache)
	return cache, nil
}
