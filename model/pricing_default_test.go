package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultVendorRulesMapMimoToMimo(t *testing.T) {
	assert.Equal(t, "Mimo", defaultVendorRules["mimo"])
	assert.Equal(t, "XiaomiMiMo", defaultVendorIcons["Mimo"])
}

func TestDefaultModelDescriptionsCoverCurrentCatalog(t *testing.T) {
	want := map[string]string{
		"deepseek-v4-flash":            "高性价比通用模型，适合日常问答、摘要、改写、信息提取和批量任务。",
		"deepseek-v4-flash-vision-exp": "支持视觉理解的实验性模型，适合截图分析、OCR、图表识别和多模态 Agent。",
		"deepseek-v4-pro":              "世界知识储备丰富的旗舰模型，适合研究问答、复杂分析、长文本理解和专业内容生成。",
		"glm-5.3":                      "编程与软件工程能力突出，适合代码生成、调试、重构、项目开发和 Agent 任务。",
		"glm-5.3-flash":                "快速经济的通用及编程模型，适合代码补全、简单脚本、摘要、分类和批量调用。",
		"mimo-v2.5":                    "原生全模态模型，支持图像、视频、音频和文本理解，适合 OCR、图表分析、视觉问答和多模态 Agent。",
		"mimo-v2.5-pro":                "面向复杂任务的旗舰文本模型，擅长长程推理、复杂问答、Agent 工作流和专业内容处理。",
		"mimo-v2.5-pro-ultraspeed":     "MiMo-V2.5-Pro 的高速模式，峰值约 1000 tokens/s，适合实时响应、高并发调用、代码生成和流式应用。",
	}

	for modelName, description := range want {
		assert.Equal(t, description, defaultModelDescriptions[modelName], modelName)
	}
	assert.Len(t, defaultModelDescriptions, len(want))
}

func TestInitDefaultVendorMappingFillsEmptyDescriptionAndPersists(t *testing.T) {
	resetPricingEndpointTestTables(t)

	existing := &Model{
		ModelName: "mimo-v2.5",
		Status:    1,
		NameRule:  NameRuleExact,
	}
	require.NoError(t, existing.Insert())

	metaMap := map[string]*Model{
		existing.ModelName: existing,
	}
	initDefaultVendorMapping(metaMap, map[int]*Vendor{}, []AbilityWithChannel{
		{Ability: Ability{Model: "mimo-v2.5"}},
	})

	assert.Equal(t, defaultModelDescriptions["mimo-v2.5"], existing.Description)

	var stored Model
	require.NoError(t, DB.First(&stored, existing.Id).Error)
	assert.Equal(t, defaultModelDescriptions["mimo-v2.5"], stored.Description)
}

func TestInitDefaultVendorMappingDoesNotOverwriteExistingDescription(t *testing.T) {
	resetPricingEndpointTestTables(t)

	existing := &Model{
		ModelName:   "glm-5.3",
		Description: "管理员已填写的描述",
		Status:      1,
		NameRule:    NameRuleExact,
	}
	require.NoError(t, existing.Insert())

	metaMap := map[string]*Model{
		existing.ModelName: existing,
	}
	initDefaultVendorMapping(metaMap, map[int]*Vendor{}, []AbilityWithChannel{
		{Ability: Ability{Model: "glm-5.3"}},
	})

	assert.Equal(t, "管理员已填写的描述", existing.Description)
}

func TestInitDefaultVendorMappingRenamesXiaomiMimoVendor(t *testing.T) {
	resetPricingEndpointTestTables(t)

	vendor := &Vendor{
		Name:   "Xiaomi Mimo",
		Icon:   "XiaomiMiMo",
		Status: 1,
	}
	require.NoError(t, vendor.Insert())

	vendorMap := map[int]*Vendor{
		vendor.Id: vendor,
	}
	initDefaultVendorMapping(map[string]*Model{}, vendorMap, nil)

	assert.Equal(t, "Mimo", vendor.Name)
	assert.Equal(t, "XiaomiMiMo", vendor.Icon)

	var stored Vendor
	require.NoError(t, DB.First(&stored, vendor.Id).Error)
	assert.Equal(t, "Mimo", stored.Name)
}

func TestInitDefaultVendorMappingCreatesMissingModelWithDescription(t *testing.T) {
	resetPricingEndpointTestTables(t)

	metaMap := map[string]*Model{}
	initDefaultVendorMapping(metaMap, map[int]*Vendor{}, []AbilityWithChannel{
		{Ability: Ability{Model: "deepseek-v4-pro"}},
	})

	created := metaMap["deepseek-v4-pro"]
	require.NotNil(t, created)
	assert.Equal(t, defaultModelDescriptions["deepseek-v4-pro"], created.Description)
	assert.Equal(t, 0, created.Id)
}
