# Debug Tool - User Search and Token Generation

## Current State

The debug tool (`backend/cmd/debug/`) is a terminal UI (TUI) application built with bubbletea that provides:
- Add Church functionality
- Add User functionality
- Assign User Role functionality

## Token Generation

There's a separate tool `backend/cmd/gentoken/main.go` that generates JWT tokens for users:
- Takes a user ID as command line argument
- Uses hardcoded secret: `your-secret-key-for-signing-wayfarer-jwts`
- Generates a JWT with:
  - UserID claim
  - UserRole claim (hardcoded to "user")
  - 24-hour expiration
  - Uses HS256 signing method

## Implementation

Added a new menu option "Search User & Generate Token" that:
1. Shows a search interface to find users by name, email, or members_id
2. Displays search results in a table popup (similar to church selection)
3. When user selects a user, generates a JWT token
4. Token is automatically copied to clipboard using pbcopy
5. Token is also displayed on screen for verification

### Files Modified:
1. `backend/cmd/debug/model.go` - Added `ScreenSearchUserToken` screen type and `tokenForm` field
2. `backend/cmd/debug/token_form.go` (new) - Created token generation screen with search and clipboard integration

### Key Features:
- User search by name/email/members_id with ILIKE pattern matching
- Returns up to 50 matching users
- Table popup for results display with columns: Name, Email, Members ID, User ID
- JWT token generation using same logic and secret as gentoken
- Automatic clipboard copy using pbcopy (macOS)
- Token valid for 24 hours
- Token claims include: user_id, user_role (hardcoded to "user"), issuer, issued_at, expires_at

### Usage:
1. Run the debug tool: `./bin/debug`
2. Select "Search User & Generate Token" from the menu
3. Type search query (name, email, or members_id)
4. Press Enter to search
5. Select a user from the popup table
6. Token is generated and automatically copied to clipboard
7. Use the token in API requests by pasting from clipboard

### Implementation Notes:
- Uses `showPopupMsg` custom message type to display search results
- Integrates with existing TablePopup component for consistent UX
- Error handling for search failures and token generation errors
- Success message confirms token generation and clipboard copy
