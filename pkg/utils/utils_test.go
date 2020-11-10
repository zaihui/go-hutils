package utils

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONMarshal(t *testing.T) {
	symbol := "<" // "\u003c"

	escape, err := json.Marshal(symbol)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(string(escape)), 8)

	noescape, err := JSONMarshal(symbol)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(string(noescape)), 4)
}

func TestGetEnv(t *testing.T) {
	value := GetEnv("key", "1")
	assert.Equal(t, value, "1")

	err := os.Setenv("key", "2")
	assert.Equal(t, err, nil)
	value = GetEnv("key", "")
	assert.Equal(t, value, "2")
}

func TestNewUUID(t *testing.T) {
	uid := NewUUID()
	assert.Equal(t, len(uid), 32)
}

func TestNewApm(t *testing.T) {
	span := NewApmSpan(context.Background(), "TestApm", "test")
	assert.NotEqual(t, span, nil)
}
