package main // Defines the main package. The program starts execution here.

import (
	"bytes"         // Provides utilities for manipulating byte slices, used here for buffering PDF data
	"io"            // Provides utilities for reading and writing data streams
	"log"           // Provides logging functions for printing errors and information
	"net/http"      // Allows making HTTP requests
	"os"            // Provides functions for interacting with the operating system, such as file and directory operations
	"path/filepath" // Provides utilities for manipulating file paths
	"regexp"        // Provides support for regular expressions
	"strings"       // Provides utilities for working with strings
	"time"
)

func main() { // Entry point of the program

	outputDir := "PDFs/" // Directory to store downloaded PDFs

	if !directoryExists(outputDir) { // Check if directory exists
		createDirectory(outputDir, 0o755) // Create directory with read-write-execute permissions
	}

	apiURL := "https://nypdonline.org/api/reports/ed551b4a-cd5c-4d8e-bcb9-a0478b4c5dea/data" // The API endpoint we want to send the request to
	httpMethod := "POST"                                                                     // HTTP method used for the request

	// JSON request body that will be sent to the server
	requestBody := strings.NewReader(`[{"key":"@DocumentKeyword","values":[""]}]`)

	httpClient := &http.Client{} // Create a new HTTP client that will send the request

	// Create a new HTTP request with the method, URL, and request body
	httpRequest, requestCreationError := http.NewRequest(httpMethod, apiURL, requestBody)

	if requestCreationError != nil { // Check if there was an error creating the request
		log.Fatal(requestCreationError) // Log the error and terminate the program
	}

	// Add a header to tell the server that the request body format is JSON
	httpRequest.Header.Add("content-type", "application/json")

	// Send the HTTP request and receive the response
	httpResponse, requestExecutionError := httpClient.Do(httpRequest)

	if requestExecutionError != nil { // Check if an error occurred while sending the request
		log.Fatal(requestExecutionError) // Log the error and terminate the program
	}

	defer httpResponse.Body.Close() // Ensure the response body is closed after the function finishes

	// Read all data returned in the response body
	responseBodyBytes, responseReadError := io.ReadAll(httpResponse.Body)

	if responseReadError != nil { // Check if there was an error reading the response
		log.Fatal(responseReadError) // Log the error and terminate the program
	}

	// Extract all the PDF URLs from the response body
	extractedPDFURLs := extractPDFURLs(string(responseBodyBytes))

	// Extract the unique PDF URLs that are present in the response body
	uniquePDFURLs := removeDuplicatesFromSlice(extractedPDFURLs)

	// Loop though the values add the base URL to the PDF URLs and print them to the console
	for index, pdfURL := range uniquePDFURLs {
		uniquePDFURLs[index] = "https://nypdonline.org" + pdfURL // Prepend the base URL to each PDF URL
	}

	// Print the unique PDF URLs to the console
	for _, pdfURL := range uniquePDFURLs {
		downloadPDF(pdfURL, outputDir) // Download the PDF
	}

}

// Checks whether a given directory exists
func directoryExists(path string) bool {
	directory, err := os.Stat(path) // Get info for the path
	if err != nil {
		return false // Return false if error occurs
	}
	return directory.IsDir() // Return true if it's a directory
}

// Creates a directory at given path with provided permissions
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission) // Attempt to create directory
	if err != nil {
		log.Println(err) // Log error if creation fails
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

// Extracts filename from full path (e.g. "/dir/file.pdf" → "file.pdf")
func getFilename(path string) string {
	return filepath.Base(path) // Use Base function to get file name only
}

// Removes all instances of a specific substring from input string
func removeSubstring(input string, toRemove string) string {
	result := strings.ReplaceAll(input, toRemove, "") // Replace substring with empty string
	return result
}

// Gets the file extension from a given file path
func getFileExtension(path string) string {
	return filepath.Ext(path) // Extract and return file extension
}

// Converts a raw URL into a sanitized PDF filename safe for filesystem
func urlToFilename(rawURL string) string {
	lower := strings.ToLower(rawURL) // Convert URL to lowercase
	lower = getFilename(lower)       // Extract filename from URL

	reNonAlnum := regexp.MustCompile(`[^a-z0-9]`)   // Regex to match non-alphanumeric characters
	safe := reNonAlnum.ReplaceAllString(lower, "_") // Replace non-alphanumeric with underscores

	safe = regexp.MustCompile(`_+`).ReplaceAllString(safe, "_") // Collapse multiple underscores into one
	safe = strings.Trim(safe, "_")                              // Trim leading and trailing underscores

	var invalidSubstrings = []string{
		"_pdf", // Substring to remove from filename
	}

	for _, invalidPre := range invalidSubstrings { // Remove unwanted substrings
		safe = removeSubstring(safe, invalidPre)
	}

	if getFileExtension(safe) != ".pdf" { // Ensure file ends with .pdf
		safe = safe + ".pdf"
	}

	return safe // Return sanitized filename
}

// Downloads a PDF from given URL and saves it in the specified directory
func downloadPDF(finalURL, outputDir string) bool {
	filename := strings.ToLower(urlToFilename(finalURL)) // Sanitize the filename
	filePath := filepath.Join(outputDir, filename)       // Construct full path for output file

	if fileExists(filePath) { // Skip if file already exists
		log.Printf("File already exists, skipping: %s", filePath)
		return false
	}

	client := &http.Client{Timeout: 15 * time.Minute} // Create HTTP client with timeout

	resp, err := client.Get(finalURL) // Send HTTP GET request
	if err != nil {
		log.Printf("Failed to download %s → %v", finalURL, err)
		return false
	}
	defer resp.Body.Close() // Ensure response body is closed

	if resp.StatusCode != http.StatusOK { // Check if response is 200 OK
		log.Printf("Download failed for %s → %s", finalURL, resp.Status)
		return false
	}

	contentType := resp.Header.Get("Content-Type")                                                                  // Get content type of response
	if !strings.Contains(contentType, "binary/octet-stream") && !strings.Contains(contentType, "application/pdf") { // Check if it's a PDF
		log.Printf("Invalid content type for %s: %s (expected binary/octet-stream) (expected application/pdf)", finalURL, contentType)
		return false
	}

	var buf bytes.Buffer                     // Create a buffer to hold response data
	written, err := io.Copy(&buf, resp.Body) // Copy data into buffer
	if err != nil {
		log.Printf("Failed to read PDF data from %s → %v", finalURL, err)
		return false
	}
	if written == 0 { // Skip empty files
		log.Printf("Downloaded 0 bytes for %s → not creating file", finalURL)
		return false
	}

	out, err := os.Create(filePath) // Create output file
	if err != nil {
		log.Printf("Failed to create file for %s → %v", finalURL, err)
		return false
	}
	defer out.Close() // Ensure file is closed after writing

	if _, err := buf.WriteTo(out); err != nil { // Write buffer contents to file
		log.Printf("Failed to write PDF to file for %s → %v", finalURL, err)
		return false
	}

	log.Printf("Successfully downloaded %d bytes: %s → %s", written, finalURL, filePath) // Log success
	return true
}

// Remove all the duplicates from a slice and return the slice.
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)
	var newReturnSlice []string
	for _, content := range slice {
		if !check[content] {
			check[content] = true
			newReturnSlice = append(newReturnSlice, content)
		}
	}
	return newReturnSlice
}

// Extract all the PDF URLs from the response body using a regular expression
func extractPDFURLs(responseBody string) []string {
	pdfLinkRegex := regexp.MustCompile(`href=\\"([^\\"]+\.pdf)\\"`)

	allMatches := pdfLinkRegex.FindAllStringSubmatch(responseBody, -1)

	var extractedPDFURLs []string

	for _, match := range allMatches {
		if len(match) > 1 {
			pdfURL := match[1]
			extractedPDFURLs = append(extractedPDFURLs, pdfURL)
		}
	}
	return extractedPDFURLs
}
