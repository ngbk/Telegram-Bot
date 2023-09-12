package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

func New(host, token string) *Client {
	return &Client{
		host:     host,
		basePath: "bot" + token,
		client:   http.Client{},
	}
}

func (c Client) GetUpdates(offset, limit int) ([]Update, error) {
	params := url.Values{}
	params.Add("offset", strconv.Itoa(offset))
	params.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(params, "getUpdates", nil)
	if err != nil {
		return nil, fmt.Errorf("can't get updates: %w", err)
	}

	var resUpdates ResponseUpdates

	err = json.Unmarshal(data, &resUpdates)
	if err != nil {
		return nil, fmt.Errorf("can't get updates: %w", err)
	}

	return resUpdates.Result, nil
}

func (c Client) SendKeyboard(chatId int, keyboard [][]string) error {
	params := url.Values{}
	params.Add("chat_id", strconv.Itoa(chatId))

	keyRes := [][]InlineKeyboardButton{}

	for _, rows := range keyboard {
		res := []InlineKeyboardButton{}
		for _, item := range rows {
			button := InlineKeyboardButton{
				Text:         item,
				CallbackData: item,
			}
			res = append(res, button)
		}
		keyRes = append(keyRes, res)
	}

	replyMarkup := ReplyMarkup{
		InlineKeyboard: keyRes,
	}
	requestBody, err := json.Marshal(Message{
		Chat:        Chat{Id: chatId},
		Text:        "Выберите",
		ReplyMarkup: replyMarkup,
	})
	if err != nil {
		return fmt.Errorf("failed marshal sendkeyboard': %w", err)
	}
	//
	_, err = c.doRequest(params, "sendMessage", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("can't send keyboard: %w", err)
	}

	return nil
}
func (c Client) SendMessage(chatId int, text string) error {

	params := url.Values{}
	params.Add("chat_id", strconv.Itoa(chatId))
	params.Add("text", text)

	_, err := c.doRequest(params, "sendMessage", nil)
	if err != nil {
		return fmt.Errorf("can't send message: %w", err)
	}
	return nil
}

func (c Client) doRequest(query url.Values, method string, bodyReq io.Reader) ([]byte, error) {
	urlReq := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}

	req, err := http.NewRequest(http.MethodPost, urlReq.String(), bodyReq)
	if err != nil {
		return nil, fmt.Errorf("can't do request: %w", err)
	}
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read request: %w", err)
	}
	return body, nil
}
