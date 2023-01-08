package rest

type Option func(*Server) error

func WithSensorHandler(s SensorHandler) Option {
	return func(server *Server) error {
		server.SensorHandler = s
		return nil
	}
}

func WithWifiHandler(w WIFIHandler) Option {
	return func(server *Server) error {
		server.WIFIHandler = w
		return nil
	}
}

func WithFormat(format Format) Option {
	return func(server *Server) error {
		server.fmt = format
		return nil
	}
}
