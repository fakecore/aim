package extension

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtension_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ext     Extension
		wantErr bool
	}{
		{
			name: "valid extension",
			ext: Extension{
				Name:      "test",
				Protocols: map[string]Protocol{"openai": {URL: "https://api.test.com"}},
			},
			wantErr: false,
		},
		{
			name:    "missing name",
			ext:     Extension{Protocols: map[string]Protocol{"openai": {URL: "https://api.test.com"}}},
			wantErr: true,
		},
		{
			name:    "no protocols",
			ext:     Extension{Name: "test"},
			wantErr: true,
		},
		{
			name:    "missing URL",
			ext:     Extension{Name: "test", Protocols: map[string]Protocol{"openai": {}}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ext.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yaml")

	content := `
name: siliconflow
version: "1.0.0"
protocols:
  openai:
    url: https://api.siliconflow.cn/v1
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	ext, err := LoadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "siliconflow", ext.Name)
	assert.Equal(t, "1.0.0", ext.Version)
	assert.Equal(t, "https://api.siliconflow.cn/v1", ext.Protocols["openai"].URL)
}

func TestLoadDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two extension files
	content1 := `
name: ext1
protocols:
  openai:
    url: https://api.ext1.com
`
	content2 := `
name: ext2
protocols:
  anthropic:
    url: https://api.ext2.com
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "ext1.yaml"), []byte(content1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "ext2.yaml"), []byte(content2), 0644))

	exts, err := LoadDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, exts, 2)
	assert.Contains(t, exts, "ext1")
	assert.Contains(t, exts, "ext2")
}
