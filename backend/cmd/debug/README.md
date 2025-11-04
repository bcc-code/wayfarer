# Wayfarer Debug TUI Tool

A Terminal User Interface (TUI) tool built with [bubbletea](https://github.com/charmbracelet/bubbletea) for performing debug operations that may not be available via APIs.

## Features

- **Add Church**: Create new churches with auto-generated or custom data
- **Add User**: Create new users with auto-generated or custom data
- **Assign User Role**: Assign roles to users with proper scope validation

## Installation

Build the tool from the backend directory:

```bash
make build
```

This builds all binaries including the debug tool to `bin/debug`.

Or build just the debug tool:

```bash
go build -o bin/debug ./cmd/debug
```

## Usage

### Prerequisites

- Database must be running and accessible via `DATABASE_URL` environment variable
- `.env` file should be configured with database connection string

### Running the Tool

**Using Make (recommended):**
```bash
make debug
```

**Or run directly:**
```bash
./bin/debug
```

**Or run without building:**
```bash
go run ./cmd/debug
```

### Navigation

- **Arrow keys**: Navigate up/down/left/right through menu items, form fields, and options
  - **Up/Down**: Move between form fields or menu items
  - **Left/Right**: Change selection in dropdown fields (gender, category, church, etc.)
- **Enter**:
  - On relation fields (user, church, project, team): Open popup table view
  - On Submit button: Submit the form
  - In popup: Select the highlighted item
- **Type**:
  - In forms: Enter text in text input fields (all letters work normally)
  - In popups: Filter the table in real-time
- **Backspace**: Delete characters in text input fields or popup filters
- **Ctrl+D**: Clear filter (in popups or when filtering users)
- **Esc**:
  - In popup: Close popup without selection
  - In form: Cancel and return to main menu
- **q** or **Ctrl+C**: Quit the application

## Popup Table View

When you press **Enter** on any field with relations (users, churches, projects, teams), a popup table view opens with:

**Features:**
- **Filterable table**: Type to search across all columns in real-time
- **Clear columns**: Displays relevant information (Name, Email, ID, etc.)
- **Large lists**: Automatically scrolls and shows position (e.g., "Showing 1-15 of 75 rows")
- **Quick navigation**: Arrow keys to move, Enter to select, Esc to cancel
- **Filter stats**: Shows match count (e.g., "Filter: 'john' (3/75 matches)")

**Example workflow:**
1. Navigate to "User" field in role assignment
2. Press **Enter** to open user selection popup
3. Type "john" to filter users with "john" in name or email
4. Use arrow keys to select the right user
5. Press **Enter** to select and close popup
6. Continue with the form

This makes it much easier to work with large datasets and find the exact record you need!

## Features in Detail

### Add Church

Creates a new church with:
- **Name**: Auto-generated from faker (company name + "Church") or custom input
- **Country**: Auto-generated from faker or custom input
- **Category**: Select from S (Small), L (Large), or XL (Extra Large)

All empty fields are automatically filled with realistic fake data using the faker library.

### Add User

Creates a new user with:
- **Name**: Auto-generated based on gender or custom input
- **Email**: Auto-generated or custom input
- **Members ID**: Auto-generated (format: MEM-#####) or custom input
- **Gender**: Select MALE or FEMALE
- **Church**: Select from existing churches (loaded from database)
- **Avatar URL**: Auto-generated (pravatar.cc) or custom input
- **Birthdate**: Auto-generated (age 13-80)

The tool loads existing churches from the database. If no churches exist, you must add a church first.

### Assign User Role

Assigns a role to a user with proper scope validation:

**Available Roles:**
- **SUPERADMIN**: Global admin (no scope)
- **ADMIN**: Global admin (no scope)
- **CHURCH_ADMIN**: Requires church scope
- **PROJECT_ADMIN**: Requires project scope
- **TEAM_LEAD**: Requires team scope
- **USER**: Standard user (no scope)
- **M2M**: Machine-to-machine (no scope)

**User Search/Filter:**
- When on the User or Assigned By field, simply start typing to filter users by name or email
- The filter is case-insensitive and searches across the full user display string
- Press **Ctrl+D** to clear the filter
- The display shows the number of matches (e.g., "Filter: 'john' (3/75 matches)")

**Scope Validation:**
- Global roles (SUPERADMIN, ADMIN, USER, M2M) must have no scope
- CHURCH_ADMIN must have exactly one church scope
- PROJECT_ADMIN must have exactly one project scope
- TEAM_LEAD must have exactly one team scope

The form provides helpful hints about required scopes for each role.

## Implementation Details

### Data Generation

The tool uses the same faker library (`github.com/jaswdr/faker`) as the seed tool to generate realistic fake data:
- Names based on gender
- Valid email addresses
- Realistic countries
- Avatar URLs from pravatar.cc

### ID Generation

All IDs use the same ULID-based system as the main application:
- Churches: `CH` + 26-character ULID
- Users: `US` + 26-character ULID
- User Roles: `UR` + 26-character ULID

### Database Connection

The tool uses the same configuration system as the main application:
- Reads from `.env` file
- Uses `DATABASE_URL` environment variable
- Connects using pgx connection pool

## Error Handling

The tool provides clear error messages for:
- Database connection failures
- Invalid data
- Missing required fields
- Scope validation errors
- Duplicate entries

Errors are displayed in red at the top of the screen.

## Future Enhancements

Potential additions:
- View/search existing records
- Edit existing records
- Delete records
- Bulk operations
- Export data
- Import data from files
