package geminis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	tele "gopkg.in/telebot.v3"
)

func GetPhotoResponse(c tele.Context, vmodel *genai.GenerativeModel, imageType string, text string, data []byte) (resp *genai.GenerateContentResponse) {
	ctx := context.Background()
	prompt := []genai.Part{
		genai.ImageData(imageType, data),
		genai.Text(text),
	}
	resp, err := vmodel.GenerateContent(ctx, prompt...)
	if err != nil {
		c.Send(fmt.Sprintf("Error while generating content: %v", err))
		return nil
	}
	jsonString, err := json.Marshal(resp)
	if err != nil {
		c.Send(fmt.Sprintf("Error marshalling JSON: %v", err))
		return nil
	}
	var data1 map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &data1); err != nil {
		c.Send(fmt.Sprintf("Error parsing JSON: %v", err))
		return nil
	}
	parts, ok := data1["Candidates"].([]interface{})
	if !ok || len(parts) == 0 {
		c.Send(fmt.Sprintf("No candidates found in response: %v", resp))
		return nil
	}

	for _, candidate := range parts {
		candidateMap, ok := candidate.(map[string]interface{})
		if !ok {
			// 如果候选项不是map类型，跳过
			c.Send(fmt.Printf("a"))
		}
		content, ok := candidateMap["Content"].(map[string]interface{})
		if !ok {
			c.Send(fmt.Printf("b"))
		}
		partsArray, ok := content["Parts"].([]interface{})
		if !ok {
			c.Send(fmt.Printf("c"))
		}
		for _, part := range partsArray {
			c.Send(part, &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
		}
	}
	return

}
