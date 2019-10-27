package handler

import (
	"net/http"

	"github.com/appist/appy"
)

// WelcomeIndex is the welcome index page.
func WelcomeIndex() appy.HandlerFunc {
	return func(c *appy.Context) {
		appy.Db["primary"].Handler.Exec(`SELECT * FROM public.schema_migrations`)

		c.HTML(http.StatusOK, "welcome/index.html", appy.H{
			"message": appy.T(c, "message", appy.H{
				"Name":  "John",
				"Count": 2,
			}),
		})
	}
}
