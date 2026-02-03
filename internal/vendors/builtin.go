package vendors

// BuiltinVendors contains the built-in vendor definitions
// Each vendor has multiple endpoints (where endpoint name = protocol type)
var BuiltinVendors = map[string]Vendor{
	// 深度求索 - 按量付费 / Pay-as-you-go
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
	// 智谱 GLM - 订阅制 / Subscription
	// openai-coding 端点需使用 GLM 编码套餐
	"glm": {
		Endpoints: map[string]Endpoint{
			"openai": {
				URL:          "https://open.bigmodel.cn/api/paas/v4",
				DefaultModel: "glm-4.7",
			},
			"openai-coding": { // Coding 专属端点
				URL:          "https://open.bigmodel.cn/api/coding/paas/v4",
				DefaultModel: "glm-4-coding",
			},
			"anthropic": {
				URL:          "https://open.bigmodel.cn/api/anthropic",
				DefaultModel: "glm-4.7",
			},
			"anthropic-coding": { // Coding 专属端点
				URL:          "https://open.bigmodel.cn/api/anthropic",
				DefaultModel: "glm-4-coding",
			},
		},
	},
	// 月之暗面 Kimi - 订阅制 / Subscription
	// coding 端点需使用 kimi-for-coding 模型
	"kimi": {
		Endpoints: map[string]Endpoint{
			"openai": {
				URL:          "https://api.moonshot.cn/v1",
				DefaultModel: "moonshot-v1-8k",
			},
			"openai-coding": { // Coding 专属端点
				URL:          "https://api.kimi.com/coding/v1",
				DefaultModel: "kimi-for-coding",
			},
			"anthropic": {
				URL:          "https://api.moonshot.cn/anthropic",
				DefaultModel: "moonshot-v1-8k",
			},
			"anthropic-coding": { // Coding 专属端点
				URL:          "https://api.kimi.com/coding/",
				DefaultModel: "kimi-for-coding",
			},
		},
	},
	// 通义千问 Qwen - 按量付费 / Pay-as-you-go
	"qwen": {
		Endpoints: map[string]Endpoint{
			"openai": {
				URL:          "https://dashscope.aliyuncs.com/compatible-mode/v1",
				DefaultModel: "qwen-plus",
			},
			"anthropic": {
				URL:          "https://dashscope.aliyuncs.com/api/v2/apps/claude-code-proxy",
				DefaultModel: "qwen-plus",
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
