// Package fn 存放无副作用的小工具。
package fn

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// NormalizeURL 补全缺少协议的链接，让 www.example.com 也能直接点。
// 无法识别的输入原样返回，交给浏览器处理。
func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") && !strings.HasPrefix(raw, "mailto:") {
		raw = "https://" + raw
	}
	return withASCIIHost(raw)
}

// withASCIIHost 把国际化域名换成 punycode。
// html/template 会把 href 里的非 ASCII 主机名百分号编码，而百分号编码的
// host 并不合法，因此这里提前转成 xn-- 形式，保证链接在哪都能打开。
func withASCIIHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return raw
	}
	host := parsed.Hostname()
	asciiHost := asciiDomain(host)
	if asciiHost == host {
		return raw
	}
	return strings.Replace(raw, host, asciiHost, 1)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

// DisplayURL 去掉协议和末尾斜杠，让 URL 列不至于被 https:// 占满。
func DisplayURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if i := strings.Index(raw, "://"); i >= 0 {
		raw = raw[i+3:]
	}
	return strings.TrimSuffix(raw, "/")
}

// HostOf 取出链接的域名，用于生成占位图标。
func HostOf(raw string) string {
	parsed, err := url.Parse(NormalizeURL(raw))
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

// ParseDynamicURL 支持在链接里写 {date} {year} {month} {day} 等占位符，
// 渲染时替换为当天日期，方便收藏「今天的日报 / 周刊」这类每日一变的地址。
func ParseDynamicURL(raw string) string {
	if !strings.Contains(raw, "{") {
		return raw
	}
	now := time.Now()
	replacer := strings.NewReplacer(
		"{year}", strconv.Itoa(now.Year()),
		"{month}", strconv.Itoa(int(now.Month())),
		"{day}", strconv.Itoa(now.Day()),
		"{date}", now.Format("2006-01-02"),
		"{compact-date}", now.Format("20060102"),
	)
	return replacer.Replace(raw)
}

// Match 判断关键词是否命中名称、链接或简介中的任意一处。
func Match(keyword string, name, link, desc string) bool {
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(name), keyword) ||
		strings.Contains(strings.ToLower(link), keyword) ||
		strings.Contains(strings.ToLower(desc), keyword)
}
