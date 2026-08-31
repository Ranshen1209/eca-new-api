package model

import (
	"strings"
)

// 简化的供应商映射规则
var defaultVendorRules = map[string]string{
	"gpt":      "OpenAI",
	"dall-e":   "OpenAI",
	"whisper":  "OpenAI",
	"o1":       "OpenAI",
	"o3":       "OpenAI",
	"claude":   "Anthropic",
	"gemini":   "Google",
	"moonshot": "Moonshot",
	"kimi":     "Moonshot",
	"chatglm":  "智谱",
	"glm-":     "智谱",
	"qwen":     "阿里巴巴",
	"deepseek": "DeepSeek",
	"mimo":     "Mimo",
	"abab":     "MiniMax",
	"minimax":  "MiniMax",
	"ernie":    "百度",
	"spark":    "讯飞",
	"hunyuan":  "腾讯",
	"command":  "Cohere",
	"@cf/":     "Cloudflare",
	"360":      "360",
	"yi":       "零一万物",
	"jina":     "Jina",
	"mistral":  "Mistral",
	"grok":     "xAI",
	"llama":    "Meta",
	"doubao":   "字节跳动",
	"kling":    "快手",
	"jimeng":   "即梦",
	"vidu":     "Vidu",
}

// 供应商默认图标映射
var defaultVendorIcons = map[string]string{
	"OpenAI":     "OpenAI",
	"Anthropic":  "Claude.Color",
	"Google":     "Gemini.Color",
	"Moonshot":   "Moonshot",
	"智谱":         "Zhipu.Color",
	"阿里巴巴":       "Qwen.Color",
	"DeepSeek":   "DeepSeek.Color",
	"Mimo":       "XiaomiMiMo",
	"MiniMax":    "Minimax.Color",
	"百度":         "Wenxin.Color",
	"讯飞":         "Spark.Color",
	"腾讯":         "Hunyuan.Color",
	"Cohere":     "Cohere.Color",
	"Cloudflare": "Cloudflare.Color",
	"360":        "Ai360.Color",
	"零一万物":       "Yi.Color",
	"Jina":       "Jina",
	"Mistral":    "Mistral.Color",
	"xAI":        "XAI",
	"Meta":       "Ollama",
	"字节跳动":       "Doubao.Color",
	"快手":         "Kling.Color",
	"即梦":         "Jimeng.Color",
	"Vidu":       "Vidu",
	"微软":         "AzureAI",
	"Microsoft":  "AzureAI",
	"Azure":      "AzureAI",
}

var vendorRenames = map[string]string{
	"Xiaomi Mimo": "Mimo",
}

var defaultModelDescriptions = map[string]string{
	"deepseek-v4-flash":            "高性价比通用模型，适合日常问答、摘要、改写、信息提取和批量任务。",
	"deepseek-v4-flash-vision-exp": "支持视觉理解的实验性模型，适合截图分析、OCR、图表识别和多模态 Agent。",
	"deepseek-v4-pro":              "世界知识储备丰富的旗舰模型，适合研究问答、复杂分析、长文本理解和专业内容生成。",
	"glm-5.3":                      "编程与软件工程能力突出，适合代码生成、调试、重构、项目开发和 Agent 任务。",
	"glm-5.3-flash":                "快速经济的通用及编程模型，适合代码补全、简单脚本、摘要、分类和批量调用。",
	"mimo-v2.5":                    "原生全模态模型，支持图像、视频、音频和文本理解，适合 OCR、图表分析、视觉问答和多模态 Agent。",
	"mimo-v2.5-pro":                "面向复杂任务的旗舰文本模型，擅长长程推理、复杂问答、Agent 工作流和专业内容处理。",
	"mimo-v2.5-pro-ultraspeed":     "MiMo-V2.5-Pro 的高速模式，峰值约 1000 tokens/s，适合实时响应、高并发调用、代码生成和流式应用。",
}

// initDefaultVendorMapping 简化的默认供应商映射
func initDefaultVendorMapping(metaMap map[string]*Model, vendorMap map[int]*Vendor, enableAbilities []AbilityWithChannel) {
	renameDefaultVendors(vendorMap)

	for _, ability := range enableAbilities {
		modelName := ability.Model
		if meta, exists := metaMap[modelName]; exists {
			applyDefaultModelDescription(meta)
			continue
		}

		vendorID := 0
		modelLower := strings.ToLower(modelName)
		for pattern, vendorName := range defaultVendorRules {
			if strings.Contains(modelLower, pattern) {
				vendorID = getOrCreateVendor(vendorName, vendorMap)
				break
			}
		}

		meta := &Model{
			ModelName:   modelName,
			Description: defaultModelDescriptions[modelName],
			VendorID:    vendorID,
			Status:      1,
			NameRule:    NameRuleExact,
		}
		metaMap[modelName] = meta
	}
}

func renameDefaultVendors(vendorMap map[int]*Vendor) {
	for _, vendor := range vendorMap {
		newName, ok := vendorRenames[vendor.Name]
		if !ok {
			continue
		}
		nameTaken := false
		for id, existing := range vendorMap {
			if id != vendor.Id && existing.Name == newName {
				nameTaken = true
				break
			}
		}
		if nameTaken {
			continue
		}
		vendor.Name = newName
		if icon := getDefaultVendorIcon(newName); icon != "" {
			vendor.Icon = icon
		}
		_ = vendor.Update()
	}
}

func applyDefaultModelDescription(meta *Model) {
	if meta == nil {
		return
	}
	desc := defaultModelDescriptions[meta.ModelName]
	if desc == "" || strings.TrimSpace(meta.Description) != "" {
		return
	}
	meta.Description = desc
	if meta.Id > 0 {
		_ = meta.Update()
	}
}

// 查找或创建供应商
func getOrCreateVendor(vendorName string, vendorMap map[int]*Vendor) int {
	// 查找现有供应商
	for id, vendor := range vendorMap {
		if vendor.Name == vendorName {
			return id
		}
	}

	// 创建新供应商
	newVendor := &Vendor{
		Name:   vendorName,
		Status: 1,
		Icon:   getDefaultVendorIcon(vendorName),
	}

	if err := newVendor.Insert(); err != nil {
		return 0
	}

	vendorMap[newVendor.Id] = newVendor
	return newVendor.Id
}

// 获取供应商默认图标
func getDefaultVendorIcon(vendorName string) string {
	if icon, exists := defaultVendorIcons[vendorName]; exists {
		return icon
	}
	return ""
}
