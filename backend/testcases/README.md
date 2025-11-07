# GraphQL Test Cases

This directory contains test cases for the Wayfarer GraphQL API. Each test case is a markdown file that defines a GraphQL query, expected response, and metadata.

## Running Tests

Build and run the test tool:

```bash
# Build the tool
cd backend
go build -o bin/gqltest ./cmd/gqltest

# Run a single test
./bin/gqltest testcases/example_me.md

# Run multiple tests
./bin/gqltest testcases/example_me.md testcases/example_projects.md

# Run all tests in a directory
./bin/gqltest testcases/

# Run tests against a different server
./bin/gqltest --url https://staging.example.com/graphql testcases/

# Run tests in parallel
./bin/gqltest --parallel testcases/
```

### Command-line Flags

- `--url` - GraphQL endpoint URL (default: `http://localhost:8080/graphql`)
- `--parallel` - Run tests concurrently for faster execution
- `--sequential` - Run tests one at a time (default)
- `--jwt-secret` - JWT signing secret (default: dev secret)
- `--jwt-issuer` - JWT issuer claim (default: `wayfarer`)

## Test File Format

Test cases are markdown files with specific sections. Here's the structure:

```markdown
# UserID: US01K9DGS18D92WBMV3X7ETHNPMN

## Description
Optional description or notes about this test case.
You can write anything here - it's for documentation purposes.

## Query
```graphql
query {
  me {
    id
    name
  }
}
```

## Variables
```json
{
  "projectId": "PR01ARZ3NDEKTSV4RRFFQ69G5FAV"
}
```

## Expected
```json
{
  "data": {
    "me": {
      "id": "US01K9DGS18D92WBMV3X7ETHNPMN",
      "name": "Test User"
    }
  }
}
```

## Notes
Additional notes can go in any section.
The tool only looks for: UserID header, Query, Variables, and Expected sections.
```

### Required Sections

1. **UserID Header** (required)
   - First line: `# UserID: <user-id>`
   - The test tool generates a JWT token for this user
   - Example: `# UserID: US01K9DGS18D92WBMV3X7ETHNPMN`

2. **Query Section** (required)
   - Section header: `## Query`
   - GraphQL query in a code block with optional `graphql` language tag
   - Can be a query, mutation, or subscription

3. **Expected Section** (required)
   - Section header: `## Expected`
   - JSON response in a code block with optional `json` language tag
   - Must match the structure: `{ "data": { ... } }`
   - The tool performs deep JSON comparison (order-independent)

### Optional Sections

1. **Variables Section** (optional)
   - Section header: `## Variables`
   - JSON object in a code block with optional `json` language tag
   - Variables passed to the GraphQL query

2. **Description, Notes, etc.** (optional)
   - Any other sections are ignored by the tool
   - Use these for documentation and context

## How It Works

1. **Token Generation**: For each test, the tool automatically generates a JWT token for the specified UserID
   - Tokens include both `user` and `admin` roles for maximum test coverage
   - Tokens are valid for 24 hours
   - Secret and issuer can be customized via flags

2. **Request Execution**: The tool makes a POST request to the GraphQL endpoint
   - Sets `Authorization: Bearer <token>` header
   - Sends the query and variables as JSON
   - Records timing information

3. **Response Comparison**: The actual response is compared to the expected response
   - Deep JSON comparison (structure and values)
   - Order-independent for arrays and objects
   - GraphQL errors are reported separately

4. **Results Reporting**: The tool outputs:
   - Pass/fail status for each test
   - Timing per test (milliseconds)
   - Detailed diff for failed tests
   - Summary statistics: min, max, avg, p50, p95, p99

## Example Output

```
Running 3 tests...

✓ example_me.md (123ms)
✓ example_projects.md (456ms)
✗ example_project_with_variables.md (234ms)
  Difference:
      {
        "data": {
          "project": {
            "name": string(
    -            "Summer Bible Study 2024",
    +            "Winter Bible Study 2024",
            ),
          },
        },
      }

Results: 2 passed, 1 failed
Timing: min=123ms, max=456ms, avg=271ms, p50=234ms, p95=456ms, p99=456ms
```

## Tips

1. **User IDs**: Use real user IDs from your database. The tool generates tokens but doesn't create users.

2. **Expected Responses**: The expected response should match exactly what the server returns. Use the GraphQL Playground to get the exact format.

3. **Comments**: Add Description, Notes, or other custom sections to document test intent, edge cases, or prerequisites.

4. **Variables**: Use the Variables section for parameterized queries. This is cleaner than hardcoding values in the query string.

5. **Debugging**: Run with `--sequential` (default) to see test results immediately. Use `--parallel` for faster CI runs.

6. **Server Setup**: Make sure the Wayfarer server is running with a valid database before running tests.

## Creating New Tests

1. Copy an existing test file as a template
2. Update the UserID to a valid user in your database
3. Write your GraphQL query
4. Run the test to see the actual response
5. Copy the actual response to the Expected section
6. Verify the test passes

## CI Integration

The tool exits with status code 1 if any tests fail, making it suitable for CI pipelines:

```bash
# In CI
./bin/gqltest testcases/ || exit 1
```
