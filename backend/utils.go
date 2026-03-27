package checkered

import (
	"os"
)

func ParseStringOption(flagValue string, envName string, value string) string {
	result := value
	
	if envName != "" {
		res, found := os.LookupEnv(envName)

		if found {
			result = res
		}
	}

	if flagValue != "" {
		result = flagValue
	}

	return result
}
