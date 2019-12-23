package support

import "github.com/CloudyKit/jet"

import "html/template"

type (
	// ViewEngine renders the view template.
	ViewEngine struct {
		*jet.Set
	}
)

// NewViewEngine initializes the view engine instance.
func NewViewEngine(assets *Assets) *ViewEngine {
	return &ViewEngine{
		jet.NewSetLoader(template.HTMLEscape, NewViewLoader(assets)),
	}
}
