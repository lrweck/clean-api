package envutil

func OTELExporterEndpointGo() string {
	return GetString("OTEL_EXPORTER_OTLP_ENDPOINT_GO", "127.0.0.1:4317")
}

func CurrentEnv() string {
	return GetString("ENV", "devel")
}

func AppName() string {
	return GetString("APP_NAME", "clean-api")
}
