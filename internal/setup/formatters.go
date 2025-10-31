package setup

import (
	"encoding/json"
	"fmt"
	"strings"
)

var shellReplacer = strings.NewReplacer(
	`"`, `\"`,
	`$`, `\$`,
	"`", "\\`",
	`\`, `\\`,
	`!`, `\!`,
	`&`, `\&`,
	`|`, `\|`,
	`(`, `\(`,
	`)`, `\)`,
	`[`, `\[`,
	`]`, `\]`,
	`{`, `\{`,
	`}`, `\}`,
	`;`, `\;`,
	`<`, `\<`,
	`>`, `\>`,
	` `, `\ `,
)

// EscapeShellValue escapes special characters in shell values
func EscapeShellValue(value string) string {
	return shellReplacer.Replace(value)
}

// ZshFormatter Zsh environment variable formatter
type ZshFormatter struct{}

func (f *ZshFormatter) GetName() string {
	return "zsh"
}

func (f *ZshFormatter) GetContentType() string {
	return "application/x-sh"
}

func (f *ZshFormatter) FormatEnv(result *SetupResult) string {
	var output strings.Builder

	for key, value := range result.EnvVars {
		escapedValue := EscapeShellValue(value)
		output.WriteString(fmt.Sprintf("export %s=\"%s\"\n", key, escapedValue))
	}

	return output.String()
}

// BashFormatter Bash environment variable formatter
type BashFormatter struct{}

func (f *BashFormatter) GetName() string {
	return "bash"
}

func (f *BashFormatter) GetContentType() string {
	return "application/x-sh"
}

func (f *BashFormatter) FormatEnv(result *SetupResult) string {
	var output strings.Builder

	for key, value := range result.EnvVars {
		escapedValue := EscapeShellValue(value)
		output.WriteString(fmt.Sprintf("export %s=\"%s\"\n", key, escapedValue))
	}

	return output.String()
}

// FishFormatter Fish environment variable formatter
type FishFormatter struct{}

func (f *FishFormatter) GetName() string {
	return "fish"
}

func (f *FishFormatter) GetContentType() string {
	return "application/x-fish"
}

func (f *FishFormatter) FormatEnv(result *SetupResult) string {
	var output strings.Builder

	for key, value := range result.EnvVars {
		// Fish uses different syntax
		escapedValue := strings.ReplaceAll(value, `"`, `\"`)
		escapedValue = strings.ReplaceAll(escapedValue, `\\`, `\\\\`)

		output.WriteString(fmt.Sprintf("set -gx %s \"%s\"\n", key, escapedValue))
	}

	return output.String()
}

// JSONFormatter JSON environment variable formatter
type JSONFormatter struct {
	pretty bool
}

func (f *JSONFormatter) GetName() string {
	return "json"
}

func (f *JSONFormatter) GetContentType() string {
	return "application/json"
}

func (f *JSONFormatter) FormatEnv(result *SetupResult) string {
	// Create JSON output structure
	envOutput := map[string]interface{}{
		"tool":      result.Request.ToolName,
		"key":       result.Request.KeyName,
		"env_vars":  result.EnvVars,
		"generated": result.Generated,
		"metadata": map[string]interface{}{
			"duration": result.Metadata.Duration.String(),
			"source":   result.Metadata.Source,
			"version":  result.Metadata.Version,
		},
	}

	var data []byte
	var err error

	if f.pretty {
		data, err = json.MarshalIndent(envOutput, "", "  ")
	} else {
		data, err = json.Marshal(envOutput)
	}

	if err != nil {
		// If JSON serialization fails, return error message
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}

	return string(data)
}

// RawCommandFormatter Raw command formatter
type RawCommandFormatter struct{}

func (f *RawCommandFormatter) GetName() string {
	return "raw"
}

func (f *RawCommandFormatter) GetContentType() string {
	return "text/plain"
}

func (f *RawCommandFormatter) FormatCommand(result *SetupResult) string {
	return result.Command
}

// ShellCommandFormatter Shell command formatter
type ShellCommandFormatter struct{}

func (f *ShellCommandFormatter) GetName() string {
	return "shell"
}

func (f *ShellCommandFormatter) GetContentType() string {
	return "text/plain"
}

func (f *ShellCommandFormatter) FormatCommand(result *SetupResult) string {
	// Escape shell special characters
	escaped := strings.ReplaceAll(result.Command, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, `$`, `\$`)
	escaped = strings.ReplaceAll(escaped, "`", "\\`")
	escaped = strings.ReplaceAll(escaped, `&`, `\&`)
	escaped = strings.ReplaceAll(escaped, `;`, `\;`)
	escaped = strings.ReplaceAll(escaped, `|`, `\|`)
	escaped = strings.ReplaceAll(escaped, `>`, `\>`)
	escaped = strings.ReplaceAll(escaped, `<`, `\<`)
	escaped = strings.ReplaceAll(escaped, `(`, `\(`)
	escaped = strings.ReplaceAll(escaped, `)`, `\)`)

	return fmt.Sprintf(`"%s"`, escaped)
}

// JSONCommandFormatter JSON command formatter
type JSONCommandFormatter struct{}

func (f *JSONCommandFormatter) GetName() string {
	return "json"
}

func (f *JSONCommandFormatter) GetContentType() string {
	return "application/json"
}

func (f *JSONCommandFormatter) FormatCommand(result *SetupResult) string {
	// Create JSON output structure
	commandOutput := map[string]interface{}{
		"tool":      result.Request.ToolName,
		"key":       result.Request.KeyName,
		"command":   result.Command,
		"generated": result.Generated,
		"metadata": map[string]interface{}{
			"duration": result.Metadata.Duration.String(),
			"source":   result.Metadata.Source,
			"version":  result.Metadata.Version,
		},
	}

	data, err := json.MarshalIndent(commandOutput, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}

	return string(data)
}

// SimpleCommandFormatter Simple command formatter (without environment variables)
type SimpleCommandFormatter struct{}

func (f *SimpleCommandFormatter) GetName() string {
	return "simple"
}

func (f *SimpleCommandFormatter) GetContentType() string {
	return "text/plain"
}

func (f *SimpleCommandFormatter) FormatCommand(result *SetupResult) string {
	// Extract base command from command containing environment variables
	// Command format: export VAR1="value1" && export VAR2="value2" && command
	parts := strings.Split(result.Command, " && ")
	if len(parts) == 0 {
		return ""
	}

	// The last part is the actual command
	return parts[len(parts)-1]
}
