package handler

import (
	"net/http"

	"github.com/appist/appy"
)

// WelcomeIndex is the welcome index page.
func WelcomeIndex() appy.HandlerFuncT {
	return func(c *appy.ContextT) {
		c.HTML(http.StatusOK, "welcome/index.html", appy.H{
			"message": appy.T(c, "message", map[string]interface{}{
				"Name":  "John Doe",
				"Count": 2,
			}),
		})
	}
}
