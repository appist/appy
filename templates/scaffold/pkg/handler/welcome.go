package handler

import (
	"net/http"

	"github.com/appist/appy"
)

func welcomeIndex(c *appy.Context) {
	c.HTML(http.StatusOK, "welcome/index.html", nil)
}
