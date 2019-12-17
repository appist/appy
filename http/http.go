package http

type (
	// ContextKey is the context key with appy namespace.
	ContextKey string
)

func (c ContextKey) String() string {
	return "appy." + string(c)
}
