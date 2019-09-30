package http

import (
	"appist/appy/support"
	"appist/appy/test"
	"bufio"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

type I18nSuiteT struct {
	test.SuiteT
	I18nBundle *i18n.Bundle
}

func (s *I18nSuiteT) SetupTest() {
	s.I18nBundle = i18n.NewBundle(language.English)
	s.I18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	s.I18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	s.I18nBundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	s.I18nBundle.LoadMessageFile("../testdata/locale/en.yml")
	s.I18nBundle.LoadMessageFile("../testdata/locale/zh-CN.yml")
	s.I18nBundle.LoadMessageFile("../testdata/locale/zh-TW.yml")
}

func (s *I18nSuiteT) TestTranslate() {
	data := &viewDataT{}

	localizer := i18n.NewLocalizer(s.I18nBundle, "en")
	data.setLocalizer(localizer)
	s.Equal("Password", data.T("password"))
	s.Equal("test has no message.", data.T("messages", map[string]interface{}{"Name": "test", "Count": 0}))
	s.Equal("test has 1 message.", data.T("messages", map[string]interface{}{"Name": "test", "Count": 1}))
	s.Equal("test has 10 messages.", data.T("messages", map[string]interface{}{"Name": "test", "Count": 10}))

	localizer = i18n.NewLocalizer(s.I18nBundle, "zh-cn")
	data.setLocalizer(localizer)
	s.Equal("密码", data.T("password"))
	s.Equal("test没有讯息。", data.T("messages", map[string]interface{}{"Name": "test", "Count": 0}))
	s.Equal("test有1则讯息。", data.T("messages", map[string]interface{}{"Name": "test", "Count": 1}))
	s.Equal("test有10则讯息。", data.T("messages", map[string]interface{}{"Name": "test", "Count": 10}))

	localizer = i18n.NewLocalizer(s.I18nBundle, "zh-tw")
	data.setLocalizer(localizer)
	s.Equal("密碼", data.T("password"))
	s.Equal("test沒有訊息。", data.T("messages", map[string]interface{}{"Name": "test", "Count": 0}))
	s.Equal("test有1則訊息。", data.T("messages", map[string]interface{}{"Name": "test", "Count": 1}))
	s.Equal("test有10則訊息。", data.T("messages", map[string]interface{}{"Name": "test", "Count": 10}))
}

func TestI18n(t *testing.T) {
	test.Run(t, new(I18nSuiteT))
}

type RenderSuiteT struct {
	test.SuiteT
	Server *ServerT
}

func (s *RenderSuiteT) SetupTest() {
	config := support.Config
	config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.Server = NewServer(config)
}

func (s *RenderSuiteT) TestRenderHTML() {
	s.Server.htmlRenderer.AddFromString("test", "<div>test</div>")
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderHTML(c, http.StatusOK, "test", nil)
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("<div>test</div>", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderASCIIJSON() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderASCIIJSON(c, http.StatusOK, gin.H{
			"lang": "GO语言",
			"tag":  "<br>",
		})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/json", recorder.Header().Get("Content-Type"))
	s.Equal(`{"lang":"GO\u8bed\u8a00","tag":"\u003cbr\u003e"}`, recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderData() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderData(c, http.StatusOK, "text/html", []byte("<div>test</div>"))
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("text/html", recorder.Header().Get("Content-Type"))
	s.Equal("<div>test</div>", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderDataFromReader() {
	f, _ := os.Open("../testdata/test.html")
	fi, _ := f.Stat()
	reader := bufio.NewReader(f)
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderDataFromReader(c, http.StatusOK, fi.Size(), "text/html", reader, map[string]string{})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("text/html", recorder.Header().Get("Content-Type"))
	s.Equal("i am a test\n", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderFile() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderFile(c, "../testdata/test.html")
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("i am a test\n", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderFileAttachment() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderFileAttachment(c, "../testdata", "test.html")
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(301, recorder.Code)
	s.Equal(`attachment; filename="test.html"`, recorder.Header().Get("content-disposition"))
}

func (s *RenderSuiteT) TestRenderIndentedJSON() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderIndentedJSON(c, http.StatusOK, gin.H{"a": 1})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("{\n    \"a\": 1\n}", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderJSON() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderJSON(c, http.StatusOK, gin.H{"a": 1})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("{\"a\":1}", strings.Trim(recorder.Body.String(), "\n"))
}

func (s *RenderSuiteT) TestRenderJSONP() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderJSONP(c, http.StatusOK, gin.H{"a": 1})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render?callback=x", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/javascript; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("x({\"a\":1});", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderPureJSON() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderPureJSON(c, http.StatusOK, gin.H{"html": "<div>test</div>"})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/json", recorder.Header().Get("Content-Type"))
	s.Equal(`{"html":"\u003cdiv\u003etest\u003c/div\u003e"}`, recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderSecureJSON() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderSecureJSON(c, http.StatusOK, []string{"lena", "austin", "foo"})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("while(1);[\"lena\",\"austin\",\"foo\"]", recorder.Body.String())

	s.Server.SecureJSONPrefix(";;")
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal(";;[\"lena\",\"austin\",\"foo\"]", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderSSE() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderSSEvent(c, "sse", gin.H{"state": true})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("text/event-stream", recorder.Header().Get("Content-Type"))
	s.Equal("event:sse\ndata:{\"state\":true}\n\n", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderXML() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderXML(c, http.StatusOK, gin.H{"user": "john"})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/xml; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("<map><user>john</user></map>", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderYAML() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderYAML(c, http.StatusOK, gin.H{"user": "john"})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("application/x-yaml; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("user: john\n", recorder.Body.String())
}

func (s *RenderSuiteT) TestRenderString() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.RenderString(c, http.StatusOK, "test")
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("test", recorder.Body.String())
}

func (s *RenderSuiteT) TestRedirect() {
	s.Server.GET("/render", func(c *gin.Context) {
		s.Server.Redirect(c, http.StatusTemporaryRedirect, "/paradise")
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(307, recorder.Code)
}

func TestRender(t *testing.T) {
	test.Run(t, new(RenderSuiteT))
}

type BeforeRenderSuiteT struct {
	test.SuiteT
	Server *ServerT
}

func (s *BeforeRenderSuiteT) SetupTest() {
	s.Server = NewServer(support.Config)
}

func (s *BeforeRenderSuiteT) TestRemoveSetCookieIfAPIOnly() {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = &http.Request{
		Header: http.Header{},
	}
	c.Writer.Header().Set("Set-Cookie", "session=1")
	s.Server.beforeRender(c)
	s.Equal("session=1", c.Writer.Header().Get("Set-Cookie"))

	c.Request.Header.Set("x-api-only", "1")
	s.Server.beforeRender(c)
	s.Equal("", c.Writer.Header().Get("Set-Cookie"))
}

func (s *BeforeRenderSuiteT) TestCSRFValidation() {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = &http.Request{
		Header: http.Header{},
	}
	c.Set("appy.csrfError", errors.New("the CSRF token is invalid"))
	s.Server.beforeRender(c)
	s.Equal(true, c.IsAborted())
	s.Equal(403, c.Writer.Status())
}

func TestBeforeRender(t *testing.T) {
	test.Run(t, new(BeforeRenderSuiteT))
}
