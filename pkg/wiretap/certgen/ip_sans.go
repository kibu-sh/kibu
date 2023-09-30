package certgen

import "net"

var loopbackIPs = []net.IP{
	net.IPv4(127, 0, 0, 1),
	net.IPv6loopback,
}

func parseHostIntoIPSANS(hostname string) []net.IP {
	var ipSANS []net.IP
	addr, _, _ := net.SplitHostPort(hostname)
	if addr == "" {
		addr = hostname
	}
	if ip := net.ParseIP(addr); ip != nil {
		ipSANS = append(ipSANS, ip)
	}
	return ipSANS
}
