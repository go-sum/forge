package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestBuildIPExtractor(t *testing.T) {
	tests := []struct {
		name           string
		trustProxy     string
		trustedProxies []string
		remoteAddr     string
		xffHeader      string
		wantIP         string
		wantErr        string
	}{
		{
			name:       "direct mode uses RemoteAddr regardless of XFF",
			trustProxy: "direct",
			remoteAddr: "1.2.3.4:9999",
			xffHeader:  "9.9.9.9",
			wantIP:     "1.2.3.4",
		},
		{
			name:       "empty trust_proxy defaults to direct",
			trustProxy: "",
			remoteAddr: "1.2.3.4:9999",
			xffHeader:  "9.9.9.9",
			wantIP:     "1.2.3.4",
		},
		{
			name:           "xff with trusted IPv4 CIDR yields XFF client IP",
			trustProxy:     "xff",
			trustedProxies: []string{"10.0.0.0/8"},
			remoteAddr:     "10.0.0.1:9999",
			xffHeader:      "9.9.9.9",
			wantIP:         "9.9.9.9",
		},
		{
			name:           "xff with trusted IPv6 loopback yields XFF client IP",
			trustProxy:     "xff",
			trustedProxies: []string{"::1/128"},
			remoteAddr:     "[::1]:9999",
			xffHeader:      "9.9.9.9",
			wantIP:         "9.9.9.9",
		},
		{
			name:           "xff with untrusted RemoteAddr yields RemoteAddr",
			trustProxy:     "xff",
			trustedProxies: []string{"10.0.0.0/8"},
			remoteAddr:     "1.2.3.4:9999",
			xffHeader:      "9.9.9.9",
			wantIP:         "1.2.3.4",
		},
		{
			name:           "xff private net not trusted implicitly",
			trustProxy:     "xff",
			trustedProxies: nil,
			remoteAddr:     "192.168.1.1:9999",
			xffHeader:      "9.9.9.9",
			wantIP:         "192.168.1.1",
		},
		{
			name:           "xff with invalid CIDR returns error",
			trustProxy:     "xff",
			trustedProxies: []string{"not-a-cidr"},
			wantErr:        "invalid CIDR",
		},
		{
			name:       "unknown trust_proxy mode returns error",
			trustProxy: "bogus",
			wantErr:    "unknown trust_proxy mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor, err := BuildIPExtractor(tt.trustProxy, tt.trustedProxies)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("BuildIPExtractor() error = nil, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("BuildIPExtractor() error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildIPExtractor() unexpected error = %v", err)
			}

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xffHeader != "" {
				req.Header.Set("X-Forwarded-For", tt.xffHeader)
			}

			if got := extractor(req); got != tt.wantIP {
				t.Errorf("extractor(req) = %q, want %q", got, tt.wantIP)
			}
		})
	}
}
