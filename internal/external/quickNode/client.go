package quicknode

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
