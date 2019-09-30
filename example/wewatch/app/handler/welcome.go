package handler

import (
	"net/http"

	"github.com/appist/appy"
)

func WelcomeIndex() appy.HandlerFuncT {
	return func(c *appy.ContextT) {
		c.JSON(http.StatusOK, appy.H{"a": 1})
	}
}
