package tests

import (
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestSet(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "Set and Get Simple Value",
			commands: []string{"SET k v", "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "Overwrite Existing Key",
			commands: []string{"SET k v1", "SET k v2", "GET k"},
			expected: []interface{}{"OK", "OK", "v2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.DeepEqual(t, tc.expected[i], result)
			}
		})
	}
}

func TestSetWithOptions(t *testing.T) {
	conn := getLocalConnection()
	expiryTimeInMs := strconv.FormatInt(time.Now().Add(1*time.Minute).UnixMilli(), 10)
	expiryTimeInSec := strconv.FormatInt(time.Now().Add(1*time.Minute).Unix(), 10)
	defer conn.Close()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
	}{
		{
			name:     "XX on non-existing key",
			commands: []string{"DEL k", "SET k v XX", "GET k"},
			expected: []interface{}{int64(0), "(nil)", "(nil)"},
		},
		{
			name:     "PXAT option",
			commands: []string{"SET k v PXAT " + expiryTimeInMs, "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "PXAT option with delete",
			commands: []string{"SET k1 v1 PXAT " + expiryTimeInMs, "GET k1", "SLEEP 2", "DEL k1"},
			expected: []interface{}{"OK", "v1", "OK", int64(1)},
		},
		{
			name:     "PXAT option with invalid unix time ms",
			commands: []string{"SET k2 v2 PXAT 123123", "GET k2"},
			expected: []interface{}{"OK", "(nil)"},
		},
		{
			name:     "XX on existing key",
			commands: []string{"SET k v1", "SET k v2 XX", "GET k"},
			expected: []interface{}{"OK", "OK", "v2"},
		},
		{
			name:     "Multiple XX operations",
			commands: []string{"SET k v1", "SET k v2 XX", "SET k v3 XX", "GET k"},
			expected: []interface{}{"OK", "OK", "OK", "v3"},
		},
		{
			name:     "EX option",
			commands: []string{"SET k v EX 1", "GET k", "SLEEP 2", "GET k"},
			expected: []interface{}{"OK", "v", "OK", "(nil)"},
		},
		{
			name:     "XX option",
			commands: []string{"SET k v XX EX 1", "GET k", "SLEEP 2", "GET k", "SET k v XX EX 1", "GET k"},
			expected: []interface{}{"(nil)", "(nil)", "OK", "(nil)", "(nil)", "(nil)"},
		},
		{
			name:     "EXAT option",
			commands: []string{"SET k v EXAT " + expiryTimeInSec, "GET k"},
			expected: []interface{}{"OK", "v"},
		},
		{
			name:     "EXAT option with delete",
			commands: []string{"SET k1 v1 EXAT " + expiryTimeInSec, "GET k1", "SLEEP 2", "DEL k1"},
			expected: []interface{}{"OK", "v1", "OK", int64(1)},
		},
		{
			name:     "EXAT option with invalid unix time ms",
			commands: []string{"SET k2 v2 EXAT 123123", "GET k2"},
			expected: []interface{}{"OK", "(nil)"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestKeys([]string{"k", "k1", "k2"})
			for i, cmd := range tc.commands {
				result := fireCommand(conn, cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
