package spf

import "testing"

func TestNewDNSWithResolver(t *testing.T) {
	proto := "udp"
	type args struct {
		nameserver string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test with google public nameservers",
			args: args{
				nameserver: "8.8.8.8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewDNSWithResolver(tt.args.nameserver, proto)
		})
	}
}
