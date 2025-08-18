package models

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// JSONB 类型用于处理 PostgreSQL 的 JSONB 字段
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// ToJSONSchema 将 JSONB 转换为 JSON Schema
func (j *JSONB) ToJSONSchema() (*jsonschema.Schema, error) {
	if j == nil {
		return nil, nil
	}

	// 将 JSONB 转换为 map
	data := map[string]interface{}(*j)
	if data == nil {
		return nil, nil
	}

	// 创建 JSON Schema
	schema := &jsonschema.Schema{}

	// 设置基本属性
	if schemaType, ok := data["type"].(string); ok {
		schema.Type = schemaType
	}

	if title, ok := data["title"].(string); ok {
		schema.Title = title
	}

	if description, ok := data["description"].(string); ok {
		schema.Description = description
	}

	// 处理 properties
	if properties, ok := data["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]*jsonschema.Schema)
		for propName, propValue := range properties {
			if propMap, ok := propValue.(map[string]interface{}); ok {
				propSchema := &jsonschema.Schema{}

				if propType, ok := propMap["type"].(string); ok {
					propSchema.Type = propType
				}

				if propDesc, ok := propMap["description"].(string); ok {
					propSchema.Description = propDesc
				}

				schema.Properties[propName] = propSchema
			}
		}
	}

	// 处理 required 字段
	if required, ok := data["required"].([]interface{}); ok {
		schema.Required = make([]string, len(required))
		for i, req := range required {
			if reqStr, ok := req.(string); ok {
				schema.Required[i] = reqStr
			}
		}
	}

	return schema, nil
}
