package utils

import (
	"fmt"
	"os"
)

func MustGetEnv(envName string) string {
	envValue, ok := os.LookupEnv(envName)

	if !ok {
		panic(fmt.Sprintf("failed to get name=%s env", envName))
	}

	return envValue
}