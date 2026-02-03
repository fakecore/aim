package vendors

// BuiltinVendors contains the built-in vendor definitions
// Each vendor has multiple endpoints (where endpoint name = protocol type)
var BuiltinVendors = map[string]Vendor{
	"deepseek": {
		Endpoints: map[string]Endpoint{
			"openai": {
				URL:          "https://api.deepseek.com/v1",
				DefaultModel: "deepseek-chat",
			},
			"anthropic": {
				URL:          "https://api.deepseek.com/anthropic",
				DefaultModel: "deepseek-chat",
			},
		},
	},
	"glm": {
		Endpoints: map[string]Endpoint{
			"openai": {
				URL:          "https://open.bigmodel.cn/api/paas/v4",
				DefaultModel: "glm-4.7",
			},
			"anthropic": {
				URL:          "https://open.bigmodel.cn/api/anthropic",
				DefaultModel: "glm-4.7",
			},
		},
	},
	"kimi": {
		Endpoints: map[string]Endpoint{
			"openai": {
				URL:          "https://api.moonshot.cn/v1",
				DefaultModel: "kimi-k2.5",
			},
			"anthropic": {
				URL:          "https://api.moonshot.cn/anthropic",
				DefaultModel: "claude-3-sonnet",
			},
		},
	},
	"qwen": {
		Endpoints: map[string]Endpoint{
			"openai": {
				URL:          "https://dashscope.aliyuncs.com/compatible-mode/v1",
				DefaultModel: "qwen3-max",
			},
			"anthropic": {
				URL:          "https://dashscope.aliyuncs.com/api/v2/apps/claude-code-proxy",
				DefaultModel: "qwen3-max",
			},
		},
	},
}

// Vendor represents a vendor with its endpoints
type Vendor struct {
	Endpoints map[string]Endpoint // endpoint name (e.g., "openai", "anthropic") -> config
}

// Endpoint represents a single endpoint configuration
type Endpoint struct {
	URL          string `yaml:"url"`
	DefaultModel string `yaml:"default_model,omitempty"`
}
