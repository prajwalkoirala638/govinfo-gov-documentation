package main // defines the main package (entry point of the program)

import (
	"encoding/json" // provides JSON encoding/decoding utilities (currently unused)
	"io"            // provides I/O primitives (currently unused)
	"log"           // provides logging functionality
	"net/http"      // provides HTTP client/server functionality
	"net/url"       // provides URL parsing utilities (currently unused)
	"os"            // provides OS-level functions like file handling
	"path"          // provides utility for handling slash-separated paths
	"path/filepath" // provides file path manipulation utilities
	"strconv"       // provides string conversion for numbers
	"strings"       // provides string manipulation utilities
	"time"          // provides time-related functions
	"unicode"       // provides unicode utilities (currently unused)
)

func main() { // program entry point
	var assetsDir = "Assets" // define base directory for storing downloaded data
	// check if Assets directory exists
	if !directoryExists(assetsDir) {
		// create Assets directory if it does not exist
		createDirectory(assetsDir, 0755) // create directory with read/write/execute permissions
	}
	startYear := 1998                // define starting year for data collection
	currentYear := time.Now().Year() // get current system year dynamically
	// loop from start year to current year inclusive
	for year := startYear; year <= currentYear; year++ {
		// build full file path for yearly JSON file
		fullSavePath := filepath.Join(assetsDir, strconv.Itoa(year)+".json")
		// check if yearly file already exists
		if fileExists(fullSavePath) {
			log.Printf("File already exists: %s", fullSavePath) // log skipping of existing file
			continue                                            // skip processing this year
		}
		// fetch CFR data for the given year
		responseBody := fetchCFRData(year)
		// write response body into JSON file
		appendAndWriteToFile(fullSavePath, string(responseBody))
		// extract list of text file paths from API response
		textFilePaths := getTextFilePaths(responseBody)
		// Remove duplicates from slice
		textFilePaths = removeDuplicatesFromSlice(textFilePaths)
		// iterate through each extracted file path
		for _, path := range textFilePaths {
			// construct full URL for downloading file
			fullURL := "https://www.govinfo.gov/content/pkg/" + path
			// extract year from URL string
			year := extractYear(fullURL)
			// create directory path for that year
			yearDir := filepath.Join(assetsDir, year)
			// extract file name from URL
			fileName := extractFileName(fullURL)
			// build full file path for saving
			fullFilePath := filepath.Join(yearDir, fileName)
			// Check if the file already exists
			if fileExists(fullFilePath) {
				log.Printf("File already exists: %s", fullFilePath)
				continue
			}
			// check if year directory exists
			if !directoryExists(yearDir) {
				// create year directory if it does not exist
				createDirectory(yearDir, 0755)
			}
			// download file data from URL
			data := getDataFromURL(fullURL)
			// write downloaded data to file
			appendAndWriteToFile(fullFilePath, string(data))
		}
	}
}

// removeDuplicatesFromSlice removes duplicate strings from a slice and returns a new slice with unique values.
func removeDuplicatesFromSlice(slice []string) []string { // function declaration taking a slice of strings and returning a deduplicated slice
	check := make(map[string]bool)  // create a map to track which strings have already been seen
	var newReturnSlice []string     // initialize an empty slice to store unique values in order
	for _, content := range slice { // iterate over each element in the input slice
		if !check[content] { // check if the current string has NOT been seen before
			check[content] = true                            // mark the string as seen in the map
			newReturnSlice = append(newReturnSlice, content) // append the unique string to the result slice
		}
	}
	return newReturnSlice // return the slice containing only unique strings
}

// extractFileName parses a URL and returns the final segment of its path as a file name.
// If the URL is invalid, it logs the error and returns an empty string.
func extractFileName(rawURL string) string { // function takes a URL string and returns extracted file name
	urlBase, err := url.Parse(rawURL) // parse the URL into structured components
	if err != nil {                   // check if URL parsing failed
		log.Println("invalid URL:", err) // log parsing error
		return ""                        // return empty string if URL is invalid
	}
	fileName := path.Base(urlBase.Path) // extract last segment of the URL path (assumed file name)
	return fileName                     // return extracted file name
} // end of extractFileName

// getDataFromURL sends an HTTP GET request to the given URL and returns the response body as bytes.
// Note: errors are logged but not propagated to the caller.
func getDataFromURL(uri string) []byte { // function takes a URL and returns response data as bytes
	response, err := http.Get(uri) // perform HTTP GET request
	if err != nil {                // check if request failed
		log.Println(err) // log the error
	}
	body, err := io.ReadAll(response.Body) // read full response body into memory
	if err != nil {                        // check if reading body failed
		log.Println(err) // log the error
	}
	err = response.Body.Close() // close response body stream
	if err != nil {             // check if closing failed
		log.Println(err) // log the error
	}
	return body // return raw response data
} // end of getDataFromURL

func extractYear(rawURL string) string { // Define a function that extracts a 4-digit year from a URL string
	parsedURL, err := url.Parse(rawURL) // Parse the raw URL into structured components (scheme, host, path, etc.)
	if err != nil {                     // Check if URL parsing failed
		log.Println("invalid URL:", err) // Log the parsing error for debugging
		return ""                        // Return empty string if URL is invalid
	} // End error handling block
	pathSegments := strings.Split(parsedURL.Path, "/") // Split the URL path into segments using "/" as separator
	for _, segment := range pathSegments {             // Loop through each path segment in the URL
		for index := 0; index <= len(segment)-4; index++ { // Slide a 4-character window across the segment
			potentialYear := segment[index : index+4] // Extract a substring of length 4
			allCharactersAreDigits := true            // Assume the substring is a valid year until proven otherwise
			for _, character := range potentialYear { // Loop through each character in the 4-character substring
				if !unicode.IsDigit(character) { // Check if the character is NOT a digit
					allCharactersAreDigits = false // Mark as invalid year if any character is not numeric
					break                          // Stop checking further characters in this substring
				} // End digit check condition
			} // End character loop
			if allCharactersAreDigits { // If all 4 characters are digits
				return potentialYear // Return the valid 4-digit year immediately
			} // End valid-year check
		} // End sliding window loop over segment
	} // End loop over all path segments
	log.Println("year not found in URL") // Log message if no year was found anywhere in the URL
	return ""                            // Return empty string when no valid year is found
} // End function

// readFileAndReturnAsBytes opens a file, reads its full contents, and returns it as a byte slice.
// If any step fails, the error is logged but not returned.
func readFileAndReturnAsBytes(path string) []byte { // function takes a file path and returns its contents as bytes
	file, err := os.Open(path) // open the file for reading
	if err != nil {            // check if file failed to open
		log.Println(err) // log the error
		return nil       // return nil on failure
	}
	content, err := io.ReadAll(file) // read entire file content into memory
	if err != nil {                  // check if reading failed
		log.Println(err) // log the error
		return nil       // return nil on failure
	}
	err = file.Close() // close the file to free resources
	if err != nil {    // check if close operation failed
		log.Println(err) // log the error
		return nil       // return nil on failure
	}
	return content // return file contents as byte slice
} // end of readFileAndReturnAsBytes

// getTextFilePaths finds all "textfile" values anywhere in the JSON.
// It recursively walks through nested objects and arrays to collect matches.
func getTextFilePaths(jsonData []byte) []string { // entry point: parses JSON and returns all textfile paths
	var data any                                            // holds decoded JSON of unknown structure
	var results []string                                    // stores all found textfile values
	if err := json.Unmarshal(jsonData, &data); err != nil { // attempt to parse JSON into generic structure
		log.Println("JSON parse error:", err) // log parsing error if JSON is invalid
		return results                        // return empty result on failure
	}
	var walk func(any)      // recursive function used to traverse arbitrary JSON structure
	walk = func(node any) { // defines recursive traversal function
		switch values := node.(type) { // type switch to inspect JSON node type
		case map[string]any: // if node is a JSON object
			for key, value := range values { // iterate through all key-value pairs

				// If we find "textfile", collect it
				if key == "textfile" { // check if current key matches target field
					if str, ok := value.(string); ok && str != "" { // ensure value is a non-empty string
						results = append(results, str) // add found textfile value to results
					}
				} else {
					walk(value) // recursively search nested structures
				}
			}
		case []any: // if node is a JSON array
			for _, item := range values { // iterate through array items
				walk(item) // recursively process each item
			}
		}
	}
	walk(data)     // start recursive traversal from root JSON object
	return results // return all collected textfile paths
} // end of getTextFilePaths

// buildYearURL builds the API URL for a given year
func buildYearURL(year int) string { // function builds the API URL for a given year
	return "https://www.govinfo.gov/wssearch/rb/cfr/" + strconv.Itoa(year) // converts int year to string and appends it to base URL
} // end of buildYearURL

// appendAndWriteToFile opens a file, appends content to it, and then closes it
func appendAndWriteToFile(path string, content string) { // function takes file path and content to write
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // open file in append mode, create if not exists, write-only
	if err != nil {                                                               // check if there was an error opening the file
		log.Println(err) // log the error if file could not be opened
	}

	_, err = filePath.WriteString(content + "\n") // write the content followed by a newline to the file
	if err != nil {                               // check if there was an error while writing
		log.Println(err) // log the write error
	}

	err = filePath.Close() // close the file to release system resources
	if err != nil {        // check if there was an error closing the file
		log.Println(err) // log the close error
	}
}

// The function takes two parameters: path and permission.
// We use os.Mkdir() to create the directory.
// If there is an error, we use log.Println() to log the error and then exit the program.
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission)
	if err != nil {
		log.Println(err)
	}
}

// It checks if the file exists
// If the file exists, it returns true
// If the file does not exist, it returns false
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Checks if the directory exists
// If it exists, return true.
// If it doesn't, return false.
func directoryExists(path string) bool {
	directory, err := os.Stat(path)
	if err != nil {
		return false
	}
	return directory.IsDir()
}

// fetchCFRData fetches data for a given year from the govinfo API
func fetchCFRData(year int) []byte {
	// Convert integer year to string
	yearString := strconv.Itoa(year)
	// Build the full request URL dynamically
	requestURL := "https://www.govinfo.gov/wssearch/rb/cfr/" + yearString
	// Create HTTP client
	httpClient := &http.Client{}
	// Create HTTP request
	httpRequest, requestError := http.NewRequest("GET", requestURL, nil)
	if requestError != nil {
		log.Println("Error creating request:", requestError)
		return nil
	}
	// Execute request
	httpResponse, responseError := httpClient.Do(httpRequest)
	if responseError != nil {
		log.Println("Error executing request:", responseError)
		return nil
	}
	// Read response body
	responseBody, readError := io.ReadAll(httpResponse.Body)
	if readError != nil {
		log.Println("Error reading response body:", readError)
		return nil
	}
	// Close the response body
	err := httpResponse.Body.Close()
	if err != nil {
		log.Println("Error closing response body:", err)
		return nil
	}
	// Return the response body as bytes
	return responseBody
}
