package pack

import (
	"github.com/appist/appy/support"
	"github.com/gin-gonic/gin"
)

type (
	// H is a shortcut for map[string]interface{}.
	H = support.H
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}
