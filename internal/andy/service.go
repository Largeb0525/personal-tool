package andy

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Largeb0525/personal-tool/database"
	"github.com/Largeb0525/personal-tool/internal/external/quickNode"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func prepareOrderReqBody(req RequestPayload) (body []byte, err error) {
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

	reqBody.Md5_sign = reqBodyMd5Encode(reqBody, apiKey)
	body, err = json.Marshal(reqBody)
	return
}

func reqBodyMd5Encode(payload Payload, apiKey string) string {
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

func AskEnergy(address string) (energyMsg string, orderID string, success bool) {
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
		orderID = result.OrderID
		success = true
		fmt.Printf("success, address:%s\n", address)
		return
	}
	energyMsg = fmt.Sprintf("request failed with status code: %d", resp.StatusCode)
	return
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
	return balance, nil

}

func getPlatform(c *gin.Context) string {
	p := c.Param("platform")
	switch p {
	case "b", "i", "j":
		return p
	default:
		return ""
	}
}

func freezeBalance(balance int64) (quickNode.Transaction, error) {
	req := quickNode.FreezeRequest{
		OwnerAddress:  EnergyAddress,
		Resource:      "ENERGY",
		FrozenBalance: balance * 1000000,
		Visible:       true,
	}

	tx, err := quickNode.CreateFreezeTx(req)
	if err != nil {
		log.Println(err)
		return tx, err
	}
	log.Printf("freeze tx: %v", tx)
	return tx, err
}

func delegateResource(receiverAddress string) (quickNode.Transaction, error) {
	req := quickNode.DelegateResourceRequest{
		OwnerAddress:    EnergyAddress,
		ReceiverAddress: receiverAddress,
		Balance:         FreezeUnit * 1000000,
		Resource:        "ENERGY",
		Lock:            false,
		Visible:         true,
	}

	tx, err := quickNode.CreateDelegateResourceTx(req)
	if err != nil {
		return tx, err
	}
	if tx.RawData == nil {
		err = errors.New("Balance not enough")
		return tx, err
	}
	return tx, err
}

func undelegateResource(receiverAddress string) (quickNode.Transaction, error) {
	req := quickNode.UndelegateResourceRequest{
		OwnerAddress:    EnergyAddress,
		ReceiverAddress: receiverAddress,
		Balance:         FreezeUnit * 1000000,
		Resource:        "ENERGY",
		Visible:         true,
	}

	tx, err := quickNode.CreateUndelegateResourceTx(req)
	if err != nil {
		return tx, err
	}
	if tx.RawData == nil {
		err = errors.New("undelegate RawData is nil")
		return tx, err
	}
	return tx, err
}

func delegateEnergy(address string) (energyMsg string, orderID string, success bool, err error) {
	tx, err := delegateResource(address)
	if err != nil {
		return
	}

	sign, err := signTransaction(tx.TxID)
	if err != nil {
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
		return
	}
	energyMsg = fmt.Sprintf("delegate success, order id: %s", resp.Txid)
	db := database.GetDB()
	dbErr := database.InsertDelegateRecord(db, address, resp.Txid)
	if dbErr != nil {
		log.Printf("Failed to insert delegate record: %v", err)
	}

	return energyMsg, resp.Txid, true, nil
}

func undelegateEnergy(address string, txID string) (err error) {
	tx, err := undelegateResource(address)
	if err != nil {
		return
	}

	sign, err := signTransaction(tx.TxID)
	if err != nil {
		return
	}

	broadcastReq := quickNode.BroadcastRequest{
		TxID:       tx.TxID,
		RawData:    tx.RawData,
		RawDataHex: tx.RawDataHex,
		Signature:  sign,
		Visible:    true,
	}
	_, err = quickNode.BroadcastTransaction(broadcastReq)
	if err != nil {
		return
	}
	db := database.GetDB()
	dbErr := database.UpdateUndelegatedByTxid(db, txID)
	if dbErr != nil {
		log.Printf("Failed to update undelegated: %v", err)
	}

	return
}

func getAddressUSDT(addr string) (*big.Float, error) {
	hexAddr, err := TronToHexPadded32(addr)
	if err != nil {
		return nil, err
	}
	param := fmt.Sprintf("%064s", hexAddr[2:]) // 去掉 0x 開頭

	req := quickNode.TriggerSmartContractRequest{
		OwnerAddress:     addr,
		ContractAddress:  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		FunctionSelector: "balanceOf(address)",
		Parameter:        param,
		Visible:          true,
	}

	resp, err := quickNode.CallTriggerSmartContract(req)
	if err != nil {
		return nil, err
	}

	if len(resp.ConstantResult) == 0 {
		return big.NewFloat(0), nil
	}

	return parseTrc20AmountToFloat(resp.ConstantResult[0], 6) // USDT 小數 6 位
}

func getIndiaOrder(orderId string) (*IndiaOrder, error) {
	url := fmt.Sprintf("https://india-api.jj-otc.com/api/orders/?merchant_order_id=%s", orderId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", viper.GetString("andy.i.order_api_key")))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResponse IndiaOrderResponse
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	if len(apiResponse.Results) == 0 {
		return nil, errors.New("no results found for the given order ID")
	}

	return &apiResponse.Results[0], nil
}
