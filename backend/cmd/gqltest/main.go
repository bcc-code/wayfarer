package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
)

type WayfarerClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"`
	jwt.RegisteredClaims
}

type TestCase struct {
	FilePath string
	UserID   string
	Query    string
	Variables map[string]interface{}
	Expected  map[string]interface{}
}

type TestResult struct {
	TestCase *TestCase
	Passed   bool
	Duration time.Duration
	Error    string
	Diff     string
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{}              `json:"data"`
	Errors []map[string]interface{} `json:"errors,omitempty"`
}

var (
	urlFlag        = flag.String("url", "http://localhost:8080/graphql", "GraphQL endpoint URL")
	parallelFlag   = flag.Bool("parallel", false, "Run tests in parallel")
	sequentialFlag = flag.Bool("sequential", true, "Run tests sequentially (default)")
	jwtSecret      = flag.String("jwt-secret", "your-secret-key-for-signing-wayfarer-jwts", "JWT signing secret")
	jwtIssuer      = flag.String("jwt-issuer", "wayfarer", "JWT issuer")
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gqltest [flags] <test-file-paths...>")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	testFiles := flag.Args()
	testCases, err := loadTestCases(testFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading test cases: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running %d tests...\n\n", len(testCases))

	var results []*TestResult
	if *parallelFlag {
		results = runTestsParallel(testCases)
	} else {
		results = runTestsSequential(testCases)
	}

	printResults(results)
}

func loadTestCases(filePaths []string) ([]*TestCase, error) {
	var testCases []*TestCase

	for _, path := range filePaths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if info.IsDir() {
			files, err := filepath.Glob(filepath.Join(path, "*.md"))
			if err != nil {
				return nil, fmt.Errorf("failed to glob directory %s: %w", path, err)
			}
			for _, file := range files {
				// Skip README and other documentation files
				baseName := strings.ToUpper(filepath.Base(file))
				if baseName == "README.MD" {
					continue
				}
				tc, err := parseTestCase(file)
				if err != nil {
					return nil, fmt.Errorf("failed to parse %s: %w", file, err)
				}
				testCases = append(testCases, tc)
			}
		} else {
			tc, err := parseTestCase(path)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", path, err)
			}
			testCases = append(testCases, tc)
		}
	}

	return testCases, nil
}

func parseTestCase(filePath string) (*TestCase, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tc := &TestCase{
		FilePath: filePath,
	}

	// Extract UserID from header
	userIDRegex := regexp.MustCompile(`(?m)^#\s*UserID:\s*(\S+)`)
	if matches := userIDRegex.FindSubmatch(content); len(matches) > 1 {
		tc.UserID = string(matches[1])
	} else {
		return nil, fmt.Errorf("UserID not found in %s", filePath)
	}

	// Extract Query section
	queryRegex := regexp.MustCompile(`(?s)##\s*Query\s*\n\s*` + "```" + `(?:graphql)?\s*\n(.*?)\n` + "```")
	if matches := queryRegex.FindSubmatch(content); len(matches) > 1 {
		tc.Query = strings.TrimSpace(string(matches[1]))
	} else {
		return nil, fmt.Errorf("Query section not found in %s", filePath)
	}

	// Extract Variables section (optional)
	variablesRegex := regexp.MustCompile(`(?s)##\s*Variables\s*\n\s*` + "```" + `(?:json)?\s*\n(.*?)\n` + "```")
	if matches := variablesRegex.FindSubmatch(content); len(matches) > 1 {
		variablesJSON := strings.TrimSpace(string(matches[1]))
		if err := json.Unmarshal([]byte(variablesJSON), &tc.Variables); err != nil {
			return nil, fmt.Errorf("failed to parse Variables in %s: %w", filePath, err)
		}
	}

	// Extract Expected section
	expectedRegex := regexp.MustCompile(`(?s)##\s*Expected\s*\n\s*` + "```" + `(?:json)?\s*\n(.*?)\n` + "```")
	if matches := expectedRegex.FindSubmatch(content); len(matches) > 1 {
		expectedJSON := strings.TrimSpace(string(matches[1]))
		if err := json.Unmarshal([]byte(expectedJSON), &tc.Expected); err != nil {
			return nil, fmt.Errorf("failed to parse Expected in %s: %w", filePath, err)
		}
	} else {
		return nil, fmt.Errorf("Expected section not found in %s", filePath)
	}

	return tc, nil
}

func generateToken(userID string) (string, error) {
	claims := WayfarerClaims{
		UserID:    userID,
		UserRoles: []string{"user", "admin"}, // Give full access for testing
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    *jwtIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(*jwtSecret))
}

func runTestsSequential(testCases []*TestCase) []*TestResult {
	results := make([]*TestResult, 0, len(testCases))
	for _, tc := range testCases {
		result := runTest(tc)
		results = append(results, result)
		printTestResult(result)
	}
	return results
}

func runTestsParallel(testCases []*TestCase) []*TestResult {
	resultChan := make(chan *TestResult, len(testCases))

	for _, tc := range testCases {
		go func(testCase *TestCase) {
			resultChan <- runTest(testCase)
		}(tc)
	}

	results := make([]*TestResult, 0, len(testCases))
	for i := 0; i < len(testCases); i++ {
		result := <-resultChan
		results = append(results, result)
		printTestResult(result)
	}

	return results
}

func runTest(tc *TestCase) *TestResult {
	result := &TestResult{
		TestCase: tc,
	}

	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()

	// Generate token
	token, err := generateToken(tc.UserID)
	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Failed to generate token: %v", err)
		return result
	}

	// Prepare GraphQL request
	reqBody := GraphQLRequest{
		Query:     tc.Query,
		Variables: tc.Variables,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Failed to marshal request: %v", err)
		return result
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", *urlFlag, bytes.NewBuffer(reqJSON))
	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		return result
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Failed to read response: %v", err)
		return result
	}

	// Parse response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Failed to parse response: %v", err)
		return result
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		errJSON, _ := json.MarshalIndent(gqlResp.Errors, "  ", "  ")
		result.Passed = false
		result.Error = fmt.Sprintf("GraphQL errors:\n  %s", string(errJSON))
		return result
	}

	// Build actual response structure matching expected format
	actualResponse := map[string]interface{}{
		"data": gqlResp.Data,
	}

	// Compare responses
	diff := cmp.Diff(tc.Expected, actualResponse)
	if diff != "" {
		result.Passed = false
		result.Diff = diff
	} else {
		result.Passed = true
	}

	return result
}

func printTestResult(result *TestResult) {
	fileName := filepath.Base(result.TestCase.FilePath)
	if result.Passed {
		fmt.Printf("✓ %s (%dms)\n", fileName, result.Duration.Milliseconds())
	} else {
		fmt.Printf("✗ %s (%dms)\n", fileName, result.Duration.Milliseconds())
		if result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
		}
		if result.Diff != "" {
			fmt.Printf("  Difference:\n")
			for _, line := range strings.Split(result.Diff, "\n") {
				if line != "" {
					fmt.Printf("    %s\n", line)
				}
			}
		}
	}
}

func printResults(results []*TestResult) {
	fmt.Println()

	passed := 0
	failed := 0
	durations := make([]time.Duration, len(results))

	for i, result := range results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
		durations[i] = result.Duration
	}

	// Sort durations for percentile calculations
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	// Calculate statistics
	var total time.Duration
	for _, d := range durations {
		total += d
	}

	min := durations[0]
	max := durations[len(durations)-1]
	avg := total / time.Duration(len(durations))
	p50 := durations[len(durations)*50/100]
	p95 := durations[len(durations)*95/100]
	p99 := durations[len(durations)*99/100]

	fmt.Printf("Results: %d passed, %d failed\n", passed, failed)
	fmt.Printf("Timing: min=%dms, max=%dms, avg=%dms, p50=%dms, p95=%dms, p99=%dms\n",
		min.Milliseconds(), max.Milliseconds(), avg.Milliseconds(),
		p50.Milliseconds(), p95.Milliseconds(), p99.Milliseconds())

	if failed > 0 {
		os.Exit(1)
	}
}
