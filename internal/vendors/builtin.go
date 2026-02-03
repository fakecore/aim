package vendors

// BuiltinVendors contains the built-in vendor definitions
var BuiltinVendors = map[string]Vendor{
	"deepseek": {
		Protocols: map[string]string{
			"openai":    "https://api.deepseek.com/v1",
			"anthropic": "https://api.deepseek.com/anthropic",
		},
		DefaultModels: map[string]string{
			"openai":    "deepseek-chat",
			"anthropic": "claude-3-sonnet",
		},
	},
	"glm": {
		Protocols: map[string]string{
			"openai":    "https://open.bigmodel.cn/api/paas/v4",
			"anthropic": "https://open.bigmodel.cn/api/anthropic",
		},
		DefaultModels: map[string]string{
			"openai":    "glm-4",
			"anthropic": "claude-3-sonnet",
		},
	},
	"kimi": {
		Protocols: map[string]string{
			"openai":    "https://api.moonshot.cn/v1",
			"anthropic": "https://api.moonshot.cn/anthropic",
		},
		DefaultModels: map[string]string{
			"openai":    "kimi-latest",
			"anthropic": "claude-3-sonnet",
		},
	},
	"qwen": {
		Protocols: map[string]string{
			"openai": "https://dashscope.aliyuncs.com/compatible-mode/v1",
		},
		DefaultModels: map[string]string{
			"openai": "qwen-max",
		},
	},
}

// Vendor represents a vendor with its protocols
type Vendor struct {
	Protocols     map[string]string            // protocol -> URL
	DefaultModels map[string]string            // protocol -> default model name
}
