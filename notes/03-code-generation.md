# Code Generation

This document describes the code generation setup for GraphQL resolvers and SQL queries.

## GraphQL Generation (gqlgen)

### Overview

We use gqlgen to generate type-safe GraphQL server code from schema files. The system has **three separate GraphQL APIs**, each with its own configuration and generated code:

1. **User API** - For end users (mobile/web apps)
2. **Admin API** - For system administrators
3. **M2M API** - For machine-to-machine communication with external systems

### Schema Files

All schemas are located in `/gql/` at the project root:

- `shared.graphqls` - Common types, enums, and interfaces shared across all three APIs
- `user.graphqls` - User-facing query and mutation roots
- `admin.graphqls` - Admin query and mutation roots
- `m2m.graphqls` - Machine-to-machine query and mutation roots

### Configuration Files

Each API has its own gqlgen configuration file in `backend/`:

- `gqlgen.yml` - User API configuration
- `gqlgen.admin.yml` - Admin API configuration
- `gqlgen.m2m.yml` - M2M API configuration

### Generated Code Structure

```
backend/internal/graph/
├── scalars/
│   └── scalars.go              # Custom scalar implementations (DateTime, Date, HTML)
├── user/
│   ├── generated.go            # Generated execution engine (User API)
│   ├── *.resolvers.go          # Resolver implementations (User API)
│   └── model/
│       ├── doc.go
│       └── models_gen.go       # Generated model types (User API)
├── admin/
│   ├── generated.go            # Generated execution engine (Admin API)
│   ├── *.resolvers.go          # Resolver implementations (Admin API)
│   └── model/
│       ├── doc.go
│       └── models_gen.go       # Generated model types (Admin API)
└── m2m/
    ├── generated.go            # Generated execution engine (M2M API)
    ├── *.resolvers.go          # Resolver implementations (M2M API)
    └── model/
        ├── doc.go
        └── models_gen.go       # Generated model types (M2M API)
```

### Custom Scalars

We have implemented three custom scalar types in `internal/graph/scalars/`:

1. **DateTime** - RFC3339 formatted datetime (e.g., `2025-10-31T14:30:00Z`)
2. **Date** - YYYY-MM-DD formatted date (e.g., `2025-10-31`)
3. **HTML** - String type for sanitized HTML content

All scalars implement the `graphql.Marshaler` and `graphql.Unmarshaler` interfaces.

### Generating Code

To regenerate GraphQL code after schema changes:

```bash
# Generate all three APIs
make generate

# Or generate individually
make generate-user
make generate-admin
make generate-m2m
```

Alternatively, run gqlgen directly:

```bash
go run github.com/99designs/gqlgen generate --config gqlgen.yml        # User API
go run github.com/99designs/gqlgen generate --config gqlgen.admin.yml  # Admin API
go run github.com/99designs/gqlgen generate --config gqlgen.m2m.yml    # M2M API
```

### Important Notes

1. **Do not modify generated files** - Files ending in `generated.go` and `models_gen.go` are auto-generated and will be overwritten
2. **Resolver implementations** - Files ending in `.resolvers.go` are generated once but preserved on subsequent runs - these are where you implement business logic
3. **Schema validation** - gqlgen validates schemas during generation and will fail if there are syntax errors or type mismatches
4. **GitHub PR Reviews** - Generated files are collapsed in PR reviews via `.gitattributes` configuration

### Common Issues

**Issue**: `Cannot have multiple schema entry points`
- **Cause**: Multiple schema definitions in different files
- **Solution**: Each API must be generated separately with its own config file

**Issue**: `OBJECT field must be one of SCALAR, ENUM, INPUT_OBJECT`
- **Cause**: Using an output type (type) in an input context (input)
- **Solution**: Create separate input types for use in mutations and filters

**Issue**: `unable to load package`
- **Cause**: Target package doesn't exist yet
- **Solution**: Create empty package with `package model` before running gqlgen

## SQL Generation (SQLc)

### Overview

SQLc will generate type-safe Go code from SQL queries. This is not yet configured but will follow a similar pattern to gqlgen.

### Configuration (Planned)

- Configuration file: `backend/sqlc.yaml`
- SQL queries: `backend/internal/database/queries/`
- Generated code: `backend/internal/database/sqlc/`

### Generating Code (Planned)

```bash
# Will be added to Makefile
make generate-sql
```

## Workflow

1. Modify GraphQL schema files in `/gql/`
2. Run `make generate` to regenerate code
3. Implement resolver logic in `.resolvers.go` files
4. Test the API
5. Commit both schema and generated code changes

## Build Integration

The Makefile includes targets for common development tasks:

- `make generate` - Generate all GraphQL code
- `make test` - Run tests
- `make build` - Build binaries
- `make fmt` - Format code
- `make tidy` - Tidy Go modules
