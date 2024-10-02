package httpx

import (
	"net"
	"net/http"
)

type CloudflareMeta struct {
	Country      string
	TrueClientIP net.IP
}

func CloudflareMetaFromHeaders(header http.Header) (meta CloudflareMeta) {
	meta.TrueClientIP = net.ParseIP(header.Get("True-Client-IP"))
	meta.Country = header.Get("cf-ipcountry")
	return
}
