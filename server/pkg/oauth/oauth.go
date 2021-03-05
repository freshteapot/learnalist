package oauth

func NewHandlers() *Handlers {
	return &Handlers{}
}

func (h *Handlers) AddGoogle(handler OAuth2ConfigInterface) {
	h.keys = append(h.keys, IDPKeyGoogle)
	h.Google = handler
}

func (h *Handlers) AddAppleID(handler OAuth2ConfigInterface) {
	h.keys = append(h.keys, IDPKeyApple)
	h.AppleID = handler
}

func (h *Handlers) Keys() []string {
	return h.keys
}
