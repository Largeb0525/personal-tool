package quicknode

import (
	"reflect"
	"testing"
)

func TestParseExpressionToAddresses(t *testing.T) {
	type args struct {
		expression string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				expression: "tx_logs_topic0 == '0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef' && (tx_logs_topic2 in ('0x00000000000000000000000046ae42ed9c71a9f2ff80f44b7ba6131bfe49e679', '0x0000000000000000000000000bc33af5fd0f228a7bbf46aa83e3429ec61651b3'))",
			},
			want: []string{"0x00000000000000000000000046ae42ed9c71a9f2ff80f44b7ba6131bfe49e679", "0x0000000000000000000000000bc33af5fd0f228a7bbf46aa83e3429ec61651b3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseExpressionToAddresses(tt.args.expression); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTopic2Addresses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAddressesToExpression(t *testing.T) {
	type args struct {
		addresses []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				addresses: []string{"0x00000000000000000000000046ae42ed9c71a9f2ff80f44b7ba6131bfe49e679", "0x0000000000000000000000000bc33af5fd0f228a7bbf46aa83e3429ec61651b3"},
			},
			want: "tx_logs_topic0 == '0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef' && (tx_logs_topic2 in ('0x00000000000000000000000046ae42ed9c71a9f2ff80f44b7ba6131bfe49e679', '0x0000000000000000000000000bc33af5fd0f228a7bbf46aa83e3429ec61651b3'))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseAddressesToExpression(tt.args.addresses); got != tt.want {
				t.Errorf("ParseAddressesToExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}
