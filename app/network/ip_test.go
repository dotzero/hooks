package network

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expected   string
	}{
		{
			name:       "IPv4",
			remoteAddr: "1.1.1.1:80",
			headers:    map[string]string{},
			expected:   "1.1.1.1",
		},
		{
			name:       "IPv6",
			remoteAddr: "[::1]:80",
			headers:    map[string]string{},
			expected:   "::1",
		},
		{
			name:       "X-Forwarded one",
			remoteAddr: "1.1.1.1:80",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
			},
			expected: "203.0.113.195",
		},
		{
			name:       "X-Forwarded many",
			remoteAddr: "1.1.1.1:80",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178",
			},
			expected: "203.0.113.195",
		},
		{
			name:       "X-Real-Ip",
			remoteAddr: "1.1.1.1:80",
			headers: map[string]string{
				"X-Real-Ip": "203.0.113.195",
			},
			expected: "203.0.113.195",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
			assert.NoError(t, err)

			req.RemoteAddr = c.remoteAddr

			for k, v := range c.headers {
				req.Header.Add(k, v)
			}

			ip := ClientIP(req)

			assert.Equal(t, c.expected, ip.String())
		})
	}
}
