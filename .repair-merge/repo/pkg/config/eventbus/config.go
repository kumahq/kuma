package eventbus

type Config struct {
	// BufferSize controls the buffer for every single event listener.
	// If we go over buffer, additional delay may happen to various operation like insight recomputation or KDS.
	BufferSize uint `json:"bufferSize" envconfig:"kuma_event_bus_buffer_size"`
}

func (c Config) Validate() error {
	return nil
}

func Default() Config {
	return Config{
		BufferSize: 100,
	}
}
