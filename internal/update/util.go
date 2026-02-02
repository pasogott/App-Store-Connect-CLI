package update

import (
	"os"
	"strconv"
)

func parseBool(value string) (bool, error) {
	return strconv.ParseBool(value)
}

func getEnv(name string) string {
	return os.Getenv(name)
}
