package treeformatter

import (
	"testing"
)

// TestBoxDrawingCharacters tests all box drawing character combinations
func TestBoxDrawingCharacters(t *testing.T) {
	formatter := NewDefaultFormatter()

	tests := []struct {
		name           string
		isLast         bool
		depth          int
		parentPrefix   string
		expectedPrefix string
	}{
		{
			name:           "root level",
			isLast:         true,
			depth:          0,
			parentPrefix:   "",
			expectedPrefix: "",
		},
		{
			name:           "first level last child",
			isLast:         true,
			depth:          1,
			parentPrefix:   "",
			expectedPrefix: "└── ",
		},
		{
			name:           "first level middle child",
			isLast:         false,
			depth:          1,
			parentPrefix:   "",
			expectedPrefix: "├── ",
		},
		{
			name:           "second level last child",
			isLast:         true,
			depth:          2,
			parentPrefix:   "│  ",
			expectedPrefix: "│  └── ",
		},
		{
			name:           "second level middle child",
			isLast:         false,
			depth:          2,
			parentPrefix:   "│  ",
			expectedPrefix: "│  ├── ",
		},
		{
			name:           "third level last child",
			isLast:         true,
			depth:          3,
			parentPrefix:   "│  ├── ",
			expectedPrefix: "│  ├── └── ",
		},
		{
			name:           "third level middle child",
			isLast:         false,
			depth:          3,
			parentPrefix:   "│  ├── ",
			expectedPrefix: "│  ├── ├── ",
		},
		{
			name:           "nested structure deep last",
			isLast:         true,
			depth:          3,
			parentPrefix:   "│  │  ",
			expectedPrefix: "│  │  └── ",
		},
		{
			name:           "nested structure deep middle",
			isLast:         false,
			depth:          3,
			parentPrefix:   "│  │  ",
			expectedPrefix: "│  │  ├── ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.GetTreePrefix(tt.isLast, tt.depth, tt.parentPrefix)
			if result != tt.expectedPrefix {
				t.Errorf("Expected %q, got %q", tt.expectedPrefix, result)
			}
		})
	}
}

// TestGetParentPrefix tests the parent prefix calculation for nested structures
func TestGetParentPrefix(t *testing.T) {
	formatter := NewDefaultFormatter()

	tests := []struct {
		name           string
		prefix         string
		isLast         bool
		expectedPrefix string
	}{
		{
			name:           "last child prefix",
			prefix:         "└── ",
			isLast:         true,
			expectedPrefix: "   ",
		},
		{
			name:           "middle child prefix",
			prefix:         "├── ",
			isLast:         false,
			expectedPrefix: "│  ",
		},
		{
			name:           "nested last child",
			prefix:         "│  └── ",
			isLast:         true,
			expectedPrefix: "│     ",
		},
		{
			name:           "nested middle child",
			prefix:         "│  ├── ",
			isLast:         false,
			expectedPrefix: "│  │  ",
		},
		{
			name:           "deep nested last",
			prefix:         "│  │  └── ",
			isLast:         true,
			expectedPrefix: "│  │     ",
		},
		{
			name:           "deep nested middle",
			prefix:         "│  │  ├── ",
			isLast:         false,
			expectedPrefix: "│  │  │  ",
		},
		{
			name:           "empty prefix",
			prefix:         "",
			isLast:         true,
			expectedPrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.GetParentPrefix(tt.prefix, tt.isLast)
			if result != tt.expectedPrefix {
				t.Errorf("Expected %q, got %q", tt.expectedPrefix, result)
			}
		})
	}
}

// TestComplexTreeStructure tests a realistic tree structure
func TestComplexTreeStructure(t *testing.T) {
	formatter := NewDefaultFormatter()
	formatter.config.ShowEmojis = true

	// Simulate building a complex tree
	rootPrefix := formatter.GetTreePrefix(true, 0, "")
	child1Prefix := formatter.GetTreePrefix(false, 1, rootPrefix)
	child2Prefix := formatter.GetTreePrefix(false, 1, rootPrefix)
	child3Prefix := formatter.GetTreePrefix(true, 1, rootPrefix)

	// Test child prefixes
	if child1Prefix != "├── " {
		t.Errorf("Expected child1 prefix '├── ', got %q", child1Prefix)
	}

	if child2Prefix != "├── " {
		t.Errorf("Expected child2 prefix '├── ', got %q", child2Prefix)
	}

	if child3Prefix != "└── " {
		t.Errorf("Expected child3 prefix '└── ', got %q", child3Prefix)
	}

	// Test nested prefixes
	child1ParentPrefix := formatter.GetParentPrefix(child1Prefix, false)
	child1_1Prefix := formatter.GetTreePrefix(false, 2, child1ParentPrefix)
	child1_2Prefix := formatter.GetTreePrefix(true, 2, child1ParentPrefix)

	if child1_1Prefix != "│  ├── " {
		t.Errorf("Expected child1_1 prefix '│  ├── ', got %q", child1_1Prefix)
	}

	if child1_2Prefix != "│  └── " {
		t.Errorf("Expected child1_2 prefix '│  └── ', got %q", child1_2Prefix)
	}
}