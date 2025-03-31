package andy

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	c.Status(resp.StatusCode)

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write response body"})
	}
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

func eventHandler(c *gin.Context) {
	var payloads []TransactionReceipt

	err := c.ShouldBindJSON(&payloads)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(payloads) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No payloads provided"})
		return
	}

	transactionData, err := parseTransactionData(payloads[0])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	usdtFloat, err := strconv.ParseFloat(transactionData.USDT, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	platform := c.Param("platform")
	usdtMsg := ""
	energyMsg := ""
	threshold := 3000.0
	name := ""

	switch platform {
	case "b":
		threshold = CollectAmong
		name = BName
	case "i":
		threshold = CollectAmongIndia
		name = IName
	}

	if usdtFloat >= threshold {
		usdtMsg = "amount >= threshold , ask energy"
		energyMsg = AskEnergy(transactionData.ToAddress, platform)
		// ask energy
	} else {
		// wait trongrid 7s
		time.Sleep(time.Second * 7)
		walletUsdt, err := CheckTronAddressUSDT(transactionData.ToAddress)
		if err != nil {
			usdtMsg = "[error] " + err.Error()
		} else {
			walletUsdtFloat, err := strconv.ParseFloat(walletUsdt, 64)
			if err != nil {
				usdtMsg = "[error] " + err.Error()
			} else {
				if walletUsdtFloat >= threshold {
					usdtFloat = walletUsdtFloat
					usdtMsg = walletUsdt
					energyMsg = AskEnergy(transactionData.ToAddress, platform)
				} else {
					usdtMsg = walletUsdt
					energyMsg = "pass"
				}
			}
		}
	}

	message := fmt.Sprintf(
		`🟢 %s Transaction Notification
Amount: %s
From: %s
To: %s
Wallet Msg: %s
Energy Msg: %s`,
		name, transactionData.USDT, transactionData.FromAddress, transactionData.ToAddress, usdtMsg, energyMsg)

	payload := TelegramMessagePayload{
		ChatID: TelegramChatId,
		Text:   message,
		ReplyMarkup: InlineKeyboardMarkup{
			InlineKeyboard: [][]InlineKeyboardButton{
				{
					{
						Text: "TRONSCAN",
						URL:  transactionData.URL,
					},
				},
			},
		},
	}
	err = SendTelegramMessage(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func parseTransactionData(receipt TransactionReceipt) (ParsedTransaction, error) {
	var parsedTransaction ParsedTransaction
	if len(receipt.Logs) == 0 {
		return parsedTransaction, errors.New("no logs in transaction receipt")
	}
	logData := receipt.Logs[0]

	// 擷取交易哈希（移除前綴 0x）
	txHash := strings.TrimPrefix(receipt.TransactionHash, "0x")
	url := tronscanURL + txHash

	// 解析 USDT 數值（hex -> int -> 除以 1e6）
	dataHex := strings.TrimPrefix(logData.Data, "0x")
	amount := new(big.Int)
	amount.SetString(dataHex, 16)
	usdt := new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(1e6))

	// 解析地址，從 topic[1] 和 topic[2] 拿出末 40 位（地址）
	if len(logData.Topics) < 3 {
		return parsedTransaction, errors.New("not enough topics in log")
	}
	fromHex := "41" + logData.Topics[1][26:]
	toHex := "41" + logData.Topics[2][26:]

	fromAddr, err := toBase58CheckAddress(fromHex)
	if err != nil {
		return parsedTransaction, fmt.Errorf("invalid from address: %w", err)
	}
	toAddr, err := toBase58CheckAddress(toHex)
	if err != nil {
		return parsedTransaction, fmt.Errorf("invalid to address: %w", err)
	}

	usdtStr := usdt.Text('f', 6)

	parsedTransaction = ParsedTransaction{
		TransactionHash: txHash,
		URL:             url,
		USDT:            usdtStr,
		FromAddress:     fromAddr,
		ToAddress:       toAddr,
	}
	return parsedTransaction, nil
}

func toBase58CheckAddress(hexAddr string) (string, error) {
	if len(hexAddr) >= 2 && hexAddr[:2] == "0x" {
		hexAddr = hexAddr[2:]
	}

	addrBytes, err := hex.DecodeString(hexAddr)
	if err != nil {
		return "", errors.New("invalid hex address")
	}

	if len(addrBytes) != 21 {
		return "", errors.New("address should be 21 bytes including '41' prefix")
	}

	// Double SHA256 checksum
	first := sha256.Sum256(addrBytes)
	second := sha256.Sum256(first[:])
	checksum := second[:4]

	full := append(addrBytes, checksum...)
	return base58Encode(full), nil
}

func CheckTronAddressUSDT(address string) (string, error) {
	url := trongridWalletURL + address

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		_ = json.Unmarshal(bodyBytes, &errData)
		description := "No error description available"
		if desc, ok := errData["description"].(string); ok {
			description = desc
		}
		log.Printf("Failed to get balance: %d, %s\n", resp.StatusCode, description)
		return "", errors.New("get balance fail")
	}

	var tronResp TronWalletFullResponse
	if err := json.Unmarshal(bodyBytes, &tronResp); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	balance := getUSDTBalance(tronResp.Data)
	if balance != "" {
		return balance, nil
	}
	return "", errors.New("wallet has no USDT")
}

func getUSDTBalance(tokens []TokenInfo) string {
	for _, token := range tokens {
		if token.TokenAbbr == "USDT" {
			return token.Balance
		}
	}
	return ""
}

func AskEnergy(address string, platform string) (energyMsg string) {
	// 準備 form data
	form := url.Values{}
	form.Set("address", address)
	form.Set("token", EnergyToken)

	// 建立 request
	req, err := http.NewRequest("POST", EnergyUrl, bytes.NewBufferString(form.Encode()))
	if err != nil {
		energyMsg = fmt.Sprintf("failed to create request: %s", err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 發送 request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		energyMsg = fmt.Sprintf("request failed: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	// 讀取 response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		energyMsg = fmt.Sprintf("failed to read response: %s", err.Error())
		return
	}

	if resp.StatusCode == http.StatusOK {
		var result AskEnergyResponse
		err := json.Unmarshal(body, &result)
		if err != nil {
			energyMsg = fmt.Sprintf("resp failed to parse JSON: %s", err.Error())
			return
		}
		energyMsg = fmt.Sprintf("success, order id: %s", result.OrderID)
		switch platform {
		case "b":
			BDailyCount++
		case "i":
			IDailyCount++
		}
		return
	}
	energyMsg = fmt.Sprintf("request failed with status code: %d", resp.StatusCode)
	return
}

func SendTelegramMessage(payload TelegramMessagePayload) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf(botReqURL, TelegramBotToken)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var errResp map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&errResp)
	desc := "No error description available"
	if d, ok := errResp["description"].(string); ok {
		desc = d
	}
	log.Printf("Failed to send message: %d, %s\n", resp.StatusCode, desc)
	return fmt.Errorf("telegram error: %s", desc)
}

func thresholdHandler(c *gin.Context) {
	var payload thresholdRequest

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.BThreshold != 0.0 {
		CollectAmong = payload.BThreshold
	}
	if payload.IThreshold != 0.0 {
		CollectAmongIndia = payload.IThreshold
	}

	c.JSON(http.StatusOK, gin.H{"BThreshold": CollectAmong, "IThreshold": CollectAmongIndia})
}

func dailyReportHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{BName: BDailyCount, IName: IDailyCount})
}
