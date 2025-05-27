package andy

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Largeb0525/personal-tool/database"
	quicknode "github.com/Largeb0525/personal-tool/internal/external/quickNode"
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
	platform := getPlatform(c)
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Platform not supported"})
		return
	}
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

	walletMsg := ""
	energyMsg := ""
	threshold := 3000.0
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
		energyMsg, orderID, askEnergySuccess = AskEnergy(transactionData.ToAddress)
		// ask energy
	} else {
		// wait trongrid 7s
		time.Sleep(time.Second * 7)
		walletUsdt, err := CheckTronAddressUSDT(transactionData.ToAddress)
		if err != nil {
			walletMsg = "[error] " + err.Error()
		} else if walletUsdt == "" {
			walletMsg = "wallet has no USDT"
		} else {
			walletUsdtFloat, err := strconv.ParseFloat(walletUsdt, 64)
			if err != nil {
				walletMsg = "[error] " + err.Error()
			} else {
				if walletUsdtFloat >= threshold {
					walletMsg = walletUsdt
					energyMsg, orderID, askEnergySuccess = AskEnergy(transactionData.ToAddress)
				} else {
					walletMsg = walletUsdt
					energyMsg = "pass"
				}
			}
		}
	}
	go func() {
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
	}()

	message := fmt.Sprintf(
		`🟢 %s Transaction Notification
%s
Amount: %s
From: %s
To: %s
Wallet Msg: %s
Energy Msg: %s`,
		name, transactionData.URL, transactionData.USDT, transactionData.FromAddress, transactionData.ToAddress, walletMsg, energyMsg)

	err = telegram.SendTelegramMessage(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
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

// TODO complete function
func refreshHandler(c *gin.Context) {
	db := database.GetDB()
	addresses, err := database.GetAddressesByPlatform(db, "b")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for addr := range addresses {
		walletUsdt, err := CheckTronAddressUSDT(addr)
		if err != nil {
			if err.Error() != "wallet has no USDT" {
				fmt.Printf("Error checking address:%s ,err:%v\n", addr, err)
			}
			continue
		}
		walletUsdtFloat, err := strconv.ParseFloat(walletUsdt, 64)
		if err != nil {
			fmt.Printf("Error parsing address:%s ,err:%v\n", addr, err)
			time.Sleep(time.Second * 1)
			continue
		}
		if walletUsdtFloat >= 10 {
			AskEnergy(addr)
			time.Sleep(time.Second * 1)
		}
	}
	time.Sleep(time.Second * 1)
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

	if updateAlert == "true" {
		go func() {
			quickAlert, err := quicknode.GetQuickAlertInfo()
			if err != nil {
				log.Println(err)
				return
			}
			addresses := quicknode.ParseExpressionToAddresses(quickAlert.Expression)
			addresses = append(addresses, newHexAddresses...)
			err = quicknode.PatchQuickAlert(addresses)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("PatchQuickAlert success")
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "newAddress": newAddresses, "insertFailedAddresses": insertFailedAddresses})
}
