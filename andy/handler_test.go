package andy

import (
	"reflect"
	"testing"
)

func Test_parseTransactionData(t *testing.T) {
	type args struct {
		receipt TransactionReceipt
	}
	tests := []struct {
		name    string
		args    args
		want    ParsedTransaction
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				receipt: TransactionReceipt{
					TransactionHash: "0x8825515c28dd04b6f74c9d4420e154a50a6be2040a7d9a1926690977069405db",
					Logs: []Log{
						{
							Data: "0x000000000000000000000000000000000000000000000000000000009099d280",
							Topics: []string{
								"0x0000000000000000000000000000000000000000000000000000000000000000",
								"0x0000000000000000000000000ba20112baf064cc1957034a7843e7569e13ddb5",
								"0x000000000000000000000000e4803a5c20ba80cd40becc50bc1acfd0bb965d64",
							},
						},
					},
				},
			},
			want: ParsedTransaction{
				TransactionHash: "8825515c28dd04b6f74c9d4420e154a50a6be2040a7d9a1926690977069405db",
				URL:             "https://tronscan.org/#/transaction/8825515c28dd04b6f74c9d4420e154a50a6be2040a7d9a1926690977069405db",
				USDT:            "2426.000000",
				FromAddress:     "TB2iWRBWNwY9tKHf8CGspaU9TwD4nixxbV",
				ToAddress:       "TWoQgEPuLAF1y3jPENRaPqtQJNWFqxyw24",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTransactionData(tt.args.receipt)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTransactionData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTransactionData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUSDTBalance(t *testing.T) {
	type args struct {
		tokens []TokenInfo
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				tokens: []TokenInfo{
					{
						TokenAbbr: "USDT",
						Balance:   "2426.000000",
					},
				},
			},
			want: "2426.000000",
		},
		{
			name: "test2",
			args: args{
				tokens: []TokenInfo{
					{
						TokenAbbr: "asd",
						Balance:   "0.000000",
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUSDTBalance(tt.args.tokens); got != tt.want {
				t.Errorf("getUSDTBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}
