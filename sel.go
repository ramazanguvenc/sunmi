package main

import (
	"fmt"
	"log"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func sel() {
	chromeDriverPath := "/usr/bin/chromedriver" // Update this path to match your actual Chromedriver path

	// Start ChromeDriver service
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(chromeDriverPath, 9515, opts...)
	if err != nil {
		log.Fatalf("Error starting the ChromeDriver server: %v", err)
	}
	defer service.Stop()

	// Chrome options
	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: []string{
			"--headless",              // Run Chrome in headless mode
			"--no-sandbox",            // Disable sandboxing
			"--disable-dev-shm-usage", // Disable /dev/shm usage
		},
	}

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	caps.AddChrome(chromeCaps)

	// Create WebDriver session
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 9515))
	if err != nil {
		log.Fatalf("Failed to open session: %v", err)
	}
	defer wd.Quit()

	// Navigate to the URL
	if err := wd.Get("https://example.com"); err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}

	// Get page source
	pageSource, err := wd.PageSource()
	if err != nil {
		log.Fatalf("Failed to get page source: %v", err)
	}

	// Print page source
	fmt.Println(pageSource)
}
