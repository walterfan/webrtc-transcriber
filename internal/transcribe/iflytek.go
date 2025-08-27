package transcribe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// IflyTekTranscriber is the implementation of the transcribe.Service,
// using Xunfei's WebSocket API for speech recognition
type IflyTekTranscriber struct {
	appID     string
	apiKey    string
	apiSecret string
	appUrl    string
	ctx       context.Context
}

// IflyTekStream implements the transcribe.Stream interface,
// it should map one to one with the audio stream coming from the client
type IflyTekStream struct {
	conn        *websocket.Conn
	results     chan Result
	ctx         context.Context
	transcriber *IflyTekTranscriber
}

// Xunfei API request/response structures
type XunfeiRequest struct {
	Common   XunfeiCommon   `json:"common"`
	Business XunfeiBusiness `json:"business"`
	Data     XunfeiData     `json:"data"`
}

type XunfeiCommon struct {
	AppID string `json:"app_id"`
}

type XunfeiBusiness struct {
	Language string `json:"language"`
	Domain   string `json:"domain"`
	VAD      int    `json:"vad_eos"`
	// Removed unsupported fields: Format, SampleRate, Channel, Punctuation, DynamicCorrection
}

type XunfeiData struct {
	Status   int    `json:"status"`
	Format   string `json:"format"`
	Audio    string `json:"audio"`
	Encoding string `json:"encoding"`
}

type XunfeiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Result struct {
			Ws []struct {
				Cw []struct {
					W string `json:"w"`
				} `json:"cw"`
			} `json:"ws"`
		} `json:"result"`
		Status int `json:"status"`
	} `json:"data"`
}

// CreateStream creates a new transcription stream
func (t *IflyTekTranscriber) CreateStream() (Stream, error) {
	// Generate authentication URL
	authURL, err := t.generateAuthURL()
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth URL: %w", err)
	}

	// Connect to WebSocket
	log.Printf("Attempting to connect to Xunfei WebSocket: %s", authURL)
	conn, resp, err := websocket.DefaultDialer.Dial(authURL, nil)
	if err != nil {
		if resp != nil {
			log.Printf("WebSocket connection failed with HTTP status: %d", resp.StatusCode)
			log.Printf("Response headers: %v", resp.Header)

			// Try to read response body for more error details
			if resp.Body != nil {
				defer resp.Body.Close()
				bodyBytes, readErr := io.ReadAll(resp.Body)
				if readErr == nil {
					log.Printf("Response body: %s", string(bodyBytes))
				}
			}
		}
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	log.Printf("Successfully connected to Xunfei WebSocket")

	// Send initial configuration
	config := XunfeiRequest{
		Common: XunfeiCommon{
			AppID: t.appID,
		},
		Business: XunfeiBusiness{
			Language: "zh_cn", // Chinese by default
			Domain:   "iat",
			VAD:      3000, // Voice activity detection end-of-speech timeout
		},
		Data: XunfeiData{
			Status:   0, // Start of audio stream
			Format:   "audio/L16;rate=48000",
			Encoding: "raw",
		},
	}

	log.Printf("Sending Xunfei configuration: AppID=%s, Language=%s, Domain=%s, VAD=%d",
		config.Common.AppID, config.Business.Language, config.Business.Domain, config.Business.VAD)

	configBytes, err := json.Marshal(config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	log.Printf("Sending config message: %s", string(configBytes))
	if err := conn.WriteMessage(websocket.TextMessage, configBytes); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send config: %w", err)
	}
	log.Printf("Config message sent successfully")

	stream := &IflyTekStream{
		conn:        conn,
		results:     make(chan Result),
		ctx:         t.ctx,
		transcriber: t,
	}

	// Start listening for responses in background
	go stream.listenForResults()

	return stream, nil
}

// Results returns a channel that will receive the transcription
// results when they're ready
func (st *IflyTekStream) Results() <-chan Result {
	return st.results
}

// Close flushes the recognition stream and
// pipes the results to the channel
func (st *IflyTekStream) Close() error {
	// Send end-of-stream marker
	endData := XunfeiData{
		Status:   2, // End of audio stream
		Format:   "audio/L16;rate=48000",
		Audio:    "",
		Encoding: "raw",
	}

	endRequest := XunfeiRequest{
		Common: XunfeiCommon{
			AppID: st.transcriber.appID, // Use the actual AppID from the transcriber
		},
		Business: XunfeiBusiness{
			Language: "zh_cn",
			Domain:   "iat",
			VAD:      3000,
		},
		Data: endData,
	}

	endBytes, err := json.Marshal(endRequest)
	if err == nil {
		st.conn.WriteMessage(websocket.TextMessage, endBytes)
	}

	// Close WebSocket connection
	if err := st.conn.Close(); err != nil {
		log.Printf("Error closing WebSocket: %v", err)
	}

	// Close results channel
	close(st.results)
	return nil
}

func (st *IflyTekStream) Write(buffer []byte) (int, error) {
	// Send audio data
	audioData := XunfeiData{
		Status:   1, // Audio data
		Format:   "audio/L16;rate=48000",
		Audio:    base64.StdEncoding.EncodeToString(buffer),
		Encoding: "raw",
	}

	request := XunfeiRequest{
		Common: XunfeiCommon{
			AppID: st.transcriber.appID, // Use the actual AppID from the transcriber
		},
		Business: XunfeiBusiness{
			Language: "zh_cn",
			Domain:   "iat",
			VAD:      3000,
		},
		Data: audioData,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal audio request: %w", err)
	}

	if err := st.conn.WriteMessage(websocket.TextMessage, requestBytes); err != nil {
		return 0, fmt.Errorf("failed to send audio data: %w", err)
	}

	return len(buffer), nil
}

// listenForResults listens for WebSocket messages and processes transcription results
func (st *IflyTekStream) listenForResults() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in listenForResults: %v", r)
		}
	}()

	for {
		select {
		case <-st.ctx.Done():
			return
		default:
			_, message, err := st.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			var response XunfeiResponse
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("Failed to unmarshal response: %v", err)
				continue
			}

			// Check for errors
			if response.Code != 0 {
				log.Printf("Xunfei API error: %s", response.Message)
				continue
			}

			// Process results
			if response.Data.Status == 2 { // Final result
				text := ""
				for _, ws := range response.Data.Result.Ws {
					for _, cw := range ws.Cw {
						text += cw.W
					}
				}

				if text != "" {
					st.results <- Result{
						Text:       text,
						Confidence: 0.9, // Xunfei doesn't provide confidence scores in this format
						Final:      true,
					}
				}
			} else if response.Data.Status == 1 { // Partial result
				text := ""
				for _, ws := range response.Data.Result.Ws {
					for _, cw := range ws.Cw {
						text += cw.W
					}
				}

				if text != "" {
					st.results <- Result{
						Text:       text,
						Confidence: 0.8, // Partial results have lower confidence
						Final:      false,
					}
				}
			}
		}
	}
}

// HmacWithShaTobase64 creates HMAC-SHA256 signature and returns base64 encoded result
func HmacWithShaTobase64(algorithm string, data string, secret string) string {
	var hashFunc func() hash.Hash
	switch algorithm {
	case "hmac-sha256":
		hashFunc = sha256.New
	default:
		hashFunc = sha256.New
	}

	h := hmac.New(hashFunc, []byte(secret))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// @hosturl :  like  wss://iat-api.xfyun.cn/v2/iat
// @apikey : apiKey
// @apiSecret : apiSecret
func assembleAuthUrl(hosturl string, apiKey, apiSecret string) string {
	ul, err := url.Parse(hosturl)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return ""
	}

	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)

	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}

	//拼接签名字符串
	sgin := strings.Join(signString, "\n")

	log.Printf("AssembleAuthUrl - Sign string: %q", sgin)
	log.Printf("AssembleAuthUrl - API Secret length: %d", len(apiSecret))

	//签名结果
	sha := HmacWithShaTobase64("hmac-sha256", sgin, apiSecret)
	log.Printf("AssembleAuthUrl - Generated signature: %s", sha)

	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("api_key=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)

	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))
	log.Printf("AssembleAuthUrl - Authorization (base64): %s", authorization)

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)

	//将编码后的字符串url encode后添加到url后面
	callurl := hosturl + "?" + v.Encode()
	log.Printf("AssembleAuthUrl - Final URL: %s", callurl)

	return callurl
}

// generateAuthURL generates the authenticated WebSocket URL for Xunfei API
func (t *IflyTekTranscriber) generateAuthURL() (string, error) {
	// Use appUrl from struct, fallback to default if not set
	baseURL := t.appUrl
	if baseURL == "" {
		baseURL = "wss://iat-api.xfyun.cn/v2/iat"
	}

	// Use the working assembleAuthUrl function from Xunfei documentation
	authURL := assembleAuthUrl(baseURL, t.apiKey, t.apiSecret)
	log.Printf("Generated Xunfei auth URL using assembleAuthUrl: %s", authURL)
	return authURL, nil
}

// NewIflyTekTranscriber creates a new instance of the transcribe.Service that uses
// Xunfei's speech recognition API
func NewIflyTekTranscriber(ctx context.Context, appID, apiKey, apiSecret, appUrl string) (Service, error) {
	if appID == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("appID, apiKey, and apiSecret are required")
	}

	return &IflyTekTranscriber{
		appID:     appID,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		appUrl:    appUrl,
		ctx:       ctx,
	}, nil
}
