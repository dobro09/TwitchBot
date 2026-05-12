package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
)
//много повтора кода
type TokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
    TokenType   string `json:"token_type"`
}

func GetAppAccessToken(clientID string, secret string) (string, error) {
	url := fmt.Sprintf("https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials", clientID, secret)
	req, err := http.NewRequest("POST", url, nil)
	if err !=nil{
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}
	resp, err:=http.DefaultClient.Do(req)
	if err !=nil{
		return "", fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err !=nil{
		return "", fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
    	return "", fmt.Errorf("Twitch API вернул статус %d: %s", resp.StatusCode, string(body))
	}
	var token TokenResponse
	if err:= json.Unmarshal(body, &token); err != nil{
		return "", fmt.Errorf("ошибка парсинга jsona: %w", err)
	}
	return token.AccessToken, nil
}


type TwitchUserResponse struct {
    Data []TwitchUser `json:"data"`
}
type TwitchUser struct {
    ID    string `json:"id"`
    Login string `json:"login"`
}

func GetBroadcasterID(token string, clientID string, channelName string) (string, error) {
	//url:= fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", channelName)
	url:= "https://api.twitch.tv/helix/users?login=" + url.QueryEscape(channelName)
	req, err:= http.NewRequest("GET", url, nil)
	if err !=nil{
		return "", fmt.Errorf("ошибка создания гет запроса: %w", err)
	}
	req.Header.Add("Client-Id", clientID)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err:= http.DefaultClient.Do(req)
	if err != nil{
		return "", fmt.Errorf("ошибка отправки запроса %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err !=nil{
		return "", fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
    	return "", fmt.Errorf("Twitch API вернул статус %d: %s", resp.StatusCode, string(body))
	}
	var twitch TwitchUserResponse
	if err := json.Unmarshal(body, &twitch); err != nil{
		return "", fmt.Errorf("ошибка парсинга jsona: %w", err)
	}
	if len(twitch.Data) == 0{
		return "", fmt.Errorf("канал %s не найден", channelName)
	}
	return twitch.Data[0].ID, nil
}

type ClipsResponse struct {
    Data []Clip `json:"data"`
}

type Clip struct {
    ID           string `json:"id"`
    URL          string `json:"url"`
    EmbedURL     string `json:"embed_url"`
    Title        string `json:"title"`
    BroadcasterName string `json:"broadcaster_name"`
    CreatorName  string `json:"creator_name"`
    ViewCount    int    `json:"view_count"`
    CreatedAt    string `json:"created_at"`
}

func GetClips(token string, clientID string, broadcasterID string,) ([]string, error) { //без пагинации пока что(?)
	url:= fmt.Sprintf("https://api.twitch.tv/helix/clips?broadcaster_id=%s&first=100", broadcasterID)
	req, err:= http.NewRequest("GET", url, nil)
	if err !=nil{
		return nil, fmt.Errorf("ошибка создания гет запроса: %w", err)
	}
	req.Header.Add("Client-Id", clientID)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err:= http.DefaultClient.Do(req)
	if err != nil{
		return nil, fmt.Errorf("ошибка отправки запроса %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err !=nil{
		return nil, fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
    	return nil, fmt.Errorf("Twitch API вернул статус %d: %s", resp.StatusCode, string(body))
	}
	var clips ClipsResponse
	if err := json.Unmarshal(body, &clips); err != nil{
		return nil, fmt.Errorf("ошибка парсинга jsona: %w", err)
	}
	if len(clips.Data) == 0{
		return nil, fmt.Errorf("клипов на канале не найдено")
	}
	var clipsID []string
	for _,val := range clips.Data{
		clipsID = append(clipsID, val.URL)
	}
	return clipsID, nil
}

func RandomClip( urls []string) (string, error) {
	if len(urls) == 0 {
		return "", fmt.Errorf("нет клипов")
	}
	return  urls[rand.IntN(len(urls))], nil
}