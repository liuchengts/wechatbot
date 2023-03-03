package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qingconglaixueit/wechatbot/config"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const BASEURL_TURBO = "https://api.openai.com/v1/"

// ChatGPTResponseBodyTurbo 响应体
type ChatGPTResponseBodyTurbo struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Choices []ChoiceItemTurbo      `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChoiceItemTurbo struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatGPTRequestBodyTurbo 请求体
type ChatGPTRequestBodyTurbo struct {
	Model            string    `json:"model"`
	MaxTokens        uint      `json:"max_tokens"`
	Temperature      float64   `json:"temperature"`
	TopP             int       `json:"top_p"`
	FrequencyPenalty int       `json:"frequency_penalty"`
	PresencePenalty  int       `json:"presence_penalty"`
	Messages         []Message `json:"messages"`
}

// gpt-3.5-turbo
func CompletionsTurbo(msg string) (string, error) {
	cfg := config.LoadConfig()
	var messages []Message
	var message Message
	message.Role = "user"
	message.Content = msg
	messages = append(messages, message)
	requestBody := ChatGPTRequestBodyTurbo{
		Model:            cfg.Model,
		MaxTokens:        cfg.MaxTokens,
		Temperature:      cfg.Temperature,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Messages:         messages,
	}
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	logger.Info(fmt.Sprintf("request gpt json string : %v", string(requestData)))
	req, err := http.NewRequest("POST", BASEURL_TURBO+"chat/completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	apiKey := config.LoadConfig().ApiKey
	proxyUrl := config.LoadConfig().ProxyUrl
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{Timeout: 30 * time.Second, Transport: transport}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return "", errors.New(fmt.Sprintf("请求GTP出错了，gpt api status code not equals 200,code is %d ,details:  %v ", response.StatusCode, string(body)))
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	logger.Info(fmt.Sprintf("response gpt json string : %v", string(body)))

	gptResponseBody := &ChatGPTResponseBodyTurbo{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Message.Content
	}
	logger.Info(fmt.Sprintf("gpt response text: %s ", reply))
	return reply, nil
}
