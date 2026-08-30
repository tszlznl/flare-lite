package fn

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"  https://a.com  ", "https://a.com"},
		{"www.zaqizaba.xyz", "https://www.zaqizaba.xyz"},
		{"lunduke.com/", "https://lunduke.com/"},
		{"http://a.com", "http://a.com"},
		{"mailto:me@example.com", "mailto:me@example.com"},
		// 国际化域名必须转成 punycode，百分号编码的 host 不合法。
		{"https://琳.tw", "https://xn--jgy.tw"},
		{"琳.tw", "https://xn--jgy.tw"},
	}
	for _, c := range cases {
		if got := NormalizeURL(c.in); got != c.want {
			t.Errorf("NormalizeURL(%q) = %q, 期望 %q", c.in, got, c.want)
		}
	}
}

func TestDisplayURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"https://lunduke.com/", "lunduke.com"},
		{"http://a.com/x", "a.com/x"},
		{"www.zaqizaba.xyz", "www.zaqizaba.xyz"},
	}
	for _, c := range cases {
		if got := DisplayURL(c.in); got != c.want {
			t.Errorf("DisplayURL(%q) = %q, 期望 %q", c.in, got, c.want)
		}
	}
}

func TestParseDynamicURL(t *testing.T) {
	if got := ParseDynamicURL("https://a.com"); got != "https://a.com" {
		t.Errorf("不含占位符时不应改动: %q", got)
	}
	now := time.Now()
	got := ParseDynamicURL("https://a.com/{year}/{compact-date}")
	want := "https://a.com/" + now.Format("2006") + "/" + now.Format("20060102")
	if got != want {
		t.Errorf("ParseDynamicURL = %q, 期望 %q", got, want)
	}
}

func TestHostOf(t *testing.T) {
	if got := HostOf("https://www.zaqizaba.xyz/a?b=1"); got != "www.zaqizaba.xyz" {
		t.Errorf("HostOf = %q", got)
	}
}

func TestMatch(t *testing.T) {
	if !Match("", "a", "b", "c") {
		t.Error("空关键词应命中全部")
	}
	// 名称、链接、简介各命中一次，且大小写不敏感。
	if !Match("lun", "The Lunduke Journal", "", "") {
		t.Error("应命中名称")
	}
	if !Match("github", "", "https://x.github.io", "") {
		t.Error("应命中链接")
	}
	if !Match("运维", "", "", "Linux 企业运维知识。") {
		t.Error("应命中简介")
	}
	if Match("zzz", "a", "b", "c") {
		t.Error("无关关键词不应命中")
	}
}

func TestNormalizeURLKeepsSchemelessPath(t *testing.T) {
	got := NormalizeURL("example.com/a b")
	if !strings.HasPrefix(got, "https://") {
		t.Errorf("应补全协议: %q", got)
	}
}

// TestPunycodeEncode 校验 RFC 3492 编码实现。
// 期望值取自 .NET System.Globalization.IdnMapping（Windows 内置的权威 IDNA 实现）。
func TestPunycodeEncode(t *testing.T) {
	cases := []struct{ in, want string }{
		{"bücher", "bcher-kva"},
		{"münchen", "mnchen-3ya"},
		{"日本語", "wgv71a119e"},
		{"北京", "1lq90i"},
		{"中文", "fiq228c"},
		{"琳", "jgy"},
		{"", ""},
	}
	for _, c := range cases {
		if got := punycodeEncode(c.in); got != c.want {
			t.Errorf("punycodeEncode(%q) = %q, 期望 %q", c.in, got, c.want)
		}
	}
}

func TestAsciiDomain(t *testing.T) {
	cases := []struct{ in, want string }{
		{"www.zaqizaba.xyz", "www.zaqizaba.xyz"},
		{"lunduke.com", "lunduke.com"},
		{"münchen.de", "xn--mnchen-3ya.de"},
		{"北京.cn", "xn--1lq90i.cn"},
		{"琳.tw", "xn--jgy.tw"},
		// 已编码的标签不重复处理
		{"xn--1lq90i.cn", "xn--1lq90i.cn"},
	}
	for _, c := range cases {
		if got := asciiDomain(c.in); got != c.want {
			t.Errorf("asciiDomain(%q) = %q, 期望 %q", c.in, got, c.want)
		}
	}
}
