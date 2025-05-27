package andy

import "testing"

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

func TestTronToHexPadded32(t *testing.T) {
	type args struct {
		tronAddr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				tronAddr: "TGQw4PERdLnWBbqGntJyxhHhYzA61gzfhn",
			},
			want:    "0x00000000000000000000000046ae42ed9c71a9f2ff80f44b7ba6131bfe49e679",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TronToHexPadded32(tt.args.tronAddr)
			if (err != nil) != tt.wantErr {
				t.Errorf("TronToHexPadded32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TronToHexPadded32() = %v, want %v", got, tt.want)
			}
		})
	}
}
