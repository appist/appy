package handler

import (
	"net/http"

	"github.com/appist/appy/pack"
)

func welcomeIndex(c *pack.Context) {
	c.HTML(http.StatusOK, "welcome/index.html", nil)
}
