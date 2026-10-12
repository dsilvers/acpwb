// ja4probe makes TLS handshakes with byte-exact browser ClientHellos
// (uTLS profiles), for checking acpwb_edge's fingerprints against an
// independent implementation (Wireshark's).
//
//	ja4probe -addr localhost:8001
//	    One handshake per profile against a running edge (see
//	    scripts/ja4check.sh, which captures them with tshark and diffs
//	    tshark's JA4/JA3 against the EdgeConnection rows).
//
//	ja4probe -write-fixtures tlsfp/testdata/hellos
//	    No network: writes each profile's ClientHello bytes to <name>.bin,
//	    plus hellos.pcap (one synthetic TCP segment per hello, source port
//	    = index.tsv's port) for scripts/ja4golden.sh to run tshark over.
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	utls "github.com/refraction-networking/utls"

	"acpwb_edge/tlsfp"
)

// Pinned profiles (not *_Auto) so fixtures don't move when uTLS updates.
var profiles = []struct {
	name string
	id   utls.ClientHelloID
}{
	{"chrome_133", utls.HelloChrome_133},
	{"chrome_131", utls.HelloChrome_131},
	{"chrome_120_pq", utls.HelloChrome_120_PQ},
	{"chrome_120", utls.HelloChrome_120},
	{"chrome_115_pq", utls.HelloChrome_115_PQ},
	{"chrome_106_shuffle", utls.HelloChrome_106_Shuffle},
	{"chrome_72", utls.HelloChrome_72},
	{"edge_106", utls.HelloEdge_106},
	{"edge_85", utls.HelloEdge_85},
	{"firefox_120", utls.HelloFirefox_120},
	{"firefox_105", utls.HelloFirefox_105},
	{"firefox_65", utls.HelloFirefox_65},
	{"safari_16_0", utls.HelloSafari_16_0},
	{"ios_14", utls.HelloIOS_14},
	{"ios_12_1", utls.HelloIOS_12_1},
	{"android_11_okhttp", utls.HelloAndroid_11_OkHttp},
	{"qq_11_1", utls.HelloQQ_11_1},
	{"360_11_0", utls.Hello360_11_0},
	{"golang", utls.HelloGolang},
}

const basePort = 40000

func main() {
	addr := flag.String("addr", "localhost:8001", "edge address to handshake with")
	sni := flag.String("sni", "acpwb.com", "SNI server name")
	fixtures := flag.String("write-fixtures", "", "write ClientHello fixtures to this directory instead of connecting")
	flag.Parse()

	if *fixtures != "" {
		if err := writeFixtures(*fixtures, *sni); err != nil {
			log.Fatal(err)
		}
		return
	}
	failed := 0
	for _, p := range profiles {
		alpn, err := handshake(*addr, *sni, p.id)
		status := "ok  " + alpn
		if err != nil {
			status, failed = "FAIL "+err.Error(), failed+1
		}
		fmt.Printf("%-20s %s\n", p.name, status)
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func handshake(addr, sni string, id utls.ClientHelloID) (string, error) {
	c, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(10 * time.Second))
	// Verification is off: this only talks to the dev stack's self-signed
	// cert, and only to measure the handshake.
	uc := utls.UClient(c, &utls.Config{ServerName: sni, InsecureSkipVerify: true}, id)
	if err := uc.Handshake(); err != nil {
		return "", err
	}
	return uc.ConnectionState().NegotiatedProtocol, nil
}

// HelloBytes returns the exact bytes a profile sends up to the end of its
// ClientHello, captured over an in-memory pipe.
func helloBytes(sni string, id utls.ClientHelloID) ([]byte, error) {
	client, server := net.Pipe()
	defer server.Close()
	go func() {
		uc := utls.UClient(client, &utls.Config{ServerName: sni, InsecureSkipVerify: true}, id)
		_ = uc.Handshake()
		client.Close()
	}()
	var raw []byte
	buf := make([]byte, 4096)
	for {
		server.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := server.Read(buf)
		raw = append(raw, buf[:n]...)
		if _, perr := tlsfp.Parse(raw); perr == nil {
			return raw, nil
		} else if !errors.Is(perr, tlsfp.ErrNeedMore) {
			return nil, perr
		}
		if err != nil {
			return nil, err
		}
	}
}

func writeFixtures(dir, sni string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var index bytes.Buffer
	var pcap bytes.Buffer
	pcapHeader(&pcap)
	for i, p := range profiles {
		raw, err := helloBytes(sni, p.id)
		if err != nil {
			return fmt.Errorf("%s: %w", p.name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, p.name+".bin"), raw, 0o644); err != nil {
			return err
		}
		port := basePort + i
		fmt.Fprintf(&index, "%d\t%s\n", port, p.name)
		pcapRecord(&pcap, uint16(port), raw)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.tsv"), index.Bytes(), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "hellos.pcap"), pcap.Bytes(), 0o644)
}

// pcap with LINKTYPE_RAW (bare IPv4): just enough for tshark's TLS
// dissector, which picks up port 443 without needing a TCP handshake.
func pcapHeader(b *bytes.Buffer) {
	binary.Write(b, binary.LittleEndian, struct {
		Magic                  uint32
		Major, Minor           uint16
		Zone, SigFigs, SnapLen uint32
		LinkType               uint32
	}{0xa1b2c3d4, 2, 4, 0, 0, 262144, 101})
}

func pcapRecord(b *bytes.Buffer, srcPort uint16, payload []byte) {
	ip := make([]byte, 20)
	ip[0] = 0x45
	binary.BigEndian.PutUint16(ip[2:], uint16(40+len(payload)))
	ip[8], ip[9] = 64, 6 // TTL, TCP
	copy(ip[12:], []byte{10, 0, 0, 1})
	copy(ip[16:], []byte{10, 0, 0, 2})
	tcp := make([]byte, 20)
	binary.BigEndian.PutUint16(tcp[0:], srcPort)
	binary.BigEndian.PutUint16(tcp[2:], 443)
	binary.BigEndian.PutUint32(tcp[4:], 1)
	tcp[12] = 5 << 4
	tcp[13] = 0x18 // PSH|ACK
	binary.BigEndian.PutUint16(tcp[14:], 65535)
	pkt := append(append(ip, tcp...), payload...)
	binary.Write(b, binary.LittleEndian, [4]uint32{uint32(srcPort - basePort + 1), 0, uint32(len(pkt)), uint32(len(pkt))})
	b.Write(pkt)
}
