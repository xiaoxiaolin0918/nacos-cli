package cmd

import (
	"strings"
	"testing"
)

func TestWebLangFromAcceptLanguage(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"zh-CN,zh;q=0.9,en;q=0.8", "zh"},
		{"zh-TW,zh;q=0.8", "zh"},
		{"zh", "zh"},
		{"en-US,en;q=0.9", "en"},
		{"en", "en"},
		{"", "en"},
		{"fr-FR,fr;q=0.9", "en"},
		{"en-US,zh-CN;q=0.5", "zh"},
	}
	for _, c := range cases {
		if got := webLang(c.in); got != c.want {
			t.Errorf("webLang(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestFormHTMLChineseLightCard(t *testing.T) {
	h := formHTML("http://nacos.example:8848", "nacos", "", "zh", false)
	for _, want := range []string{
		"登录", "用户名", "密码", "目标",
		"密码仅留在本表单",
		"card", "#1677ff",
		"nacos.example:8848", `value="nacos"`,
		`lang="zh"`,
	} {
		if !strings.Contains(h, want) {
			t.Errorf("zh form missing %q", want)
		}
	}
	for _, no := range []string{"Username", "Password", "Sign in", "Target:"} {
		if strings.Contains(h, no) {
			t.Errorf("zh form should not show English %q", no)
		}
	}
}

func TestFormHTMLEnglishLightCard(t *testing.T) {
	h := formHTML("http://nacos.example:8848", "nacos", "", "en", false)
	for _, want := range []string{
		"Login", "Username", "Password", "Target",
		"Password stays in this form", "Sign in",
		"card", "#1677ff",
		`lang="en"`,
	} {
		if !strings.Contains(h, want) {
			t.Errorf("en form missing %q", want)
		}
	}
	for _, no := range []string{"用户名", "密码仅留在本表单", "目标"} {
		if strings.Contains(h, no) {
			t.Errorf("en form should not show Chinese %q", no)
		}
	}
}

func TestFormHTMLEscapesAndShowsError(t *testing.T) {
	h := formHTML("http://x\"><script>", "u\"><x", "err<script>", "en", false)
	if strings.Contains(h, "<script>") {
		t.Fatal("unescaped script in formHTML")
	}
	if !strings.Contains(h, "&lt;script&gt;") {
		t.Fatal("expected escaped error/username/server")
	}
	if !strings.Contains(h, "#b00020") {
		t.Fatal("expected red error styling")
	}
}

func TestOkHTMLChineseAndEnglish(t *testing.T) {
	zh := okHTML("http://nacos.example", "token", "zh", false, nil)
	for _, want := range []string{"登录已保存", "服务器", "认证模式", "可关闭此标签页", "card", "token"} {
		if !strings.Contains(zh, want) {
			t.Errorf("zh ok missing %q", want)
		}
	}
	if strings.Contains(zh, "Login saved") {
		t.Fatal("zh ok should not show English title")
	}

	en := okHTML("http://nacos.example", "token", "en", false, nil)
	for _, want := range []string{"Login saved", "Server", "Auth mode", "close this tab", "card", "#1677ff"} {
		if !strings.Contains(en, want) {
			t.Errorf("en ok missing %q", want)
		}
	}
	if strings.Contains(en, "登录已保存") {
		t.Fatal("en ok should not show Chinese title")
	}
}

func TestFailHTMLChineseAndEnglish(t *testing.T) {
	zh := failHTML("bad<script>", "zh")
	for _, want := range []string{"登录失败", "card", "#b00020", "&lt;script&gt;"} {
		if !strings.Contains(zh, want) {
			t.Errorf("zh fail missing %q", want)
		}
	}
	if strings.Contains(zh, "<script>") {
		t.Fatal("unescaped script in failHTML")
	}
	if strings.Contains(zh, "Login failed") {
		t.Fatal("zh fail should not show English title")
	}

	en := failHTML("nope", "en")
	for _, want := range []string{"Login failed", "card", "#b00020", "nope"} {
		if !strings.Contains(en, want) {
			t.Errorf("en fail missing %q", want)
		}
	}
	if strings.Contains(en, "登录失败") {
		t.Fatal("en fail should not show Chinese title")
	}
}

func TestOkHTMLEscapes(t *testing.T) {
	h := okHTML("s\"><x", "m\"><y", "en", false, nil)
	if strings.Contains(h, "s\"><x") || strings.Contains(h, "m\"><y") {
		t.Fatal("raw injection payload still present in okHTML")
	}
	if !strings.Contains(h, "s&#34;&gt;&lt;x") {
		t.Fatal("expected escaped server value in okHTML")
	}
	if !strings.Contains(h, "m&#34;&gt;&lt;y") {
		t.Fatal("expected escaped auth mode in okHTML")
	}
}