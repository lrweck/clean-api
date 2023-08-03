package envutil

import (
	"os"
	"strconv"
	"time"
)

func GetBool(key string, def bool) bool {
	s, ok := os.LookupEnv(key)
	if !ok {
		return def
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}

	return b
}

func GetInt(key string, def int) int {
	s, ok := os.LookupEnv(key)
	if !ok {
		return def
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}

	return i
}

func GetDuration(key string, def time.Duration) time.Duration {
	s, ok := os.LookupEnv(key)
	if !ok {
		return def
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}

	return d
}

func GetString(key string, def string) string {
	s, ok := os.LookupEnv(key)
	if !ok {
		return def
	}

	return s
}
