package utils

import (
	"github.com/litmuschaos/admission-controller/pkg/log"
	"regexp"
)

func MatchRegex(pattern, s string) bool {
	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Logger.Errorf("Error compiling regex:", err)
		return false
	}

	return re.MatchString(s)
}
