package db

import (
	"reflect"
	"testing"
)

func TestParseCQLType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *CQLTypeInfo
		wantErr  bool
	}{
		// Primitive types
		{
			name:  "simple text type",
			input: "text",
			expected: &CQLTypeInfo{
				BaseType: "text",
			},
		},
		{
			name:  "simple int type",
			input: "int",
			expected: &CQLTypeInfo{
				BaseType: "int",
			},
		},
		{
			name:  "uuid type",
			input: "uuid",
			expected: &CQLTypeInfo{
				BaseType: "uuid",
			},
		},
		{
			name:  "timestamp type",
			input: "timestamp",
			expected: &CQLTypeInfo{
				BaseType: "timestamp",
			},
		},

		// Frozen types
		{
			name:  "frozen text",
			input: "frozen<text>",
			expected: &CQLTypeInfo{
				BaseType: "text",
				Frozen:   true,
			},
		},

		// List types
		{
			name:  "list of text",
			input: "list<text>",
			expected: &CQLTypeInfo{
				BaseType: "list",
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
				},
			},
		},
		{
			name:  "frozen list of int",
			input: "frozen<list<int>>",
			expected: &CQLTypeInfo{
				BaseType: "list",
				Frozen:   true,
				Parameters: []*CQLTypeInfo{
					{BaseType: "int"},
				},
			},
		},
		{
			name:  "list of frozen text",
			input: "list<frozen<text>>",
			expected: &CQLTypeInfo{
				BaseType: "list",
				Parameters: []*CQLTypeInfo{
					{BaseType: "text", Frozen: true},
				},
			},
		},

		// Set types
		{
			name:  "set of uuid",
			input: "set<uuid>",
			expected: &CQLTypeInfo{
				BaseType: "set",
				Parameters: []*CQLTypeInfo{
					{BaseType: "uuid"},
				},
			},
		},

		// Map types
		{
			name:  "map of text to int",
			input: "map<text, int>",
			expected: &CQLTypeInfo{
				BaseType: "map",
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
					{BaseType: "int"},
				},
			},
		},
		{
			name:  "frozen map",
			input: "frozen<map<text, int>>",
			expected: &CQLTypeInfo{
				BaseType: "map",
				Frozen:   true,
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
					{BaseType: "int"},
				},
			},
		},

		// Tuple types
		{
			name:  "simple tuple",
			input: "tuple<int, text, uuid>",
			expected: &CQLTypeInfo{
				BaseType: "tuple",
				Parameters: []*CQLTypeInfo{
					{BaseType: "int"},
					{BaseType: "text"},
					{BaseType: "uuid"},
				},
			},
		},
		{
			name:  "frozen tuple",
			input: "frozen<tuple<text, int>>",
			expected: &CQLTypeInfo{
				BaseType: "tuple",
				Frozen:   true,
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
					{BaseType: "int"},
				},
			},
		},

		// Nested collections
		{
			name:  "list of lists",
			input: "list<list<text>>",
			expected: &CQLTypeInfo{
				BaseType: "list",
				Parameters: []*CQLTypeInfo{
					{
						BaseType: "list",
						Parameters: []*CQLTypeInfo{
							{BaseType: "text"},
						},
					},
				},
			},
		},
		{
			name:  "map of text to list of int",
			input: "map<text, list<int>>",
			expected: &CQLTypeInfo{
				BaseType: "map",
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
					{
						BaseType: "list",
						Parameters: []*CQLTypeInfo{
							{BaseType: "int"},
						},
					},
				},
			},
		},
		{
			name:  "complex nested frozen",
			input: "frozen<map<text, frozen<list<frozen<set<uuid>>>>>>",
			expected: &CQLTypeInfo{
				BaseType: "map",
				Frozen:   true,
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
					{
						BaseType: "list",
						Frozen:   true,
						Parameters: []*CQLTypeInfo{
							{
								BaseType: "set",
								Frozen:   true,
								Parameters: []*CQLTypeInfo{
									{BaseType: "uuid"},
								},
							},
						},
					},
				},
			},
		},

		// UDT types
		{
			name:  "simple UDT",
			input: "address",
			expected: &CQLTypeInfo{
				BaseType: "udt",
				UDTName:  "address",
			},
		},
		{
			name:  "keyspace qualified UDT",
			input: "myks.address",
			expected: &CQLTypeInfo{
				BaseType: "udt",
				UDTName:  "address",
				Keyspace: "myks",
			},
		},
		{
			name:  "frozen UDT",
			input: "frozen<address>",
			expected: &CQLTypeInfo{
				BaseType: "udt",
				UDTName:  "address",
				Frozen:   true,
			},
		},
		{
			name:  "list of UDTs",
			input: "list<frozen<address>>",
			expected: &CQLTypeInfo{
				BaseType: "list",
				Parameters: []*CQLTypeInfo{
					{
						BaseType: "udt",
						UDTName:  "address",
						Frozen:   true,
					},
				},
			},
		},
		{
			name:  "map with UDT value",
			input: "map<uuid, frozen<myks.user_profile>>",
			expected: &CQLTypeInfo{
				BaseType: "map",
				Parameters: []*CQLTypeInfo{
					{BaseType: "uuid"},
					{
						BaseType: "udt",
						UDTName:  "user_profile",
						Keyspace: "myks",
						Frozen:   true,
					},
				},
			},
		},

		// Edge cases with whitespace
		{
			name:  "type with spaces",
			input: "  map < text , int >  ",
			expected: &CQLTypeInfo{
				BaseType: "map",
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
					{BaseType: "int"},
				},
			},
		},
		{
			name:  "nested with spaces",
			input: "frozen < list < frozen < text > > >",
			expected: &CQLTypeInfo{
				BaseType: "list",
				Frozen:   true,
				Parameters: []*CQLTypeInfo{
					{BaseType: "text", Frozen: true},
				},
			},
		},

		// Error cases
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "unclosed frozen",
			input:   "frozen<text",
			wantErr: true,
		},
		{
			name:    "unclosed list",
			input:   "list<text",
			wantErr: true,
		},
		{
			name:    "map missing value type",
			input:   "map<text>",
			wantErr: true,
		},
		{
			name:    "map missing comma",
			input:   "map<text int>",
			wantErr: true,
		},
		{
			name:    "invalid frozen syntax",
			input:   "frozen text",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCQLType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCQLType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseCQLType() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestCQLTypeInfo_String(t *testing.T) {
	tests := []struct {
		name     string
		typeInfo *CQLTypeInfo
		expected string
	}{
		{
			name: "simple text",
			typeInfo: &CQLTypeInfo{
				BaseType: "text",
			},
			expected: "text",
		},
		{
			name: "frozen text",
			typeInfo: &CQLTypeInfo{
				BaseType: "text",
				Frozen:   true,
			},
			expected: "frozen<text>",
		},
		{
			name: "list of int",
			typeInfo: &CQLTypeInfo{
				BaseType: "list",
				Parameters: []*CQLTypeInfo{
					{BaseType: "int"},
				},
			},
			expected: "list<int>",
		},
		{
			name: "map of text to int",
			typeInfo: &CQLTypeInfo{
				BaseType: "map",
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
					{BaseType: "int"},
				},
			},
			expected: "map<text, int>",
		},
		{
			name: "tuple",
			typeInfo: &CQLTypeInfo{
				BaseType: "tuple",
				Parameters: []*CQLTypeInfo{
					{BaseType: "int"},
					{BaseType: "text"},
					{BaseType: "uuid"},
				},
			},
			expected: "tuple<int, text, uuid>",
		},
		{
			name: "simple UDT",
			typeInfo: &CQLTypeInfo{
				BaseType: "udt",
				UDTName:  "address",
			},
			expected: "address",
		},
		{
			name: "keyspace qualified UDT",
			typeInfo: &CQLTypeInfo{
				BaseType: "udt",
				UDTName:  "address",
				Keyspace: "myks",
			},
			expected: "myks.address",
		},
		{
			name: "nested frozen collections",
			typeInfo: &CQLTypeInfo{
				BaseType: "map",
				Frozen:   true,
				Parameters: []*CQLTypeInfo{
					{BaseType: "text"},
					{
						BaseType: "list",
						Frozen:   true,
						Parameters: []*CQLTypeInfo{
							{BaseType: "int"},
						},
					},
				},
			},
			expected: "frozen<map<text, frozen<list<int>>>>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.typeInfo.String()
			if got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that parsing and then stringifying gives back the same result
	testCases := []string{
		"text",
		"frozen<text>",
		"list<int>",
		"map<text, int>",
		"tuple<int, text, uuid>",
		"frozen<list<frozen<text>>>",
		"map<uuid, frozen<list<int>>>",
		"frozen<map<text, frozen<list<frozen<set<uuid>>>>>>",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			parsed, err := ParseCQLType(tc)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			result := parsed.String()
			// Re-parse to ensure it's equivalent
			reparsed, err := ParseCQLType(result)
			if err != nil {
				t.Fatalf("Failed to re-parse: %v", err)
			}
			if !reflect.DeepEqual(parsed, reparsed) {
				t.Errorf("Round trip failed: original=%+v, reparsed=%+v", parsed, reparsed)
			}
		})
	}
}