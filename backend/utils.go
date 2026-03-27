package checkered

import (
	"os"
)

// Determines the proper value of a given option. Priority goes in the following order:
// 1. `flagValue` is checked,
// 2. `envName` is looked up as an environment variable
// 3. `value` is used.
// The first successful branch will be the output of the function
func ParseStringOption(flagValue string, envName string, value string) string {
	if flagValue != "" { return flagValue }

	if envName != "" {
		env, success := os.LookupEnv(envName)
		if(success) { return env }
	}

	return value
}
