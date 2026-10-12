// Package tlsfp parses a raw TLS ClientHello off the wire and computes the
// client fingerprints acpwb_edge forwards and records: JA4, JA4_r and JA3.
//
// It parses the bytes itself rather than using crypto/tls's ClientHelloInfo
// so that it still has a fingerprint for clients crypto/tls rejects (old or
// malformed hellos, scanners), and so it sees the legacy record version and
// exact extension order JA3 needs.
package tlsfp

import (
	"errors"

	"golang.org/x/crypto/cryptobyte"
)

// MaxHelloBytes caps how much of a connection's opening bytes are buffered
// while waiting for a complete ClientHello. Real hellos are well under
// 2 KB (post-quantum key shares push Chrome's to ~1.8 KB); anything larger
// is abuse or garbage.
const MaxHelloBytes = 16 << 10

var (
	// ErrNeedMore means the bytes so far are a valid prefix of a
	// ClientHello but it isn't complete yet.
	ErrNeedMore = errors.New("tlsfp: incomplete ClientHello")
	// ErrNotTLS means the connection didn't open with a TLS handshake
	// record at all (plain HTTP to :443, port scanners, junk).
	ErrNotTLS = errors.New("tlsfp: not a TLS handshake")
	// ErrMalformed means it looked like TLS but the ClientHello didn't parse.
	ErrMalformed = errors.New("tlsfp: malformed ClientHello")
)

const (
	recordTypeHandshake   = 0x16
	handshakeClientHello  = 0x01
	extServerName         = 0x0000
	extSupportedGroups    = 0x000a
	extECPointFormats     = 0x000b
	extSignatureAlgs      = 0x000d
	extALPN               = 0x0010
	extSupportedVersions  = 0x002b
	recordHeaderLen       = 5
	handshakeHeaderLen    = 4
	maxRecordFragmentSize = 1<<14 + 2048
)

// Hello is the parts of a ClientHello the fingerprints are built from, in
// wire order, GREASE values included.
type Hello struct {
	LegacyVersion       uint16
	CipherSuites        []uint16
	Extensions          []uint16
	SupportedVersions   []uint16
	SignatureAlgorithms []uint16
	SupportedGroups     []uint16
	ECPointFormats      []uint8
	ALPN                []string
	ServerName          string
}

// Parse reads a ClientHello from the start of a connection's byte stream,
// reassembling it across TLS records if the client fragmented it. It
// returns ErrNeedMore until the whole handshake message has arrived.
func Parse(raw []byte) (*Hello, error) {
	msg, err := handshakeMessage(raw)
	if err != nil {
		return nil, err
	}
	return parseClientHello(msg)
}

// handshakeMessage concatenates handshake record fragments until it holds
// one complete handshake message, and returns that message.
func handshakeMessage(raw []byte) ([]byte, error) {
	var msg []byte
	for {
		if len(raw) < recordHeaderLen {
			if len(raw) > 0 && raw[0] != recordTypeHandshake || len(raw) > 1 && raw[1] != 0x03 {
				return nil, ErrNotTLS
			}
			return nil, ErrNeedMore
		}
		if raw[0] != recordTypeHandshake || raw[1] != 0x03 {
			return nil, ErrNotTLS
		}
		n := int(raw[3])<<8 | int(raw[4])
		if n == 0 || n > maxRecordFragmentSize {
			return nil, ErrMalformed
		}
		if len(raw) < recordHeaderLen+n {
			return nil, ErrNeedMore
		}
		msg = append(msg, raw[recordHeaderLen:recordHeaderLen+n]...)
		raw = raw[recordHeaderLen+n:]

		if len(msg) >= handshakeHeaderLen {
			if msg[0] != handshakeClientHello {
				return nil, ErrMalformed
			}
			want := handshakeHeaderLen + (int(msg[1])<<16 | int(msg[2])<<8 | int(msg[3]))
			if want > MaxHelloBytes {
				return nil, ErrMalformed
			}
			if len(msg) >= want {
				return msg[handshakeHeaderLen:want], nil
			}
		}
	}
}

func parseClientHello(body []byte) (*Hello, error) {
	h := &Hello{}
	s := cryptobyte.String(body)
	var random, sessionID, ciphers, compression cryptobyte.String
	if !s.ReadUint16(&h.LegacyVersion) ||
		!s.ReadBytes((*[]byte)(&random), 32) ||
		!s.ReadUint8LengthPrefixed(&sessionID) ||
		!s.ReadUint16LengthPrefixed(&ciphers) ||
		!s.ReadUint8LengthPrefixed(&compression) {
		return nil, ErrMalformed
	}
	for !ciphers.Empty() {
		var c uint16
		if !ciphers.ReadUint16(&c) {
			return nil, ErrMalformed
		}
		h.CipherSuites = append(h.CipherSuites, c)
	}
	if s.Empty() {
		return h, nil // no extensions at all: legal for very old clients
	}
	var exts cryptobyte.String
	if !s.ReadUint16LengthPrefixed(&exts) {
		return nil, ErrMalformed
	}
	for !exts.Empty() {
		var typ uint16
		var data cryptobyte.String
		if !exts.ReadUint16(&typ) || !exts.ReadUint16LengthPrefixed(&data) {
			return nil, ErrMalformed
		}
		h.Extensions = append(h.Extensions, typ)
		// A malformed extension body still counts for the fingerprint; it
		// just doesn't contribute its contents.
		switch typ {
		case extServerName:
			h.ServerName = parseServerName(data)
		case extSupportedGroups:
			h.SupportedGroups = readUint16List(data)
		case extECPointFormats:
			var list cryptobyte.String
			if data.ReadUint8LengthPrefixed(&list) {
				h.ECPointFormats = append([]uint8(nil), list...)
			}
		case extSignatureAlgs:
			h.SignatureAlgorithms = readUint16List(data)
		case extALPN:
			h.ALPN = parseALPN(data)
		case extSupportedVersions:
			var list cryptobyte.String
			if data.ReadUint8LengthPrefixed(&list) {
				for !list.Empty() {
					var v uint16
					if !list.ReadUint16(&v) {
						break
					}
					h.SupportedVersions = append(h.SupportedVersions, v)
				}
			}
		}
	}
	return h, nil
}

func readUint16List(data cryptobyte.String) []uint16 {
	var list cryptobyte.String
	if !data.ReadUint16LengthPrefixed(&list) {
		return nil
	}
	var out []uint16
	for !list.Empty() {
		var v uint16
		if !list.ReadUint16(&v) {
			break
		}
		out = append(out, v)
	}
	return out
}

func parseServerName(data cryptobyte.String) string {
	var list cryptobyte.String
	if !data.ReadUint16LengthPrefixed(&list) {
		return ""
	}
	for !list.Empty() {
		var nameType uint8
		var name cryptobyte.String
		if !list.ReadUint8(&nameType) || !list.ReadUint16LengthPrefixed(&name) {
			return ""
		}
		if nameType == 0 {
			return string(name)
		}
	}
	return ""
}

func parseALPN(data cryptobyte.String) []string {
	var list cryptobyte.String
	if !data.ReadUint16LengthPrefixed(&list) {
		return nil
	}
	var out []string
	for !list.Empty() {
		var proto cryptobyte.String
		if !list.ReadUint8LengthPrefixed(&proto) {
			break
		}
		out = append(out, string(proto))
	}
	return out
}
