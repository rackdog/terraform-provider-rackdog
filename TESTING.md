# Terraform Provider Rackdog - Testing Guide

This document describes the testing setup and how to run tests for the Rackdog Terraform provider.

## Test Suite Overview

The test suite includes:

### 1. **Unit Tests** - Mock API tests
- **Client Tests** (`client_test.go`) - HTTP client functionality
  - API key authentication
  - Request/response handling
  - Error handling
  - All CRUD operations (Create, Read, Delete)
  - RAID configuration checking

- **Data Source Tests** - Data source functionality
  - `data_source_plans_test.go` - Plans data source
  - `data_source_operating_systems_test.go` - OS data source

- **Provider Tests** (`provider_test.go`) - Provider configuration
  - Configuration validation
  - Environment variable handling
  - Resource and data source registration

### 2. **Acceptance Tests** - Real API integration tests
- **Server Resource Tests** (`resource_server_acc_test.go`)
- Requires real API credentials
- Run with `make acc`

## Running Tests

### Quick Start

```bash
# Run all unit tests
make test

# Run unit tests with coverage
make test-coverage

# Run full test suite (formatting + linting + tests + coverage)
make test-all

# View coverage in browser
make coverage-html
```

### Detailed Commands

#### Basic Unit Tests
```bash
# Run all tests with verbose output
go test ./... -v

# Run tests in a specific package
go test ./internal/provider -v

# Run a specific test
go test ./internal/provider -run TestListOperatingSystems -v
```

#### Coverage Reports
```bash
# Generate coverage report
make test-coverage

# Generate HTML coverage report and open in browser
make coverage-html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

#### Acceptance Tests
```bash
# Set up environment
export RACKDOG_API_KEY="your-api-key"
export RACKDOG_ENDPOINT="https://metal.rackdog.com"

# Run acceptance tests
make acc

# Or manually
TF_ACC=1 go test ./... -v -timeout=30m
```

#### Watch Mode (for development)
```bash
# Requires entr: brew install entr (macOS) or apt install entr (Linux)
make test-watch
```

## Test Coverage

The test suite covers:

### Client Package (`client.go`)
- ✅ `NewClient` - Client initialization
- ✅ `ListOperatingSystems` - Fetch OS list
- ✅ `ListPlans` - Fetch plans (with/without location filter)
- ✅ `CreateServer` - Server provisioning
- ✅ `GetServer` - Server retrieval
- ✅ `DeleteServer` - Server destruction
- ✅ `CheckRaid` - RAID configuration validation
- ✅ HTTP error handling
- ✅ API key header injection

### Data Sources
- ✅ Plans data source schema
- ✅ Plans data source metadata
- ✅ Plans data source read operation
- ✅ Operating systems data source schema
- ✅ Operating systems data source metadata
- ✅ Operating systems data source configure
- ✅ Operating systems data source read (success/error cases)
- ✅ Error handling for unconfigured providers

### Provider
- ✅ Provider metadata
- ✅ Provider schema
- ✅ Provider configuration with HCL
- ✅ Provider configuration with environment variables
- ✅ API key validation
- ✅ Resource registration
- ✅ Data source registration

## Test Structure

### Mock HTTP Servers
Tests use `httptest.NewServer` to mock the Rackdog API:

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Verify request
    if r.URL.Path != "/v1/ordering/os" {
        t.Fatalf("unexpected path: %s", r.URL.Path)
    }

    // Return mock response
    json.NewEncoder(w).Encode(map[string]any{
        "success": true,
        "data": []map[string]any{...},
    })
}))
defer srv.Close()
```

### Test Patterns
- **Table-driven tests** for multiple scenarios
- **Subtests** with `t.Run()` for organization
- **Explicit cleanup** with `defer` statements
- **Detailed error messages** for debugging

## Adding New Tests

### 1. Client Tests
Add to `client_test.go`:

```go
func TestNewEndpoint(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock API response
    }))
    defer srv.Close()

    c := NewClient(srv.URL, "test-key")
    result, err := c.NewMethod(context.Background())

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Add assertions
}
```

### 2. Data Source Tests
Create new file `data_source_<name>_test.go`:

```go
func TestNewDataSource_Schema(t *testing.T) {
    ds := NewMyDataSource()
    // Test schema
}

func TestNewDataSource_Read(t *testing.T) {
    // Test read operation
}
```

### 3. Resource Tests
For acceptance tests, add to `resource_<name>_acc_test.go`:

```go
func TestAccServerResource_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{...},
    })
}
```

## Continuous Integration

The test suite is designed for CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run tests
  run: make test-all

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

## Troubleshooting

### Tests Failing

1. **Check formatting**: `make fmt`
2. **Run single test**: `go test ./internal/provider -run TestName -v`
3. **Check test output**: Look for detailed error messages
4. **Verify mock data**: Ensure test data matches actual API responses

### Coverage Not Generating

```bash
# Clean old coverage files
make clean-test

# Regenerate
make test-coverage
```

### Acceptance Tests Timing Out

```bash
# Increase timeout
TF_ACC=1 go test ./... -v -timeout=60m
```

## Best Practices

1. **Always run `make test-all` before committing**
2. **Aim for >80% code coverage**
3. **Write table-driven tests for multiple scenarios**
4. **Use descriptive test names**: `Test<Package>_<Function>_<Scenario>`
5. **Mock external dependencies** (use `httptest` for HTTP)
6. **Clean up resources** with `defer` statements
7. **Test error cases** as well as success cases

## Coverage Goals

| Package | Current | Target |
|---------|---------|--------|
| Client  | ~90%    | >85%   |
| Provider| ~75%    | >70%   |
| Data Sources | ~80% | >75%  |
| Resources | TBD   | >70%   |

## Resources

- [Terraform Plugin Testing](https://developer.hashicorp.com/terraform/plugin/testing)
- [Go Testing](https://golang.org/pkg/testing/)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
