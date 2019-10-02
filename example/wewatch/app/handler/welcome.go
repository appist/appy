package handler

import (
	"net/http"

	"github.com/appist/appy"
)

// WelcomeIndex is the welcome index page.
func WelcomeIndex() appy.HandlerFuncT {
	return func(c *appy.ContextT) {
		c.JSON(http.StatusOK, appy.H{"a": 1})
	}
}
