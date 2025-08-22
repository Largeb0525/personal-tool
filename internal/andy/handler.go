package andy

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/Largeb0525/personal-tool/database"
	"github.com/Largeb0525/personal-tool/internal/external/quickNode"
	"github.com/Largeb0525/personal-tool/internal/external/telegram"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func handleOrderRequestHandler(c *gin.Context) {
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

	body, err := prepareOrderReqBody(payload)
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

func quickAlertsEventHandler(c *gin.Context) {
	var payloads MatchingReceipts
	if err := c.ShouldBindJSON(&payloads); err != nil || len(payloads.MatchingReceipts) == 0 {
		log.Printf("Invalid payload: %v", err)
		c.JSON(http.StatusOK, gin.H{"message": "received"})
		return
	}

	platform := getPlatform(c)
	if platform == "" {
		log.Printf("Platform not supported")
		c.JSON(http.StatusOK, gin.H{"message": "received"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "received"})

	go func(payload TransactionReceipt, platform string) {
		transactionData, err := parseTransactionData(payload)
		if err != nil {
			log.Printf("Failed to parse transaction data: %v", err)
			return
		}

		usdtFloat, err := strconv.ParseFloat(transactionData.USDT, 64)
		if err != nil {
			log.Printf("Failed to parse USDT: %v", err)
			return
		}

		walletMsg := ""
		energyMsg := ""
		threshold := 3500.0
		name := ""
		orderID := ""
		askEnergySuccess := false

		switch platform {
		case "b":
			threshold = CollectAmong
			name = BName
		case "i":
			threshold = CollectAmongIndia
			name = IName
		}

		if usdtFloat >= threshold {
			walletMsg = "amount >= threshold , ask energy"
			energyMsg, orderID, askEnergySuccess, err = delegateEnergy(transactionData.ToAddress)
			if err != nil {
				log.Printf("Error while delegating energy: %v", err)
				energyMsg, orderID, askEnergySuccess = AskEnergy(transactionData.ToAddress)
			}
		} else {
			walletUsdt, err := getAddressUSDT(transactionData.ToAddress)
			if err != nil {
				walletMsg = "[error] " + err.Error()
			} else if walletUsdt == big.NewFloat(0) {
				walletMsg = "wallet has no USDT"
			} else {
				walletUsdtFloat, _ := walletUsdt.Float64()
				if walletUsdtFloat >= threshold {
					walletMsg = walletUsdt.String()
					energyMsg, orderID, askEnergySuccess, err = delegateEnergy(transactionData.ToAddress)
					if err != nil {
						log.Printf("Error while delegating energy: %v", err)
						energyMsg, orderID, askEnergySuccess = AskEnergy(transactionData.ToAddress)
					}
				} else {
					walletMsg = walletUsdt.String()
					energyMsg = "pass"
				}
			}
		}
		db := database.GetDB()
		database.InsertEventHistory(db, database.EventHistory{
			TransactionHash:  transactionData.TransactionHash,
			USDT:             transactionData.USDT,
			FromAddress:      transactionData.FromAddress,
			ToAddress:        transactionData.ToAddress,
			Platform:         platform,
			WalletUSDT:       walletMsg,
			OrderID:          orderID,
			AskEnergySuccess: askEnergySuccess,
		})

		message := fmt.Sprintf(
			`🟢 %s Transaction Notification
%s
Amount: %s
From: %s
To: %s
Wallet Msg: %s
Energy Msg: %s`,
			name, transactionData.URL, transactionData.USDT, transactionData.FromAddress, transactionData.ToAddress, walletMsg, energyMsg)

		err = telegram.SendTelegramMessage(message, telegram.TelegramChatId, telegram.TelegramBotToken)
		if err != nil {
			log.Printf("Failed to send Telegram message: %v", err)
			return
		}
		// TODO
		if transactionData.ToAddress == "TYKopCHtD2dCXH4FdD3YchwagzuMGHg5pG" ||
			transactionData.ToAddress == "TNJSQNEVgxoh5fhuJWz1tA9xNtaN373DbV" ||
			transactionData.ToAddress == "TJyd1LWs93hpuhcLshY9woojT35rJY8dH" ||
			transactionData.ToAddress == "TEhXdCzMJxYK38Vw3w1htdcmQC3RKTh7Rp" ||
			transactionData.ToAddress == "TMmSstecCjztrkYwqFPcCjUBogz6uJ2aVw" ||
			transactionData.ToAddress == "TBHhUnzCQQP4pFk4Tm4DGjaeZqaZVpcn6T" {
			err = telegram.SendTelegramMessage(message, telegram.CriticalTelegramChatId, telegram.TelegramBotToken)
			if err != nil {
				log.Printf("Failed to send Critical Telegram message: %v", err)
				return
			}
		}
	}(payloads.MatchingReceipts[0], platform)
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
	db := database.GetDB()
	countMap, err := database.GetTodayEventCountGroupByPlatform(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{BName: countMap["b"], IName: countMap["i"]})
}

func refreshHandler(c *gin.Context) {
	var payload thresholdRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	platform := c.Param("platform")
	db := database.GetDB()
	addresses, err := database.GetAddressesByPlatform(db, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var threshold float64
	switch platform {
	case "b":
		threshold = payload.BThreshold
	case "i":
		threshold = payload.IThreshold
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform not supported"})
		return
	}

	for addr := range addresses {
		walletUsdt, err := getAddressUSDT(addr)
		if err != nil {
			if err.Error() != "wallet has no USDT" {
				fmt.Printf("Error checking address:%s ,err:%v\n", addr, err)
			}
			continue
		}
		log.Printf("Address: %s, USDT: %s\n", addr, walletUsdt.Text('f', 6))
		if walletUsdt == big.NewFloat(0) {
			continue
		}

		walletUsdtFloat, _ := walletUsdt.Float64()
		if walletUsdtFloat >= threshold {
			AskEnergy(addr)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func uploadAddressCsvFileHandler(c *gin.Context) {
	platform := getPlatform(c)
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Platform not supported"})
		return
	}
	updateAlert := c.DefaultQuery("updateAlert", "false")
	file, err := c.FormFile("csv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	reader := csv.NewReader(bufio.NewReader(f))
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid csv format"})
		return
	}

	db := database.GetDB()
	dbAddress, err := database.GetAddressesByPlatform(db, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newAddresses := make([]string, 0)
	newHexAddresses := make([]string, 0)
	insertFailedAddresses := make([]string, 0)

	for i, row := range records {
		if i == 0 || len(row) == 0 {
			continue
		}
		addr := strings.TrimSpace(row[0])
		if addr == "" {
			continue
		}
		_, ok := dbAddress[addr]
		if ok {
			continue
		}
		hexAddr, err := TronToHexPadded32(addr)
		if err != nil {
			log.Println(err)
			insertFailedAddresses = append(insertFailedAddresses, addr)
			continue
		}
		err = database.InsertAddress(db, addr, hexAddr, platform)
		if err != nil {
			log.Println(err)
			insertFailedAddresses = append(insertFailedAddresses, addr)
			continue
		}
		newAddresses = append(newAddresses, addr)
		newHexAddresses = append(newHexAddresses, hexAddr)
	}

	if len(newAddresses) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "success", "newAddress": newAddresses, "insertFailedAddresses": insertFailedAddresses})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "newAddress": newAddresses, "insertFailedAddresses": insertFailedAddresses})

	if updateAlert == "true" {
		patchAddrs := append([]string(nil), newHexAddresses...)
		go func(addrs []string) {
			quickAlert, err := quickNode.GetQuickAlertInfo()
			if err != nil {
				log.Println(err)
				return
			}
			addresses := quickNode.ParseExpressionToAddresses(quickAlert.Expression)
			addresses = append(addresses, addrs...)
			err = quickNode.PatchQuickAlert(addresses)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("PatchQuickAlert success")
		}(patchAddrs)
	}
}

func freezeTRXHandler(c *gin.Context) {
	var payload struct {
		Trx int64 `json:"trx"`
	}

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := freezeBalance(payload.Trx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sign, err := signTransaction(tx.TxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	broadcastReq := quickNode.BroadcastRequest{
		TxID:       tx.TxID,
		RawData:    tx.RawData,
		RawDataHex: tx.RawDataHex,
		Signature:  sign,
		Visible:    true,
	}
	resp, err := quickNode.BroadcastTransaction(broadcastReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func delegateResourceHandler(c *gin.Context) {
	var payload struct {
		Address string `json:"address"`
	}

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := delegateResource(payload.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sign, err := signTransaction(tx.TxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	broadcastReq := quickNode.BroadcastRequest{
		TxID:       tx.TxID,
		RawData:    tx.RawData,
		RawDataHex: tx.RawDataHex,
		Signature:  sign,
		Visible:    true,
	}
	resp, err := quickNode.BroadcastTransaction(broadcastReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func undelegateResourceHandler(c *gin.Context) {
	var payload struct {
		Address string `json:"address"`
	}

	err := c.ShouldBindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := undelegateResource(payload.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sign, err := signTransaction(tx.TxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	broadcastReq := quickNode.BroadcastRequest{
		TxID:       tx.TxID,
		RawData:    tx.RawData,
		RawDataHex: tx.RawDataHex,
		Signature:  sign,
		Visible:    true,
	}
	resp, err := quickNode.BroadcastTransaction(broadcastReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func getAllAddressUSDTHandler(c *gin.Context) {
	db := database.GetDB()
	addressMap, err := database.GetAddressesByPlatform(db, "b")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total := big.NewFloat(0)

	for address := range addressMap {
		balance, err := getAddressUSDT(address)
		if err != nil {
			log.Printf("❌ 查詢失敗 %s: %v", address, err)
			continue
		}
		fmt.Printf("%s ➜ %s USDT\n", address, balance.Text('f', 6))
		total.Add(total, balance)
	}

	c.JSON(http.StatusOK, total.Text('f', 6))
}
