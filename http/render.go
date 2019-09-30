package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/appist/appy/middleware"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type viewDataT struct {
	localizer *i18n.Localizer
	Data      interface{}
}

func (d *viewDataT) setLocalizer(localizer *i18n.Localizer) {
	d.localizer = localizer
}

// T provides the translate functionality.
func (d *viewDataT) T(key string, args ...map[string]interface{}) string {
	if len(args) < 1 {
		return d.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: key})
	}

	count := -1
	if _, ok := args[0]["Count"]; ok {
		count = args[0]["Count"].(int)
	}

	countKey := key
	if count != -1 {
		switch count {
		case 0:
			countKey = key + ".Zero"
		case 1:
			countKey = key + ".One"
		default:
			countKey = key + ".Other"
		}
	}

	return d.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: countKey, TemplateData: args[0]})
}

// RenderHTML renders the HTTP template specified by its file name. It also updates the HTTP code and sets the
// Content-Type as "text/html". See http://golang.org/doc/articles/wiki/.
func (s *ServerT) RenderHTML(c *gin.Context, code int, name string, obj interface{}) {
	s.beforeRender(c)

	d := &viewDataT{Data: obj}
	d.setLocalizer(middleware.I18nLocalizer(c))

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.HTML(code, name, d)
	}

	s.afterRender(c)
}

// RenderASCIIJSON serializes the given struct as JSON into the response body with unicode to ASCII string. It also
// sets the Content-Type as "application/json".
func (s *ServerT) RenderASCIIJSON(c *gin.Context, code int, obj interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.AsciiJSON(code, obj)
	}

	s.afterRender(c)
}

// RenderData writes some data into the body stream and updates the HTTP code.
func (s *ServerT) RenderData(c *gin.Context, code int, contentType string, data []byte) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.Data(code, contentType, data)
	}

	s.afterRender(c)
}

// RenderDataFromReader writes the specified reader into the body stream and updates the HTTP code.
func (s *ServerT) RenderDataFromReader(c *gin.Context, code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.DataFromReader(code, contentLength, contentType, reader, extraHeaders)
	}

	s.afterRender(c)
}

// RenderFile writes the specified file into the body stream in a efficient way.
func (s *ServerT) RenderFile(c *gin.Context, filepath string) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		http.ServeFile(c.Writer, c.Request, filepath)
	}

	s.afterRender(c)
}

// RenderFileAttachment writes the specified file into the body stream in an efficient way. On the client side, the
// file will typically be downloaded with the given filename.
func (s *ServerT) RenderFileAttachment(c *gin.Context, filepath, filename string) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		http.ServeFile(c.Writer, c.Request, filepath)
	}

	s.afterRender(c)
}

// RenderIndentedJSON serializes the given struct as pretty JSON (indented + endlines) into the response body.
// It also sets the Content-Type as "application/json".
func (s *ServerT) RenderIndentedJSON(c *gin.Context, code int, obj interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.IndentedJSON(code, obj)
	}

	s.afterRender(c)
}

// RenderJSON serializes the given struct as JSON into the response body. It also sets the Content-Type as
// "application/json".
func (s *ServerT) RenderJSON(c *gin.Context, code int, obj interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.JSON(code, obj)
	}

	s.afterRender(c)
}

// RenderJSONP serializes the given struct as JSON into the response body. It add padding to response body to request
// data from a server residing in a different domain than the client. It also sets the Content-Type as
// "application/javascript".
func (s *ServerT) RenderJSONP(c *gin.Context, code int, obj interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.JSONP(code, obj)
	}

	s.afterRender(c)
}

// RenderPureJSON serializes the given struct as JSON into the response body.
// RenderPureJSON, unlike RenderJSON, does not replace special html characters with their unicode entities.
func (s *ServerT) RenderPureJSON(c *gin.Context, code int, obj interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.AsciiJSON(code, obj)
	}

	s.afterRender(c)
}

// RenderSecureJSON serializes the given struct as Secure JSON into the response body. Default prepends "while(1);" to
// response body if the given struct is array values. It also sets the Content-Type as "application/json".
func (s *ServerT) RenderSecureJSON(c *gin.Context, code int, obj interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.SecureJSON(code, obj)
	}

	s.afterRender(c)
}

// RenderSSEvent writes a Server-Sent Event into the body stream.
func (s *ServerT) RenderSSEvent(c *gin.Context, name string, message interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.Render(-1, sse.Event{
			Event: name,
			Data:  message,
		})
	}

	s.afterRender(c)
}

// RenderXML serializes the given struct as XML into the response body. It also sets the Content-Type as
// "application/xml".
func (s *ServerT) RenderXML(c *gin.Context, code int, obj interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.XML(code, obj)
	}

	s.afterRender(c)
}

// RenderYAML serializes the given struct as YAML into the response body. It also sets the Content-Type as
// "application/x-yaml".
func (s *ServerT) RenderYAML(c *gin.Context, code int, obj interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.YAML(code, obj)
	}

	s.afterRender(c)
}

// RenderString serializes the given struct as String into the response body. It also sets the Content-Type as
// "text/plain".
func (s *ServerT) RenderString(c *gin.Context, code int, format string, values ...interface{}) {
	s.beforeRender(c)

	if c.IsAborted() == false && len(c.Errors) == 0 {
		c.String(code, format, values...)
	}

	s.afterRender(c)
}

// Redirect returns a HTTP redirect to the specific location.
func (s *ServerT) Redirect(c *gin.Context, code int, location string) {
	s.beforeRender(c)

	if c.IsAborted() == false {
		c.Redirect(code, location)
	}

	s.afterRender(c)
}

func (s *ServerT) beforeRender(c *gin.Context) {
	if middleware.IsAPIOnly(c) == true {
		c.Writer.Header().Del("Set-Cookie")
	}

	err := middleware.CSRFError(c)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, err)
	}
}

func (s *ServerT) afterRender(c *gin.Context) {
}
