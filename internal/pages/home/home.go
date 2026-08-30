// Package home 渲染导航首页：读取数据、按关键词过滤、按分组拆表。
package home

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"nav/config/data"
	"nav/config/model"
	"nav/internal/fn"
)

// 页面完全由服务端渲染，不加载任何脚本。
const _cspValue = "script-src 'none'; object-src 'none'; base-uri 'none'"

const _maxKeywordLen = 50

// RowView 表格中的一行。
type RowView struct {
	Name    string
	Href    string
	Display string
	Desc    string
}

// Table 一张表格，Title 为空时不渲染分组标题。
type Table struct {
	Title  string
	Rows   []RowView
	NewTab bool
}

// PageData 模板上下文。
type PageData struct {
	Title      string
	Keyword    string
	HasKeyword bool
	ShowSearch bool
	Footer     string
	Tables     []Table
}

// entry 内部使用，把分组信息和展示行绑在一起，避免二次过滤。
type entry struct {
	group string
	row   RowView
}

func RegisterRouting(e *echo.Echo) {
	e.GET("/", pageHome)
	e.GET("/healthz", healthz)
}

func healthz(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func pageHome(c *echo.Context) error {
	d, err := data.Load()
	if err != nil {
		return c.String(http.StatusInternalServerError, "数据文件读取失败："+err.Error())
	}

	keyword := strings.TrimSpace(c.QueryParam("q"))
	if len([]rune(keyword)) > _maxKeywordLen {
		keyword = ""
	}
	keywordLower := strings.ToLower(keyword)

	tables := buildTables(d, keywordLower)
	shown := 0
	for _, t := range tables {
		shown += len(t.Rows)
	}

	c.Response().Header().Set("Content-Security-Policy", _cspValue)
	return c.Render(http.StatusOK, "index.html", PageData{
		Title:      d.Site.Title,
		Keyword:    keyword,
		HasKeyword: keyword != "",
		ShowSearch: d.Site.ShowSearch,
		Footer:     buildFooter(d, keyword, shown),
		Tables:     tables,
	})
}

// buildTables 过滤并组装表格。无分组时输出一张平铺表，
// 有分组时按 group 拆表，未标注 group 的条目归入第一个分组。
func buildTables(d model.Data, keyword string) []Table {
	matched := make([]entry, 0, len(d.Items))
	for _, site := range d.Items {
		if !fn.Match(keyword, site.Name, site.URL, site.Desc) {
			continue
		}
		matched = append(matched, entry{
			group: site.Group,
			row: RowView{
				Name:    site.Name,
				Href:    fn.NormalizeURL(fn.ParseDynamicURL(site.URL)),
				Display: fn.DisplayURL(site.URL),
				Desc:    site.Desc,
			},
		})
	}
	if len(matched) == 0 {
		return nil
	}

	if len(d.Groups) == 0 {
		return []Table{{Rows: rowsOf(matched), NewTab: d.Site.OpenLinkNewTab}}
	}

	defaultID := d.Groups[0].ID
	known := make(map[string]bool, len(d.Groups))
	for _, g := range d.Groups {
		known[g.ID] = true
	}

	tables := make([]Table, 0, len(d.Groups))
	var orphan []entry
	for _, g := range d.Groups {
		rows := make([]entry, 0)
		for _, m := range matched {
			id := m.group
			if id == "" {
				id = defaultID
			}
			if id == g.ID {
				rows = append(rows, m)
			}
		}
		if len(rows) == 0 {
			continue
		}
		tables = append(tables, Table{Title: g.Title, Rows: rowsOf(rows), NewTab: d.Site.OpenLinkNewTab})
	}

	// group 写了不存在的 ID 时兜底展示，避免条目凭空消失。
	for _, m := range matched {
		id := m.group
		if id == "" {
			id = defaultID
		}
		if !known[id] {
			orphan = append(orphan, m)
		}
	}
	if len(orphan) > 0 {
		tables = append(tables, Table{Title: "未分類", Rows: rowsOf(orphan), NewTab: d.Site.OpenLinkNewTab})
	}
	return tables
}

func rowsOf(in []entry) []RowView {
	rows := make([]RowView, 0, len(in))
	for _, e := range in {
		rows = append(rows, e.row)
	}
	return rows
}

func buildFooter(d model.Data, keyword string, shown int) string {
	if d.Site.Footer != "" {
		return d.Site.Footer
	}
	if keyword != "" {
		return fmt.Sprintf("搜尋「%s」：命中 %d 條 / 共 %d 條", keyword, shown, len(d.Items))
	}
	return fmt.Sprintf("共 %d 條連結", len(d.Items))
}
