package andy

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type RequestPayload struct {
	Amount   string `json:"amount" binding:"required"`
	Platform string `json:"platform" binding:"required"`
	Method   string `json:"method" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type Payload struct {
	Merchant_id       string `json:"merchant_id"`
	Merchant_order_id string `json:"merchant_order_id"`
	Amount            string `json:"amount"`
	Notify_url        string `json:"notify_url"`
	Payer             string `json:"payer"`
	Payment_method    string `json:"payment_method"`
	Apply_timestamp   int64  `json:"apply_timestamp"`
	Md5_sign          string `json:"md5_sign"`
}

func handlePostRequest(c *gin.Context) {
	var payload RequestPayload

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize request payload"})
		return
	}

	if payload.Platform != "b" && payload.Platform != "j" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Platform not supported"})
		return
	}
	url := viper.GetString(fmt.Sprintf("andy.%s.url", payload.Platform))

	body, err := prepareReqBody(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	c.JSON(resp.StatusCode, gin.H{
		"status":  resp.Status,
		"body":    string(respBody),
		"headers": resp.Header,
	})
}

func prepareReqBody(req RequestPayload) (body []byte, err error) {
	reqBody := Payload{
		Merchant_id:       viper.GetString(fmt.Sprintf("andy.%s.merchant_id", req.Platform)),
		Merchant_order_id: fmt.Sprintf("test%s", time.Now().Format("0102150405")),
		Amount:            req.Amount,
		Notify_url:        "https://example.com",
		Payer:             req.Name,
		Payment_method:    req.Method,
		Apply_timestamp:   time.Now().UnixMilli(),
	}
	apiKey := viper.GetString(fmt.Sprintf("andy.%s.api_key", req.Platform))

	reqBody.Md5_sign = md5Encode(reqBody, apiKey)
	body, err = json.Marshal(reqBody)
	return
}

func md5Encode(payload Payload, apiKey string) string {
	str := fmt.Sprintf(`{"amount":"%s","api_key":"%s","apply_timestamp":%d,"merchant_id":"%s","merchant_order_id":"%s","notify_url":"%s","payer":"%s","payment_method":"%s"}`,
		payload.Amount,
		apiKey,
		payload.Apply_timestamp,
		payload.Merchant_id,
		payload.Merchant_order_id,
		payload.Notify_url,
		payload.Payer,
		payload.Payment_method,
	)

	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
