package utils

import "strings"

func ContainsAny(s string, tokens []string) bool {
	for _, t := range tokens {
		if t != "" && strings.Contains(s, strings.ToLower(t)) {
			return true
		}
	}
	return false
}
