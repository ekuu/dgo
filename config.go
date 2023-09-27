package dgo

//
//type Config struct {
//	logTraceIDKey   string
//	logTraceHandler func(h slog.Handler) slog.Handler
//}
//
//func (c *Config) LogTraceIDKey() string {
//	if c.logTraceIDKey != "" {
//		return c.logTraceIDKey
//	}
//	return "traceID"
//}
//
//func (c *Config) LogTraceHandler() func(h slog.Handler) slog.Handler {
//	if c.logTraceHandler != nil {
//		return c.logTraceHandler
//	}
//	return func(h slog.Handler) slog.Handler {
//		return log.TraceHandler(c.LogTraceIDKey(), h)
//	}
//}
//
//func GetConfig() *Config {
//	return new(Config)
//}
