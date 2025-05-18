package tests

import (
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

func TestHomePage(t *testing.T) {
	// Use absolute path to ChromeDriver
	service, err := selenium.NewChromeDriverService(
		`C:\Users\User\Desktop\Studia_UJ\E-biznes\06_tests\tests\chromedriver.exe`,
		9515,
	)
	if err != nil {
		t.Fatalf("Failed to start ChromeDriver: %v", err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, "http://localhost:9515")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer wd.Quit()

	t.Run("HomePage", func(t *testing.T) {
		err := wd.Get("http://localhost:1323")
		if err != nil {
			t.Fatalf("Failed to load page: %v", err)
		}

		// Basic content check
		_, err = wd.FindElement(selenium.ByTagName, "body")
		if err != nil {
			t.Fatal("Page body not found")
		}
	})

	time.Sleep(1 * time.Second)
}
		