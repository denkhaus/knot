# Mocks Package

This package contains generated GoMock mocks for testing purposes.

## Generated Mocks

### MockProjectManager
A mock implementation of the `manager.ProjectManager` interface for use in unit tests.

## Usage

```go
import (
    "testing"
    "go.uber.org/mock/gomock"

    "github.com/denkhaus/knot/v2/internal/mocks"
    "github.com/denkhaus/knot/v2/internal/shared"
)

func TestSomething(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockMgr := mocks.NewMockProjectManager(ctrl)

    // Setup expectations
    mockMgr.EXPECT().
        ListProjects(gomock.Any()).
        Return([]*types.Project{}, nil)

    // Use mock in test
    appCtx := &shared.AppContext{
        ProjectManager: mockMgr,
        // ...
    }

    // Test code here
}
```

## Regenerating Mocks

To regenerate the mocks after interface changes:

```bash
mockgen -source=internal/manager/interfaces.go -destination=internal/mocks/mock_project_manager.go -package=mocks ProjectManager
```

## Why Central Mocks?

1. **DRY Principle**: Avoid duplicating mock definitions across test files
2. **Consistency**: Ensure all tests use the same mock implementation
3. **Maintenance**: Single location to update when interfaces change
4. **Standardization**: Follow Go testing best practices with GoMock