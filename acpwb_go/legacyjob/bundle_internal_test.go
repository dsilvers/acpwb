package legacyjob

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func sampleFiles() []BundleFile {
	big := strings.Repeat("C     THE LADDER WAS RATIFIED IN 1981. DO NOT ROUND. ", 4000)
	return []BundleFile{{Name: "BNCHMK.FOR", Data: []byte(big)}, {Name: "DECK.DAT", Data: []byte("HEDGECROFT INGRAM ADVISORY 1AA96252\n")}, {Name: "EMPTY.TXT", Data: []byte("x")}}
}

func TestZipAndTarRoundTrip(t *testing.T) {
	files := sampleFiles()
	mt := time.Date(1999, 4, 6, 2, 22, 0, 0, time.UTC)
	zr, err := zip.NewReader(bytes.NewReader(writeZip(files, mt)), int64(len(writeZip(files, mt))))
	if err != nil || len(zr.File) != len(files) {
		t.Fatalf("zip: %v %d", err, len(zr.File))
	}
	rc, _ := zr.File[0].Open()
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, files[0].Data) {
		t.Fatal("zip content mismatch")
	}
	tr := tar.NewReader(bytes.NewReader(writeTar(files, mt, "bnchmk-2.4")))
	n := 0
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Typeflag == tar.TypeReg {
			n++
		}
	}
	if n != len(files) {
		t.Fatalf("tar has %d files", n)
	}
}

func TestCompressLZWDecodes(t *testing.T) {
	gz, err := exec.LookPath("gzip")
	if err != nil {
		t.Skip("gzip not available")
	}
	for _, in := range [][]byte{[]byte("a"), []byte("TOBEORNOTTOBEORTOBEORNOT"), []byte(strings.Repeat("abcabcabd", 20000)), writeTar(sampleFiles(), time.Now(), "x")} {
		cmd := exec.Command(gz, "-dc")
		cmd.Stdin = bytes.NewReader(compressLZW(in))
		out, err := cmd.Output()
		if err != nil || !bytes.Equal(out, in) {
			t.Fatalf("compress output not decodable (%d bytes in, %d out): %v", len(in), len(out), err)
		}
	}
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("x"))
	_ = zw.Close()
}

func TestARCAndLZHStructure(t *testing.T) {
	files := sampleFiles()
	mt := time.Date(1989, 2, 3, 4, 5, 6, 0, time.UTC)
	arc := writeARC(files, mt)
	p := 0
	for i := 0; ; i++ {
		if arc[p] != 0x1A {
			t.Fatalf("ARC marker missing at %d", p)
		}
		if arc[p+1] == 0 {
			if i != len(files) {
				t.Fatalf("ARC has %d members", i)
			}
			break
		}
		size := int(binary.LittleEndian.Uint32(arc[p+15:]))
		crc := binary.LittleEndian.Uint16(arc[p+23:])
		data := arc[p+29 : p+29+size]
		if crc != crc16ARC(data) || !bytes.Equal(data, files[i].Data) {
			t.Fatalf("ARC member %d bad", i)
		}
		p += 29 + size
	}
	lzh := writeLZH(files, mt)
	p = 0
	for i := 0; ; i++ {
		hl := int(lzh[p])
		if hl == 0 {
			if i != len(files) {
				t.Fatalf("LZH has %d members", i)
			}
			break
		}
		h := lzh[p+2 : p+2+hl]
		var sum byte
		for _, c := range h {
			sum += c
		}
		if sum != lzh[p+1] || string(h[:5]) != "-lh0-" {
			t.Fatalf("LZH header %d bad", i)
		}
		size := int(binary.LittleEndian.Uint32(h[5:]))
		p += 2 + hl + size
	}
	if crc16ARC([]byte("123456789")) != 0xBB3D {
		t.Fatal("CRC-16/ARC check value wrong")
	}
}
