package videorequest

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"logging"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func ReadJSON(filename, tweetID string) (string, string) {
	file, err := os.Open(filename)
	if err != nil {
		logging.Println(err)
		return "", ""
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		logging.Fatal("Error:", err)
		return "", ""
	}

	var data map[string]interface{}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		logging.Fatal("Error:", err)
		return "", ""
	}

	variablesMap, _ := data["variables"].(map[string]interface{})

	variablesMap["tweetId"] = tweetID

	jsonFeatures, _ := json.MarshalIndent(data["features"], "", " ")
	jsonVariables, _ := json.MarshalIndent(data["variables"], "", " ")
	encodedJSONFeatures := url.QueryEscape(string(jsonFeatures))
	encodedJSONVariables := url.QueryEscape(string(jsonVariables))
	encodedJSONVariables = strings.Replace(encodedJSONVariables, "+", "%20", -1)
	encodedJSONVariables = strings.Replace(encodedJSONVariables, "%0A", "", -1)
	encodedJSONFeatures = strings.Replace(encodedJSONFeatures, "+", "%20", -1)
	encodedJSONFeatures = strings.Replace(encodedJSONFeatures, "%0A", "", -1)

	return encodedJSONFeatures, encodedJSONVariables
}

func getBearerToken(body string) string {
	pattern := `AAAAAAAAA[^"]+`
	regex, err := regexp.Compile(pattern)

	if err != nil {
		logging.Fatal("Error:", err)
		return ""
	}
	matches := regex.FindStringSubmatch(body)
	for _, match := range matches {
		logging.Println(match)
	}
	return matches[0]
}

func GetTokens(URL string) (bearerToken, queryID, guestToken string) {
	body := MakeRequest(URL)
	mainJSURL := getMainJSURL(body)
	logging.Println("mainJSURL: ", mainJSURL)
	mainJSBody := MakeRequest(mainJSURL)
	bearerToken = getBearerToken(mainJSBody)
	queryID = "0hWvDhmW8YQ-S_ib3azIrw" //magic
	guestToken = getGuestToken(bearerToken)

	return bearerToken, queryID, guestToken
}

func getGuestToken(bearerToken string) string {
	client := &http.Client{}
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:84.0) Gecko/20100101 Firefox/84.0",
		"Accept":          "*/*",
		"Accept-Language": "de,en-US;q=0.7,en;q=0.3",
		"Accept-Encoding": "gzip, deflate, br",
		"TE":              "trailers",
		"Authorization":   fmt.Sprintf("Bearer %s", bearerToken),
	}

	req, err := http.NewRequest("POST", "https://api.twitter.com/1.1/guest/activate.json", nil)
	if err != nil {
		logging.Fatal("Error:", err)
		return ""
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.Fatal("Error:", err)
		return ""
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var reader io.Reader
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			var err error
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				logging.Fatal("Error creating gzip reader:", err)
				return ""
			}
		default:
			reader = resp.Body
		}

		body, err := io.ReadAll(reader)
		if err != nil {
			logging.Fatal("Error reading response:", err)
			return ""
		}

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			logging.Fatal("Error decoding response:", err)
			return ""
		}
		guestToken, ok := result["guest_token"].(string)
		if !ok {
			logging.Fatal("Guest token not found in response")
			return ""
		}
		return guestToken
	}

	return ""
}

// Doesn't work with the queryID we're getting from this function. I don't know why.
func getQueryID(body string) string {
	pattern := `queryId:"(.+?)"`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		logging.Fatal("Error:", err)
		return ""
	}

	matches := regex.FindStringSubmatch(body)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func getMainJSURL(body string) string {

	pattern := `https://abs\.twimg\.com/responsive-web/client-web/main\.[^\.]+\.js`
	//fmt.Println("pattern:", body)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		logging.Fatal("Error:", err)
		return ""
	}

	matches := regex.FindStringSubmatch(body)

	for _, match := range matches {
		logging.Println(match)
	}
	return matches[0]

}

// MakeRequest handles HTTP GET requests and follows redirects
func MakeRequest(URL string) string {
	chromeDriverPath := "/usr/bin/chromedriver" // Update this path to match your actual Chromedriver path

	// Start ChromeDriver service
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(chromeDriverPath, 9515, opts...)
	if err != nil {
		logging.Fatal("Error starting the ChromeDriver server: %v", err)
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
		logging.Fatal("Failed to open session: %v", err)
	}
	defer wd.Quit()

	// Navigate to the URL
	if err := wd.Get(URL); err != nil {
		logging.Fatal("Failed to load page: %v", err)
	}

	// Get page source
	pageSource, err := wd.PageSource()
	if err != nil {
		logging.Fatal("Failed to get page source: %v", err)
	}

	// Print page source
	return pageSource
}

func GetVideo(URL, destination string) {
	bearerToken, queryID, guestToken := GetTokens(URL)
	logging.Println("Bearer Token:", bearerToken)
	logging.Println("Query ID:", queryID)
	logging.Println("Guest Token:", guestToken)

	response := getTweetDetails(URL, bearerToken, queryID, guestToken)
	logging.Println("Response: ", response)
	videoURL := parseJSON(response)
	downloadVideo(videoURL, destination)
}

func downloadVideo(videoURL, destination string) error {

	resp, err := http.Get(videoURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP response status code: %d", resp.StatusCode)
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func parseJSON(body string) string {
	var data map[string]interface{}

	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		logging.Println("Error:", err)
		return ""
	}

	variants := extractField(data, "variants")

	logging.Println("variants: ", variants)

	var maxBitrate float64 = 0
	videoURL := ""
	for _, value := range variants.([]interface{}) {
		valueMap := value.(map[string]interface{})
		bitrate := valueMap["bitrate"]
		if bitrate != nil {
			if bitrate.(float64) > maxBitrate {
				maxBitrate = bitrate.(float64)
				videoURL = valueMap["url"].(string)
			}
		}
	}
	logging.Println("videoURL: ", videoURL)
	logging.Println("bitrate: ", maxBitrate)

	return videoURL
}

func extractField(data map[string]interface{}, fieldToExtract string) interface{} {
	for key, value := range data {
		if key == fieldToExtract {
			return value
		}
		switch v := value.(type) {
		case map[string]interface{}:
			if result := extractField(v, fieldToExtract); result != nil {
				return result
			}
		case []interface{}:
			for _, elem := range v {
				if submap, ok := elem.(map[string]interface{}); ok {
					if result := extractField(submap, fieldToExtract); result != nil {
						return result
					}
				}
			}
		}
	}
	return nil
}

func getFeaturesAndVariables(tweetID string) (string, string) {
	features, variables := ReadJSON("data.json", tweetID)
	return features, variables
}

func getTweetDetails(URL, bearerToken, queryID, guestToken string) string {
	pattern := `/status/(\d+)`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		logging.Fatal("Error:", err)
		return ""
	}

	tweetID := regex.FindStringSubmatch(URL)[1]
	logging.Println("tweetID: ", tweetID)
	features, variables := getFeaturesAndVariables(tweetID)

	newURL := getDetailsURL(tweetID, queryID, features, variables)

	logging.Println("newURL: ", newURL)
	client := &http.Client{}

	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		logging.Println("Error creating request:", err)
		return ""
	}

	req.Header.Set("authorization", "Bearer "+bearerToken)
	req.Header.Set("x-guest-token", guestToken)
	var resp *http.Response
	var err1 error
	for i := 0; i < 5; i++ {

		resp, err1 = client.Do(req)
		if err1 != nil {
			logging.Println("Error performing request:", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)

		}
		if resp.StatusCode == http.StatusOK {
			break
		}

	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.Println("Error reading response body:", err)
		return ""
	}

	return string(bodyBytes)
}

func getDetailsURL(tweetID, queryIDToken, features, variables string) string {
	logging.Println("queryIDToken:", queryIDToken)
	return fmt.Sprintf("https://twitter.com/i/api/graphql/%s/TweetResultByRestId?variables=%s&features=%s", queryIDToken, variables, features)
}
