package dest

type Handler struct {
	config *config
}

type Option func(h *Handler) error

func WithHeaters(heaters HeatersImpl) Option {
	return func(h *Handler) error {
		h.config.heaters.impl = heaters
		return nil
	}
}

func New(opts ...Option) (*Handler, error) {
	h := &Handler{config: newConfig()}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	if err := h.config.heaters.init(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Handler) Config() Config {
	return h.config
}
