package quickNode

import (
	"fmt"
	"regexp"
	"strings"
)

func ParseExpressionToAddresses(expression string) []string {
	// 簡單正則找出 in (...) 的內容
	re := regexp.MustCompile(`tx_logs_topic2\s+in\s+\(([^)]+)\)`)
	matches := re.FindStringSubmatch(expression)
	if len(matches) < 2 {
		return nil
	}

	rawList := matches[1]
	rawItems := strings.Split(rawList, ",")

	addrs := make([]string, 0, len(rawItems))
	for _, s := range rawItems {
		addr := strings.Trim(s, " '") // 去掉空格與單引號
		if strings.HasPrefix(addr, "0x") && len(addr) == 66 {
			addrs = append(addrs, addr)
		}
	}

	return addrs
}

func ParseAddressesToExpression(addresses []string) string {
	if len(addresses) == 0 {
		return ""
	}

	// 將地址包裹在單引號中
	quoted := make([]string, len(addresses))
	for i, addr := range addresses {
		quoted[i] = fmt.Sprintf("'%s'", addr)
	}

	// 組合成表達式
	return fmt.Sprintf(
		"tx_logs_topic0 == '0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef' && (tx_logs_topic2 in (%s))",
		strings.Join(quoted, ", "),
	)
}
