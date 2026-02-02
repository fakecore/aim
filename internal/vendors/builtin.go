package vendors

// BuiltinVendors contains the built-in vendor definitions
var BuiltinVendors = map[string]Vendor{
	"deepseek": {
		Protocols: map[string]string{
			"openai":    "https://api.deepseek.com/v1",
			"anthropic": "https://api.deepseek.com/anthropic",
		},
	},
	"glm": {
		Protocols: map[string]string{
			"openai":    "https://open.bigmodel.cn/api/paas/v4",
			"anthropic": "https://open.bigmodel.cn/api/anthropic",
		},
	},
	"kimi": {
		Protocols: map[string]string{
			"openai": "https://api.moonshot.cn/v1",
		},
	},
	"qwen": {
		Protocols: map[string]string{
			"openai": "https://dashscope.aliyuncs.com/compatible-mode/v1",
		},
	},
}

// Vendor represents a vendor with its protocols
type Vendor struct {
	Protocols map[string]string
}
