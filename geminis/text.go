package geminis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	tele "gopkg.in/telebot.v3"
)

func GetTextResponse(c tele.Context, model *genai.GenerativeModel, text string) (resp *genai.GenerateContentResponse) {
	ctx := context.Background()
	resp, err := model.GenerateContent(ctx, genai.Text(text))
	if err != nil {
		c.Send(fmt.Sprintf("Error while generating content: %v", err))
		return nil
	}
	jsonString, err := json.Marshal(resp)
	if err != nil {
		c.Send(fmt.Sprintf("Error marshalling JSON: %v", err))
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		c.Send(fmt.Sprintf("Error parsing JSON: %v", err))
		return nil
	}
	parts, ok := data["Candidates"].([]interface{})
	if !ok || len(parts) == 0 {
		c.Send(fmt.Sprintf("No candidates found in response: %v", resp))
		return nil
	}

	// 输出Parts部分
	for _, candidate := range parts {
		candidateMap, ok := candidate.(map[string]interface{})
		if !ok {
			// 如果候选项不是map类型，跳过
			continue
		}
		content, ok := candidateMap["Content"].(map[string]interface{})
		if !ok {
			// 如果Content部分不是map类型，跳过
			continue
		}
		partsArray, ok := content["Parts"].([]interface{})
		if !ok {
			// 如果Parts部分不是数组类型，跳过
			continue
		}
		for _, part := range partsArray {
			c.Send(part, &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
		}
	}
	return
}
