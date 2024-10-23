package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Code int            `json:"code"`
	Data map[string]any `json:"data"`
	Msg  string         `json:"msg"`
}

type TextRequest struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewTextRequest(text string) *TextRequest {
	req := &TextRequest{
		MsgType: "text",
	}
	req.Content.Text = text
	return req
}

type InteractiveMsg struct {
	MsgType string `json:"msg_type"`
	Card    Card   `json:"card"`
}

type Card struct {
	Elements []Element `json:"elements"`
	Header   Header    `json:"header"`
}

type Element struct {
	Tag     string   `json:"tag"`
	Text    *Title   `json:"text,omitempty"`
	Actions []Action `json:"actions,omitempty"`
}

type Action struct {
	Tag   string `json:"tag"`
	Text  Title  `json:"text"`
	URL   string `json:"url"`
	Type  string `json:"type"`
	Value Value  `json:"value"`
}

type Title struct {
	Content string `json:"content"`
	Tag     string `json:"tag"`
}

type Value struct {
}

type Header struct {
	Title Title `json:"title"`
}

func NewInteractiveMsg(card Card) *InteractiveMsg {
	req := &InteractiveMsg{
		MsgType: "interactive",
	}
	req.Card = card
	return req
}

func SendTextMsg(webhook, text string) error {
	param := NewTextRequest(text)
	return sendRequest(param, webhook)
}

func SendInteractiveMsg(webhook string, card Card) error {
	msg := NewInteractiveMsg(card)
	return sendRequest(msg, webhook)
}

func sendRequest(data any, webhook string) error {
	bs, _ := json.Marshal(data)
	request, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(bs))
	if err != nil {
		return fmt.Errorf("create fs request failed: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("send fs request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send fs request failed: %s", resp.Status)
	}

	var respData Response
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return fmt.Errorf("decode fs response failed: %w", err)
	}

	if respData.Code != 0 || respData.Msg != "success" {
		return fmt.Errorf("send fs request failed: %s", respData.Msg)
	}

	return nil
}
