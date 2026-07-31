package main

import "testing"

func TestIsTrustedForwardProxy(t *testing.T) {
	tests := []struct {
		name           string
		trustedProxies string
		remoteAddr     string
		want           bool
	}{
		{
			name:           "empty trusted proxies rejects remote",
			trustedProxies: "",
			remoteAddr:     "10.0.0.5:1234",
			want:           false,
		},
		{
			name:           "exact IP match accepts remote",
			trustedProxies: "10.0.0.5",
			remoteAddr:     "10.0.0.5:1234",
			want:           true,
		},
		{
			name:           "CIDR match accepts remote",
			trustedProxies: "10.0.0.0/24",
			remoteAddr:     "10.0.0.5:1234",
			want:           true,
		},
		{
			name:           "CIDR miss rejects remote",
			trustedProxies: "10.0.1.0/24",
			remoteAddr:     "10.0.0.5:1234",
			want:           false,
		},
		{
			name:           "IPv6 match accepts remote",
			trustedProxies: "::1/128",
			remoteAddr:     "[::1]:1234",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FORWARD_PROXY_AUTHENTICATION_TRUSTED_PROXIES", tt.trustedProxies)

			if got := isTrustedForwardProxy(tt.remoteAddr); got != tt.want {
				t.Fatalf("isTrustedForwardProxy(%q) = %v, want %v", tt.remoteAddr, got, tt.want)
			}
		})
	}
}
