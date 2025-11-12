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
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type WayfarerClaims struct {
	UserID    string   `json:"user_id"`
	UserRoles []string `json:"user_roles"`
	jwt.RegisteredClaims
}

type TestCase struct {
	FilePath  string
	Content   string
	UserID    string
	Roles     []string
	Query     string
	Variables map[string]interface{}
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
	urlFlag   = flag.String("url", "http://localhost:8080/graphql", "GraphQL endpoint URL")
	jwtSecret = flag.String("jwt-secret", "your-secret-key-for-signing-wayfarer-jwts", "JWT signing secret")
	jwtIssuer = flag.String("jwt-issuer", "wayfarer", "JWT issuer")
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gqltest-update [flags] <test-file-paths...>")
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

	fmt.Printf("Updating %d test files...\n\n", len(testCases))

	updated := 0
	failed := 0

	for _, tc := range testCases {
		fmt.Printf("Processing %s...", filepath.Base(tc.FilePath))

		if err := updateTestCase(tc); err != nil {
			fmt.Printf(" FAILED: %v\n", err)
			failed++
		} else {
			fmt.Printf(" OK\n")
			updated++
		}
	}

	fmt.Printf("\nResults: %d updated, %d failed\n", updated, failed)

	if failed > 0 {
		os.Exit(1)
	}
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
		Content:  string(content),
	}

	// Extract UserID from header
	userIDRegex := regexp.MustCompile(`(?m)^#\s*UserID:\s*(\S+)`)
	if matches := userIDRegex.FindStringSubmatch(tc.Content); len(matches) > 1 {
		tc.UserID = matches[1]
	} else {
		return nil, fmt.Errorf("UserID not found in %s", filePath)
	}

	// Extract Query section
	queryRegex := regexp.MustCompile(`(?s)##\s*Query\s*\n\s*` + "```" + `(?:graphql)?\s*\n(.*?)\n` + "```")
	if matches := queryRegex.FindStringSubmatch(tc.Content); len(matches) > 1 {
		tc.Query = strings.TrimSpace(matches[1])
	} else {
		return nil, fmt.Errorf("Query section not found in %s", filePath)
	}

	// Extract Variables section (optional)
	variablesRegex := regexp.MustCompile(`(?s)##\s*Variables\s*\n\s*` + "```" + `(?:json)?\s*\n(.*?)\n` + "```")
	if matches := variablesRegex.FindStringSubmatch(tc.Content); len(matches) > 1 {
		variablesJSON := strings.TrimSpace(matches[1])
		if err := json.Unmarshal([]byte(variablesJSON), &tc.Variables); err != nil {
			return nil, fmt.Errorf("failed to parse Variables in %s: %w", filePath, err)
		}
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

func updateTestCase(tc *TestCase) error {
	// Generate token
	token, err := generateToken(tc.UserID)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Prepare GraphQL request
	reqBody := GraphQLRequest{
		Query:     tc.Query,
		Variables: tc.Variables,
	}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", *urlFlag, bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		errJSON, _ := json.MarshalIndent(gqlResp.Errors, "", "  ")
		return fmt.Errorf("GraphQL errors:\n%s", string(errJSON))
	}

	// Build actual response structure
	actualResponse := map[string]interface{}{
		"data": gqlResp.Data,
	}

	// Marshal to pretty JSON
	prettyJSON, err := json.MarshalIndent(actualResponse, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Update the Expected section in the file
	expectedRegex := regexp.MustCompile(`(?s)(##\s*Expected\s*\n\s*` + "```" + `(?:json)?\s*\n)(.*?)(\n` + "```)")
	newContent := expectedRegex.ReplaceAllString(tc.Content, "${1}"+string(prettyJSON)+"${3}")

	// Write back to file
	if err := os.WriteFile(tc.FilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
