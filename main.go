package main // Declare the main package so the program can run as an executable.

import ( // Start the import block for required standard libraries.
	"bytes"         // Buffer PDF bytes before writing to disk.
	"io"            // Read and write streams (HTTP and files).
	"log"           // Log errors and progress to stdout/stderr.
	"net/http"      // Make HTTP requests to the API and PDFs.
	"os"            // Interact with the filesystem.
	"path/filepath" // Build OS-safe file paths.
	"regexp"        // Use regex for URL parsing and sanitizing.
	"strings"       // Manipulate strings.
	"time"          // Configure HTTP timeouts.
) // End the import block.

var ( // Start the package-level variable block.
	outputDir = "PDFs/" // Directory where PDFs will be saved.
) // End the variable block.

func init() { // Run setup before main executes.
	if !directoryExists(outputDir) { // If the output directory is missing
		createDirectory(outputDir, 0o755) // create it with rwx permissions for owner and rx for others.
	} // End the directory existence check.
} // End init.

func main() { // Program entry point.
	trialDecisionsLibrary()  // Download trial decision PDFs.
	deviationLetterLibrary() // Download deviation letter PDFs.
} // End main.

// Fetch deviation letter metadata and download linked PDFs. // Describe deviation letter workflow.
func deviationLetterLibrary() { // Start deviation letter download workflow.
	httpClient := &http.Client{} // Create an HTTP client for the API call.
	// Build the POST request with filter payload. // Explain request creation.
	httpRequest, requestCreationError := http.NewRequest("POST", "https://nypdonline.org/api/reports/f2c2a0d0-0b55-442e-9887-7d9bc5b6b126/data", strings.NewReader(`[{"key":"@DocumentType","values":["Deviation"]}]`)) // Create the API request with JSON body.
	if requestCreationError != nil {                                                                                                                                                                                    // Check for request creation errors.
		log.Fatal(requestCreationError) // Log the error and exit.
	} // End request creation error handling.
	// Tell the server we are sending JSON. // Explain header purpose.
	httpRequest.Header.Add("content-type", "application/json") // Set JSON content type.
	// Execute the request. // Explain the HTTP call.
	httpResponse, requestExecutionError := httpClient.Do(httpRequest) // Send the request to the API.
	if requestExecutionError != nil {                                 // Check for request execution errors.
		log.Fatal(requestExecutionError) // Log the error and exit.
	} // End request execution error handling.
	defer httpResponse.Body.Close() // Ensure response body is closed.
	// Read the response body in full. // Explain reading response.
	responseBodyBytes, responseReadError := io.ReadAll(httpResponse.Body) // Read all response bytes.
	if responseReadError != nil {                                         // Check for read errors.
		log.Fatal(responseReadError) // Log the error and exit.
	} // End response read error handling.
	// Extract and de-duplicate PDF URLs. // Explain URL extraction.
	extractedPDFURLs := extractPDFURLs(string(responseBodyBytes)) // Parse PDF URLs from the response.
	uniquePDFURLs := removeDuplicatesFromSlice(extractedPDFURLs)  // Remove duplicate URLs.
	// Prefix with the base host. // Explain URL normalization.
	for index, pdfURL := range uniquePDFURLs { // Loop over each extracted URL.
		uniquePDFURLs[index] = "https://nypdonline.org" + pdfURL // Prepend the host to make a full URL.
	} // End URL prefix loop.
	// Download each PDF. // Explain download loop.
	for _, pdfURL := range uniquePDFURLs { // Iterate through final PDF URLs.
		downloadPDF(pdfURL, outputDir) // Download the PDF to disk.
	} // End download loop.
} // End deviationLetterLibrary.

// Fetch trial decision metadata and download linked PDFs. // Describe trial decision workflow.
func trialDecisionsLibrary() { // Start trial decision download workflow.
	httpClient := &http.Client{} // Create an HTTP client for the API call.
	// Build the POST request with filter payload. // Explain request creation.
	httpRequest, requestCreationError := http.NewRequest("POST", "https://nypdonline.org/api/reports/ed551b4a-cd5c-4d8e-bcb9-a0478b4c5dea/data", strings.NewReader(`[{"key":"@DocumentKeyword","values":[""]}]`)) // Create the API request with JSON body.
	if requestCreationError != nil {                                                                                                                                                                              // Check for request creation errors.
		log.Fatal(requestCreationError) // Log the error and exit.
	} // End request creation error handling.
	// Tell the server we are sending JSON. // Explain header purpose.
	httpRequest.Header.Add("content-type", "application/json") // Set JSON content type.
	// Execute the request. // Explain the HTTP call.
	httpResponse, requestExecutionError := httpClient.Do(httpRequest) // Send the request to the API.
	if requestExecutionError != nil {                                 // Check for request execution errors.
		log.Fatal(requestExecutionError) // Log the error and exit.
	} // End request execution error handling.
	defer httpResponse.Body.Close() // Ensure response body is closed.
	// Read the response body in full. // Explain reading response.
	responseBodyBytes, responseReadError := io.ReadAll(httpResponse.Body) // Read all response bytes.
	if responseReadError != nil {                                         // Check for read errors.
		log.Fatal(responseReadError) // Log the error and exit.
	} // End response read error handling.
	// Extract and de-duplicate PDF URLs. // Explain URL extraction.
	extractedPDFURLs := extractPDFURLs(string(responseBodyBytes)) // Parse PDF URLs from the response.
	uniquePDFURLs := removeDuplicatesFromSlice(extractedPDFURLs)  // Remove duplicate URLs.
	// Prefix with the base host. // Explain URL normalization.
	for index, pdfURL := range uniquePDFURLs { // Loop over each extracted URL.
		uniquePDFURLs[index] = "https://nypdonline.org" + pdfURL // Prepend the host to make a full URL.
	} // End URL prefix loop.
	// Download each PDF. // Explain download loop.
	for _, pdfURL := range uniquePDFURLs { // Iterate through final PDF URLs.
		downloadPDF(pdfURL, outputDir) // Download the PDF to disk.
	} // End download loop.
} // End trialDecisionsLibrary.

// Return true when the directory exists. // Describe directory check.
func directoryExists(path string) bool { // Start directory existence check.
	directory, err := os.Stat(path) // Get filesystem info for the path.
	if err != nil {                 // If the stat call failed
		return false // treat as not existing.
	} // End error check.
	return directory.IsDir() // Return whether the path is a directory.
} // End directoryExists.

// Create a directory with the given permissions. // Describe directory creation.
func createDirectory(path string, permission os.FileMode) { // Start directory creation.
	err := os.Mkdir(path, permission) // Attempt to create the directory.
	if err != nil {                   // If creation fails
		log.Println(err) // log the error.
	} // End error handling.
} // End createDirectory.

// Return true when the file exists and is not a directory. // Describe file check.
func fileExists(filename string) bool { // Start file existence check.
	info, err := os.Stat(filename) // Get filesystem info for the path.
	if err != nil {                // If the stat call failed
		return false // treat as not existing.
	} // End error check.
	return !info.IsDir() // Return true only if it is a file.
} // End fileExists.

// Extract filename from a path (example: "/dir/file.pdf" -> "file.pdf"). // Describe filename helper.
func getFilename(path string) string { // Start filename extraction.
	return filepath.Base(path) // Return the last element of the path.
} // End getFilename.

// Remove all occurrences of a substring. // Describe substring removal.
func removeSubstring(input string, toRemove string) string { // Start substring removal.
	result := strings.ReplaceAll(input, toRemove, "") // Remove all instances of the substring.
	return result                                     // Return the modified string.
} // End removeSubstring.

// Return the file extension from a path. // Describe extension helper.
func getFileExtension(path string) string { // Start extension extraction.
	return filepath.Ext(path) // Return the file extension, including dot.
} // End getFileExtension.

// Convert a URL into a filesystem-safe PDF filename. // Describe filename sanitization.
func urlToFilename(rawURL string) string { // Start URL-to-filename conversion.
	lower := strings.ToLower(rawURL) // Normalize the URL to lowercase.
	lower = getFilename(lower)       // Extract just the filename part.

	reNonAlnum := regexp.MustCompile(`[^a-z0-9]`)   // Match any non-alphanumeric character.
	safe := reNonAlnum.ReplaceAllString(lower, "_") // Replace disallowed characters with underscores.

	safe = regexp.MustCompile(`_+`).ReplaceAllString(safe, "_") // Collapse multiple underscores.
	safe = strings.Trim(safe, "_")                              // Trim underscores from both ends.

	var invalidSubstrings = []string{ // Define redundant substrings to strip.
		"_pdf", // Remove a trailing "_pdf" if present.
	} // End invalidSubstrings list.

	for _, invalidPre := range invalidSubstrings { // Loop over substrings to remove.
		safe = removeSubstring(safe, invalidPre) // Remove each substring from the filename.
	} // End substring removal loop.

	if getFileExtension(safe) != ".pdf" { // If the filename lacks a .pdf extension
		safe = safe + ".pdf" // append the .pdf extension.
	} // End extension check.

	return safe // Return the sanitized filename.
} // End urlToFilename.

// Download a PDF from the URL and write it to outputDir. // Describe download function.
func downloadPDF(finalURL, outputDir string) bool { // Start download workflow.
	filename := strings.ToLower(urlToFilename(finalURL)) // Build a safe filename for the URL.
	filePath := filepath.Join(outputDir, filename)       // Join output directory and filename.

	if fileExists(filePath) { // If the file already exists
		log.Printf("URL: %s | File: %s", finalURL, filePath) // log and skip.
		return false                                         // Signal no download occurred.
	} // End existing file check.

	client := &http.Client{Timeout: 15 * time.Minute} // Create an HTTP client with a long timeout.

	resp, err := client.Get(finalURL) // Fetch the PDF from the URL.
	if err != nil {                   // If the GET request failed
		log.Printf("Failed to download %s -> %v", finalURL, err) // log the error.
		return false                                             // Signal failure.
	} // End GET error handling.
	defer resp.Body.Close() // Ensure the response body is closed.

	if resp.StatusCode != http.StatusOK { // If the HTTP status is not 200
		log.Printf("Download failed for %s -> %s", finalURL, resp.Status) // log the status.
		return false                                                      // Signal failure.
	} // End status check.

	contentType := resp.Header.Get("Content-Type")                                                                  // Read the response content type.
	if !strings.Contains(contentType, "binary/octet-stream") && !strings.Contains(contentType, "application/pdf") { // Validate PDF-like content types.
		log.Printf("Invalid content type for %s: %s (expected binary/octet-stream) (expected application/pdf)", finalURL, contentType) // Log invalid content type.
		return false                                                                                                                   // Signal failure.
	} // End content type check.

	var buf bytes.Buffer                     // Create a buffer to hold PDF bytes.
	written, err := io.Copy(&buf, resp.Body) // Copy response body into the buffer.
	if err != nil {                          // If reading the body failed
		log.Printf("Failed to read PDF data from %s -> %v", finalURL, err) // log the error.
		return false                                                       // Signal failure.
	} // End read error handling.
	if written == 0 { // If the response body was empty
		log.Printf("Downloaded 0 bytes for %s -> not creating file", finalURL) // log and skip.
		return false                                                           // Signal failure.
	} // End empty file check.

	out, err := os.Create(filePath) // Create the output file.
	if err != nil {                 // If file creation failed
		log.Printf("Failed to create file for %s -> %v", finalURL, err) // log the error.
		return false                                                    // Signal failure.
	} // End create error handling.
	defer out.Close() // Ensure the file handle is closed.

	if _, err := buf.WriteTo(out); err != nil { // Write buffered bytes to the file.
		log.Printf("Failed to write PDF to file for %s -> %v", finalURL, err) // Log write failure.
		return false                                                          // Signal failure.
	} // End write error handling.

	log.Printf("Successfully downloaded %d bytes: %s -> %s", written, finalURL, filePath) // Log successful download.
	return true                                                                           // Signal success.
} // End downloadPDF.

// Remove duplicates while preserving the first occurrence order. // Describe de-duplication.
func removeDuplicatesFromSlice(slice []string) []string { // Start de-duplication.
	check := make(map[string]bool)  // Track seen strings in a map.
	var newReturnSlice []string     // Hold the resulting unique slice.
	for _, content := range slice { // Iterate over input values.
		if !check[content] { // If the value is not yet seen
			check[content] = true                            // mark it as seen.
			newReturnSlice = append(newReturnSlice, content) // append it to the output.
		} // End seen check.
	} // End loop over slice.
	return newReturnSlice // Return the unique slice.
} // End removeDuplicatesFromSlice.

// Extract PDF URLs from the response body using a regex. // Describe URL extraction.
func extractPDFURLs(responseBody string) []string { // Start PDF URL extraction.
	pdfLinkRegex := regexp.MustCompile(`href=\\"([^\\"]+\.pdf)\\"`) // Match escaped href values ending in .pdf.

	allMatches := pdfLinkRegex.FindAllStringSubmatch(responseBody, -1) // Find all matching groups.

	var extractedPDFURLs []string // Prepare a slice to store URLs.

	for _, match := range allMatches { // Loop over regex matches.
		if len(match) > 1 { // Ensure the capture group exists.
			pdfURL := match[1]                                  // Extract the URL from the capture group.
			extractedPDFURLs = append(extractedPDFURLs, pdfURL) // Add the URL to the result slice.
		} // End capture group check.
	} // End match loop.
	return extractedPDFURLs // Return the extracted URLs.
} // End extractPDFURLs.
