package classify

import (
	"net/netip"
	"strings"
	"sync/atomic"

	"github.com/oschwald/maxminddb-golang/v2"
)

// Network kinds.
const (
	NetHosting     = "hosting"     // a cloud/hosting provider's ASN
	NetResidential = "residential" // any other ASN: home, mobile and business ISPs
	NetUnknown     = "unknown"     // no ASN database, or the address isn't in it (private, reserved)
)

// hostingOrgKeywords mirrors _HOSTING_ORG_KEYWORDS in
// apps/core/ip_intel_classify.py (TestHostingKeywordsMatchPython keeps
// them the same), so the edge and IPIntelligence.is_hosting agree.
var hostingOrgKeywords = []string{
	"AMAZON", "AWS", "GOOGLE", "MICROSOFT", "AZURE",
	"DIGITALOCEAN", "DIGITAL OCEAN", "OVH", "HETZNER", "LINODE", "AKAMAI",
	"VULTR", "CHOOPA", "ALIBABA", "ALIYUN", "TENCENT", "ORACLE",
	"CONTABO", "LEASEWEB", "SCALEWAY", "ONLINE S.A.S", "M247",
	"COLOCROSSING", "ZENLAYER", "HOSTINGER", "IONOS", "GODADDY",
	"RACKSPACE", "PSYCHZ", "DEDIPATH", "FRANTECH", "DATACAMP",
	"G-CORE", "GCORE", "CLOUDFLARE", "FASTLY", "CLOUDVPS",
	"HOSTWINDS", "HOSTKEY", "NFORCE", "SERVERASTRA", "HIVELOCITY",
	"QUADRANET", "UNIFIEDLAYER", "BLUEHOST", "DREAMHOST", "OVHCLOUD",
	"KAMATERA", "UPCLOUD", "CLOUDSIGMA", "INTERSERVER", "RELIABLESITE",
}

// IsHosting mirrors classify_hosting(): a best-effort "probably not a home
// connection" from the ASN organization name.
func IsHosting(org string) bool {
	if org == "" {
		return false
	}
	u := strings.ToUpper(org)
	for _, k := range hostingOrgKeywords {
		if strings.Contains(u, k) {
			return true
		}
	}
	return false
}

// NetInfo is what an IP's ASN says about it.
type NetInfo struct {
	ASN     uint32
	Org     string
	Network string
}

// ASNDB looks client IPs up in a GeoLite2-ASN database. The zero value
// (or one whose file failed to open) answers NetUnknown for everything.
// Load swaps in a new file, so a SIGHUP can pick up the monthly update.
type ASNDB struct {
	r atomic.Pointer[maxminddb.Reader]
}

// Load opens path and, on success, replaces the current database.
func (d *ASNDB) Load(path string) error {
	r, err := maxminddb.Open(path)
	if err != nil {
		return err
	}
	// The old reader isn't closed: lookups in flight may still use it.
	// It's a read-only memory map (~10 MB), replaced about monthly, and
	// every deploy restarts the edge anyway.
	d.r.Store(r)
	return nil
}

func (d *ASNDB) Lookup(ip string) NetInfo {
	var r *maxminddb.Reader
	if d != nil {
		r = d.r.Load()
	}
	addr, err := netip.ParseAddr(ip)
	if r == nil || err != nil {
		return NetInfo{Network: NetUnknown}
	}
	var rec struct {
		ASN uint32 `maxminddb:"autonomous_system_number"`
		Org string `maxminddb:"autonomous_system_organization"`
	}
	res := r.Lookup(addr.Unmap())
	if !res.Found() || res.Decode(&rec) != nil || rec.ASN == 0 {
		return NetInfo{Network: NetUnknown}
	}
	n := NetInfo{ASN: rec.ASN, Org: rec.Org, Network: NetResidential}
	if IsHosting(rec.Org) {
		n.Network = NetHosting
	}
	return n
}
