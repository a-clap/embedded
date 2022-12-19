package embedded

type Option func(*Handler) error

type HardwareID string

func WithHeaters(heaters map[HardwareID]Heater) Option {
	return func(h *Handler) error {
		h.heaters = heaters
		return nil
	}
}

func WithLogger(l Logger) Option {
	return func(*Handler) error {
		log = l
		return nil
	}
}
