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
			"openai":    "glm-4.7",
			"anthropic": "claude-3-sonnet",
		},
	},
	"glm-coding": {
		Protocols: map[string]string{
			"openai": "https://open.bigmodel.cn/api/coding/paas/v4",
		},
		DefaultModels: map[string]string{
			"openai": "glm-4.7",
		},
	},
	"kimi": {
		Protocols: map[string]string{
			"openai":    "https://api.moonshot.cn/v1",
			"anthropic": "https://api.moonshot.cn/anthropic",
		},
		DefaultModels: map[string]string{
			"openai":    "kimi-k2.5",
			"anthropic": "claude-3-sonnet",
		},
	},
	"kimi-coding": {
		Protocols: map[string]string{
			"openai": "https://api.kimi.com/coding/v1",
		},
		DefaultModels: map[string]string{
			"openai": "kimi-for-coding",
		},
	},
	"qwen": {
		Protocols: map[string]string{
			"openai": "https://dashscope.aliyuncs.com/compatible-mode/v1",
		},
		DefaultModels: map[string]string{
			"openai": "qwen3-max",
		},
	},
}

// Vendor represents a vendor with its protocols
type Vendor struct {
	Protocols     map[string]string            // protocol -> URL
	DefaultModels map[string]string            // protocol -> default model name
}
