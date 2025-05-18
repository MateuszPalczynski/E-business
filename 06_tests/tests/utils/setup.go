// tests/utils/setup.go
package utils

import (
	"testing"

	"github.com/tebeka/selenium"
)

func Setup(t *testing.T) selenium.WebDriver {
	// Connect to ChromeDriver
	wd, err := selenium.NewRemote(selenium.Capabilities{
		"browserName": "chrome",
	}, "http://localhost:9515") // Direct ChromeDriver connection
	
	if err != nil {
		t.Fatalf("Failed to open session: %v", err)
	}
	return wd
}