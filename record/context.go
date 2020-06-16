package record

type contextKey string

func (c contextKey) String() string {
	return "record." + string(c)
}
