package common

import (
	"os"
)

// Find takes a slice and looks for an element in it. Returns a bool.
//
func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// FileExists checks if a file exists
//
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
