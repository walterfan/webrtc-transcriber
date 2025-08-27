package transcribe

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// BaiduTranscriber is the implementation of the transcribe.Service,
// using Baidu Speech Recognition API for speech recognition
type BaiduTranscriber struct {
	appID     string
	apiKey    string
	secretKey string
	ctx       context.Context
}

// BaiduStream implements the transcribe.Stream interface,
// it handles the WebSocket connection to Baidu Speech API
type BaiduStream struct {
	conn    *websocket.Conn
	results chan Result
	ctx     context.Context
}

// Baidu Speech API message structures
type baiduSpeechRequest struct {
	Type string `json:"type"`
	Data struct {
		Audio   string `json:"audio"`
		Format  string `json:"format"`
		Rate    int    `json:"rate"`
		Channel int    `json:"channel"`
		Cuid    string `json:"cuid"`
		Token   string `json:"token"`
		DevPid  int    `json:"dev_pid"`
	} `json:"data"`
}

type baiduSpeechResponse struct {
	Type   string `json:"type"`
	Result struct {
		Text string `json:"text"`
	} `json:"result"`
	Error int `json:"error"`
}

// CreateStream creates a new transcription stream
func (b *BaiduTranscriber) CreateStream() (Stream, error) {
	// Get access token
	token, err := b.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Generate WebSocket URL for Baidu Speech API
	wsURL := fmt.Sprintf("wss://vop.baidu.com/realtime_asr?sn=%s&token=%s", b.generateSN(), token)

	// Create WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Baidu Speech API: %w", err)
	}

	stream := &BaiduStream{
		conn:    conn,
		results: make(chan Result, 10),
		ctx:     b.ctx,
	}

	// Start listening for responses
	go stream.listenForResults()

	return stream, nil
}

// Results returns a channel that will receive the transcription results
func (bs *BaiduStream) Results() <-chan Result {
	return bs.results
}

// Close sends an end-of-stream marker and closes the WebSocket connection
func (bs *BaiduStream) Close() error {
	// Send end-of-stream marker
	endMsg := map[string]interface{}{
		"type": "audio.end",
	}

	endBytes, err := json.Marshal(endMsg)
	if err != nil {
		log.Printf("Warning: failed to marshal end message: %v", err)
	} else {
		if err := bs.conn.WriteMessage(websocket.TextMessage, endBytes); err != nil {
			log.Printf("Warning: failed to send end message: %v", err)
		}
	}

	// Close WebSocket connection
	if err := bs.conn.Close(); err != nil {
		log.Printf("Warning: failed to close WebSocket: %v", err)
	}

	// Close results channel
	close(bs.results)

	return nil
}

// Write sends audio data to the Baidu Speech API
func (bs *BaiduStream) Write(buffer []byte) (int, error) {
	// Encode audio data as base64
	audioData := fmt.Sprintf("%x", md5.Sum(buffer)) // Baidu expects hex format

	// Create speech request
	request := baiduSpeechRequest{
		Type: "audio",
	}
	request.Data.Audio = audioData
	request.Data.Format = "pcm"
	request.Data.Rate = 16000
	request.Data.Channel = 1
	request.Data.Cuid = "webrtc_transcriber"
	request.Data.Token = ""    // Will be set by the API
	request.Data.DevPid = 1537 // Mandarin Chinese

	// Marshal request
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send audio data
	if err := bs.conn.WriteMessage(websocket.TextMessage, requestBytes); err != nil {
		return 0, fmt.Errorf("failed to send audio data: %w", err)
	}

	return len(buffer), nil
}

// listenForResults listens for WebSocket messages and processes transcription results
func (bs *BaiduStream) listenForResults() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in Baidu stream listener: %v", r)
		}
	}()

	for {
		select {
		case <-bs.ctx.Done():
			return
		default:
			// Read message
			_, message, err := bs.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			// Parse response
			var response baiduSpeechResponse
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("Failed to unmarshal response: %v", err)
				continue
			}

			// Process different response types
			switch response.Type {
			case "result":
				if response.Result.Text != "" {
					// Send result
					result := Result{
						Text:       response.Result.Text,
						Confidence: 0.9, // Baidu doesn't provide confidence scores
						Final:      true,
					}

					select {
					case bs.results <- result:
						// Result sent successfully
					case <-bs.ctx.Done():
						return
					default:
						// Channel is full, skip this result
						log.Printf("Results channel is full, skipping result")
					}
				}

			case "error":
				if response.Error != 0 {
					log.Printf("Baidu Speech API error: %d", response.Error)
				}

			case "end":
				log.Printf("Baidu Speech API stream ended")
				return
			}
		}
	}
}

// getAccessToken retrieves an access token from Baidu API
func (b *BaiduTranscriber) getAccessToken() (string, error) {
	// Baidu token URL
	tokenURL := "https://aip.baidubce.com/oauth/2.0/token"

	// Create request
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", b.apiKey)
	data.Set("client_secret", b.secretKey)

	// Make request
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", fmt.Errorf("failed to request access token: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("Baidu API error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token received")
	}

	return tokenResp.AccessToken, nil
}

// generateSN generates a unique serial number for the session
func (b *BaiduTranscriber) generateSN() string {
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	return fmt.Sprintf("%s_%s", b.appID, timestamp)
}

// NewBaiduTranscriber creates a new instance of the transcribe.Service that uses Baidu Speech API
func NewBaiduTranscriber(ctx context.Context, appID, apiKey, secretKey string) (Service, error) {
	if appID == "" || apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("appID, apiKey, and secretKey are required")
	}

	return &BaiduTranscriber{
		appID:     appID,
		apiKey:    apiKey,
		secretKey: secretKey,
		ctx:       ctx,
	}, nil
}
