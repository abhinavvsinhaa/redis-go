package core

import (
	"testing"
)

func TestReadArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []interface{}
	}{
		{
			name:     "two bulk strings",
			input:    "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expected: []interface{}{"hello", "world"},
		},
		{
			name:     "three integers",
			input:    "*3\r\n:1\r\n:2\r\n:3\r\n",
			expected: []interface{}{int64(1), int64(2), int64(3)},
		},
		{
			name:     "empty array",
			input:    "*0\r\n",
			expected: []interface{}{},
		},
		{
			name:     "single element",
			input:    "*1\r\n$3\r\nfoo\r\n",
			expected: []interface{}{"foo"},
		},
		{
			name:     "mixed types - integers and bulk string",
			input:    "*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$5\r\nhello\r\n",
			expected: []interface{}{int64(1), int64(2), int64(3), int64(4), "hello"},
		},
		{
			name:     "all RESP types - simple string, error, integer, bulk string, nested array",
			input:    "*5\r\n+OK\r\n-ERR bad\r\n:42\r\n$6\r\nfoobar\r\n*1\r\n:1\r\n",
			expected: []interface{}{"OK", "ERR bad", int64(42), "foobar", []interface{}{int64(1)}},
		},
		{
			name:     "redis SET command format",
			input:    "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			expected: []interface{}{"SET", "key", "value"},
		},
		{
			name:  "nested arrays - two inner arrays",
			input: "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n-World\r\n",
			expected: []interface{}{
				[]interface{}{int64(1), int64(2), int64(3)},
				[]interface{}{"Hello", "World"},
			},
		},
		{
			name:  "deeply nested - three levels",
			input: "*1\r\n*1\r\n*1\r\n:42\r\n",
			expected: []interface{}{
				[]interface{}{
					[]interface{}{int64(42)},
				},
			},
		},
		{
			name:  "nested with empty inner array",
			input: "*2\r\n*0\r\n$3\r\nfoo\r\n",
			expected: []interface{}{
				[]interface{}{},
				"foo",
			},
		},
		{
			name:  "nested mixed - array then integer then simple string",
			input: "*3\r\n*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n:100\r\n+OK\r\n",
			expected: []interface{}{
				[]interface{}{"foo", "bar"},
				int64(100),
				"OK",
			},
		},
		{
			name:  "three levels nested with mixed types",
			input: "*2\r\n*2\r\n*1\r\n:1\r\n$2\r\nhi\r\n+OK\r\n",
			expected: []interface{}{
				[]interface{}{
					[]interface{}{int64(1)},
					"hi",
				},
				"OK",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := DecodeOne([]byte(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			arr, ok := result.([]interface{})
			if !ok {
				t.Fatalf("expected []interface{}, got %T", result)
			}

			assertArrayEqual(t, tc.expected, arr)
		})
	}
}

func TestReadSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "OK response",
			input:    "+OK\r\n",
			expected: "OK",
		},
		{
			name:     "longer string",
			input:    "+Hello World\r\n",
			expected: "Hello World",
		},
		{
			name:     "single character",
			input:    "+A\r\n",
			expected: "A",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := DecodeOne([]byte(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			str, ok := result.(string)
			if !ok {
				t.Fatalf("expected string, got %T", result)
			}
			if str != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, str)
			}
		})
	}
}

func TestReadError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "generic error",
			input:    "-ERR unknown command\r\n",
			expected: "ERR unknown command",
		},
		{
			name:     "wrong type error",
			input:    "-WRONGTYPE Operation against a key\r\n",
			expected: "WRONGTYPE Operation against a key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := DecodeOne([]byte(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			str, ok := result.(string)
			if !ok {
				t.Fatalf("expected string, got %T", result)
			}
			if str != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, str)
			}
		})
	}
}

func TestReadInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "single digit",
			input:    ":0\r\n",
			expected: 0,
		},
		{
			name:     "multi digit",
			input:    ":1279\r\n",
			expected: 1279,
		},
		{
			name:     "large number",
			input:    ":999999\r\n",
			expected: 999999,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := DecodeOne([]byte(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			val, ok := result.(int64)
			if !ok {
				t.Fatalf("expected int64, got %T", result)
			}
			if val != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, val)
			}
		})
	}
}

func TestReadBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short string",
			input:    "$3\r\nfoo\r\n",
			expected: "foo",
		},
		{
			name:     "longer string",
			input:    "$6\r\nfoobar\r\n",
			expected: "foobar",
		},
		{
			name:     "empty bulk string",
			input:    "$0\r\n\r\n",
			expected: "",
		},
		{
			name:     "string with spaces",
			input:    "$11\r\nhello world\r\n",
			expected: "hello world",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := DecodeOne([]byte(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			str, ok := result.(string)
			if !ok {
				t.Fatalf("expected string, got %T", result)
			}
			if str != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, str)
			}
		})
	}
}

func TestDecodeOne_UnknownType(t *testing.T) {
	data := []byte("!invalid\r\n")
	_, _, err := DecodeOne(data)
	if err == nil {
		t.Fatal("expected error for unknown RESP type, got nil")
	}
}

// assertArrayEqual recursively compares two []interface{} slices
func assertArrayEqual(t *testing.T, expected, actual []interface{}) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(actual))
	}
	for i := range expected {
		switch exp := expected[i].(type) {
		case []interface{}:
			act, ok := actual[i].([]interface{})
			if !ok {
				t.Fatalf("element %d: expected []interface{}, got %T", i, actual[i])
			}
			assertArrayEqual(t, exp, act)
		case int64:
			act, ok := actual[i].(int64)
			if !ok {
				t.Fatalf("element %d: expected int64, got %T", i, actual[i])
			}
			if act != exp {
				t.Errorf("element %d: expected %d, got %d", i, exp, act)
			}
		case string:
			act, ok := actual[i].(string)
			if !ok {
				t.Fatalf("element %d: expected string, got %T", i, actual[i])
			}
			if act != exp {
				t.Errorf("element %d: expected %q, got %q", i, exp, act)
			}
		default:
			t.Fatalf("element %d: unsupported type %T in expected", i, expected[i])
		}
	}
}
