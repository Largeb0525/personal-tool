package quickNode

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetQuickAlertInfo() (QuickAlert, error) {
	var result QuickAlert

	url := quickAlertsURL + QuickAlertID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-key", ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("bad response: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

func PatchQuickAlert(addresses []string) error {
	rawExpression := ParseAddressesToExpression(addresses)
	expression := base64.StdEncoding.EncodeToString([]byte(rawExpression))

	payload := PatchQuickAlertRequest{
		Expression: expression,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	url := fmt.Sprintf(quickAlertsURL + QuickAlertID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PATCH failed: %s", resp.Status)
	}

	return nil
}

func CreateFreezeTx(req FreezeRequest) (tx Transaction, err error) {
	body, err := json.Marshal(req)
	if err != nil {
		return tx, fmt.Errorf("json marshal failed: %w", err)
	}
	url := fmt.Sprintf(freezeBalanceURL, AppID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return tx, fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return tx, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&tx)
	if err != nil {
		return tx, fmt.Errorf("json decode failed: %w", err)
	}
	return tx, nil
}

func CreateDelegateResourceTx(req DelegateResourceRequest) (tx Transaction, err error) {
	body, err := json.Marshal(req)
	if err != nil {
		return tx, fmt.Errorf("json marshal failed: %w", err)
	}
	url := fmt.Sprintf(delegateResourceURL, AppID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return tx, fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return tx, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&tx)
	if err != nil {
		return tx, fmt.Errorf("json decode failed: %w", err)
	}
	return tx, nil
}

func CreateUndelegateResourceTx(req UndelegateResourceRequest) (tx Transaction, err error) {
	body, err := json.Marshal(req)
	if err != nil {
		return tx, fmt.Errorf("json marshal failed: %w", err)
	}

	url := fmt.Sprintf(undelegateResourceURL, AppID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return tx, fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return tx, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&tx)
	if err != nil {
		return tx, fmt.Errorf("json decode failed: %w", err)
	}
	return tx, nil
}

func BroadcastTransaction(req BroadcastRequest) (result BroadcastResponse, err error) {
	body, _ := json.Marshal(req)
	url := fmt.Sprintf(broadcastURL, AppID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return result, fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return result, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}
