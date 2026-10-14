package classify

import "testing"

func BenchmarkClassify(b *testing.B) {
	in := Input{UA: chromeUA, AcceptLanguage: "en-US", JA4: "t13d1516h2_8daaf6152771_02713d6af862", H2Fingerprint: chromeH2, Net: residential}
	for b.Loop() {
		Classify(in)
	}
}

func BenchmarkASNLookup(b *testing.B) {
	var db ASNDB
	if err := db.Load("../../acpwb/var/geoip/GeoLite2-ASN.mmdb"); err != nil {
		b.Skip(err)
	}
	for b.Loop() {
		db.Lookup("177.37.12.4")
	}
}
