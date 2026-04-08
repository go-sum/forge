package server

import (
	"fmt"
	"net"

	"github.com/labstack/echo/v5"
)

// BuildIPExtractor returns the Echo IPExtractor that matches the configured
// deployment topology.
//
// trustProxy controls how the real client IP is resolved:
//   - "" or "direct": use RemoteAddr only. Correct when the Go process is
//     directly exposed to clients with no reverse proxy in front.
//   - "xff": read X-Forwarded-For, trusting only hops whose source address
//     falls within one of the trustedProxies CIDRs. Correct behind nginx,
//     Caddy, AWS ALB, or any proxy that appends to X-Forwarded-For.
//
// trustedProxies is a list of CIDR strings (e.g. "172.16.0.0/12", "::1/128")
// identifying addresses allowed to set X-Forwarded-For. Loopback, link-local,
// and private subnets are NOT trusted by default — list them explicitly when
// needed (e.g. when Caddy runs on the same host and connects via ::1).
//
// An error is returned for an invalid trustProxy value or an unparseable CIDR.
func BuildIPExtractor(trustProxy string, trustedProxies []string) (echo.IPExtractor, error) {
	switch trustProxy {
	case "", "direct":
		return echo.ExtractIPDirect(), nil
	case "xff":
		opts := []echo.TrustOption{
			echo.TrustLoopback(false),
			echo.TrustLinkLocal(false),
			echo.TrustPrivateNet(false),
		}
		for _, raw := range trustedProxies {
			_, network, err := net.ParseCIDR(raw)
			if err != nil {
				return nil, fmt.Errorf("trusted_proxies: invalid CIDR %q: %w", raw, err)
			}
			opts = append(opts, echo.TrustIPRange(network))
		}
		return echo.ExtractIPFromXFFHeader(opts...), nil
	default:
		return nil, fmt.Errorf("unknown trust_proxy mode %q", trustProxy)
	}
}
