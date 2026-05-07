package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestObject struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

func TestUnmarshalLocalFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yaml")
	content := "name: test\nvalue: 123"
	
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	var obj TestObject
	err = UnmarshalLocalFile(path, &obj)
	require.NoError(t, err)
	assert.Equal(t, "test", obj.Name)
	assert.Equal(t, 123, obj.Value)
}

func TestUnmarshal(t *testing.T) {
	content := []byte("name: test\nvalue: 123")
	var obj TestObject
	err := Unmarshal(content, &obj)
	require.NoError(t, err)
	assert.Equal(t, "test", obj.Name)
	assert.Equal(t, 123, obj.Value)
}

func TestMarshalLocalYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yaml")
	obj := TestObject{Name: "test", Value: 123}

	err := MarshalLocalYAMLFile(path, &obj)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: test")
	assert.Contains(t, string(data), "value: 123")
}
