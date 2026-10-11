package legacyjob

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"container/list"
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Downloadable bundles of a source-code page: every file the page shows
// (source, job control, input deck, output, job log, compiler output,
// license), packed in two or three archive formats that existed in the
// article's year. Archives are built on request from the cached listing.

type BundleFile struct {
	Name string
	Data []byte
}

type ArchiveLink struct {
	Name, Format, Note string
}

type archFmt struct {
	ext, label, note string
	from, to         int
	dos              bool // 8.3 upper-case member names
}

var archiveFormats = []archFmt{
	{".ARC", "ARC", "SEA ARC 5.x, stored", 1985, 1991, true},
	{".LZH", "LHarc", "LHarc 1.13, -lh0-", 1988, 1996, true},
	{".ZIP", "PKZIP", "PKZIP 2.04g", 1990, 1999, true},
	{".tar.Z", "tar + compress", "ustar, compress -b16", 1985, 1997, false},
	{".tar", "tar", "ustar, uncompressed", 1985, 2025, false},
	{".tar.gz", "tar + gzip", "ustar, gzip -9", 1994, 2025, false},
	{".tgz", "tgz", "ustar, gzip", 1996, 2025, false},
	{".zip", "zip", "deflate", 2000, 2025, false},
}

func (l *Listing) bundleBase(f archFmt) string {
	rel := strings.NewReplacer(".", "", " ", "").Replace(l.ID.Release)
	if f.dos {
		return trunc(l.ID.Program, 6) + trunc(rel, 2)
	}
	return strings.ToLower(l.ID.Program) + "-" + strings.ToLower(l.ID.Release)
}

// ArchiveLinks returns the two or three downloads offered on the page.
func (l *Listing) ArchiveLinks() []ArchiveLink {
	x := newRnd(fmt.Sprintf("bundle_%s_%s", l.In.Path, l.ID.JobNumber))
	var fit []archFmt
	for _, f := range archiveFormats {
		if l.In.Year >= f.from && l.In.Year <= f.to {
			fit = append(fit, f)
		}
	}
	n := x.between(2, 3)
	if n > len(fit) {
		n = len(fit)
	}
	var out []ArchiveLink
	for _, f := range sample(x, fit, n) {
		out = append(out, ArchiveLink{Name: l.bundleBase(f) + f.ext, Format: f.label, Note: f.note})
	}
	return out
}

func (l *Listing) formatFor(name string) (archFmt, bool) {
	for _, f := range archiveFormats {
		if name == l.bundleBase(f)+f.ext {
			return f, true
		}
	}
	return archFmt{}, false
}

// Files lists the bundle members; dos selects 8.3 upper-case names.
func (l *Listing) Files(dos bool) []BundleFile {
	prog := strings.ToLower(l.ID.Program)
	ext := strings.ToLower(metaOf(l.ID.LangKey).ext)
	if ext == "" {
		ext = ".src"
	}
	name := func(unix, dosName string) string {
		if dos {
			return strings.ToUpper(dosName)
		}
		return unix
	}
	sec := func(id string) string {
		for _, s := range l.Sections {
			if s.ID == id {
				return stripMarkers(s.Text)
			}
		}
		return ""
	}
	src := l.sourceText
	if src == "" {
		src = sec("source")
	}
	var out []BundleFile
	add := func(n, text string) {
		if strings.TrimSpace(text) == "" {
			return
		}
		if dos {
			text = strings.ReplaceAll(text, "\n", "\r\n") + "\x1a"
		}
		out = append(out, BundleFile{Name: n, Data: []byte(text)})
	}
	add(name(prog+ext, trunc(prog, 8)+trunc(ext, 4)), src)
	add(name("jobctl.txt", "JOBCTL.TXT"), sec("jobcontrol"))
	if l.sourceText != "" {
		add(name(prog+".lst", trunc(prog, 8)+".LST"), sec("source"))
	}
	add(name("deck.dat", "DECK.DAT"), sec("sysin"))
	add(name("report.lis", "REPORT.LIS"), sec("sysprint"))
	add(name("job.log", "JOB.LOG"), sec("joblog"))
	add(name("diagnostics.txt", "DIAG.TXT"), sec("diagnostics"))
	add(name("xref.txt", "XREF.TXT"), sec("xref"))
	add(name("storage.map", "STORAGE.MAP"), sec("storagemap"))
	add(name("LICENSE", "LICENSE.TXT"), sec("license"))
	readme := fmt.Sprintf("%s rel. %s (%s)\n%s\nRecord %s  Job %s  Seal %s\n\nTabulated the engagement metrics and market benchmark tables for:\n%s\n",
		l.ID.Program, l.ID.Release, l.ID.Language, l.In.Title, strings.ToUpper(l.In.RecordID), l.ID.JobNumber, strings.ToUpper(l.In.Token), l.In.ArticleURL)
	add(name("README", "README.TXT"), readme)
	return out
}

func (l *Listing) mtime() time.Time {
	return time.Date(l.In.Year, time.Month(l.In.Month), l.In.Day, 2, 22, 0, 0, time.UTC)
}

// ── archive cache ──────────────────────────────────────────────────────────

type bundleEntry struct {
	key  string
	ct   string
	data []byte
}

var (
	bundleMu  sync.Mutex
	bundleLRU = list.New()
	bundleMap = map[string]*list.Element{}
)

const bundleMax = 64

// BuildArchive returns the content type and bytes of the named download, or
// ok=false when the name is not one of this page's archives.
func (l *Listing) BuildArchive(name string) (contentType string, data []byte, ok bool) {
	f, ok := l.formatFor(name)
	if !ok {
		return "", nil, false
	}
	key := l.In.Path + "|" + name
	bundleMu.Lock()
	if el, hit := bundleMap[key]; hit {
		bundleLRU.MoveToFront(el)
		e := el.Value.(*bundleEntry)
		bundleMu.Unlock()
		return e.ct, e.data, true
	}
	bundleMu.Unlock()

	files := l.Files(f.dos)
	mt := l.mtime()
	var ct string
	switch f.ext {
	case ".ARC":
		ct, data = "application/x-arc", writeARC(files, mt)
	case ".LZH":
		ct, data = "application/x-lzh-compressed", writeLZH(files, mt)
	case ".ZIP", ".zip":
		ct, data = "application/zip", writeZip(files, mt)
	case ".tar":
		ct, data = "application/x-tar", writeTar(files, mt, l.bundleBase(f))
	case ".tar.Z":
		ct, data = "application/x-compress", compressLZW(writeTar(files, mt, l.bundleBase(f)))
	default:
		var buf bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		zw.Name = l.bundleBase(f) + ".tar"
		zw.ModTime = mt
		_, _ = zw.Write(writeTar(files, mt, l.bundleBase(f)))
		_ = zw.Close()
		ct, data = "application/gzip", buf.Bytes()
	}

	bundleMu.Lock()
	bundleMap[key] = bundleLRU.PushFront(&bundleEntry{key: key, ct: ct, data: data})
	for bundleLRU.Len() > bundleMax {
		old := bundleLRU.Back()
		bundleLRU.Remove(old)
		delete(bundleMap, old.Value.(*bundleEntry).key)
	}
	bundleMu.Unlock()
	return ct, data, true
}

// ── writers ────────────────────────────────────────────────────────────────

func writeZip(files []BundleFile, mt time.Time) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range files {
		w, _ := zw.CreateHeader(&zip.FileHeader{Name: f.Name, Method: zip.Deflate, Modified: mt})
		_, _ = w.Write(f.Data)
	}
	_ = zw.Close()
	return buf.Bytes()
}

func writeTar(files []BundleFile, mt time.Time, dir string) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	_ = tw.WriteHeader(&tar.Header{Name: dir + "/", Typeflag: tar.TypeDir, Mode: 0755, ModTime: mt, Uname: "rsch", Gname: "staff", Format: tar.FormatUSTAR})
	for _, f := range files {
		_ = tw.WriteHeader(&tar.Header{Name: dir + "/" + f.Name, Mode: 0644, Size: int64(len(f.Data)), ModTime: mt, Uname: "rsch", Gname: "staff", Format: tar.FormatUSTAR})
		_, _ = tw.Write(f.Data)
	}
	_ = tw.Close()
	return buf.Bytes()
}

// crc16ARC is the CRC-16 (poly 0xA001, reflected) used by ARC and LHarc.
func crc16ARC(b []byte) uint16 {
	var crc uint16
	for _, c := range b {
		crc ^= uint16(c)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = crc>>1 ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func dosDateTime(t time.Time) (date, tm uint16) {
	date = uint16((t.Year()-1980)<<9 | int(t.Month())<<5 | t.Day())
	tm = uint16(t.Hour()<<11 | t.Minute()<<5 | t.Second()/2)
	return
}

func dosName(n string) []byte {
	b := make([]byte, 13)
	copy(b, trunc(strings.ToUpper(n), 12))
	return b
}

// writeARC: SEA ARC archive, every member stored (method 2).
func writeARC(files []BundleFile, mt time.Time) []byte {
	var buf bytes.Buffer
	d, t := dosDateTime(mt)
	for _, f := range files {
		buf.WriteByte(0x1A)
		buf.WriteByte(2)
		buf.Write(dosName(f.Name))
		_ = binary.Write(&buf, binary.LittleEndian, uint32(len(f.Data)))
		_ = binary.Write(&buf, binary.LittleEndian, d)
		_ = binary.Write(&buf, binary.LittleEndian, t)
		_ = binary.Write(&buf, binary.LittleEndian, crc16ARC(f.Data))
		_ = binary.Write(&buf, binary.LittleEndian, uint32(len(f.Data)))
		buf.Write(f.Data)
	}
	buf.Write([]byte{0x1A, 0x00})
	return buf.Bytes()
}

// writeLZH: LHarc level-0 headers, method -lh0- (stored).
func writeLZH(files []BundleFile, mt time.Time) []byte {
	var buf bytes.Buffer
	d, t := dosDateTime(mt)
	for _, f := range files {
		name := []byte(strings.ToUpper(trunc(f.Name, 12)))
		var h bytes.Buffer
		h.WriteString("-lh0-")
		_ = binary.Write(&h, binary.LittleEndian, uint32(len(f.Data)))
		_ = binary.Write(&h, binary.LittleEndian, uint32(len(f.Data)))
		_ = binary.Write(&h, binary.LittleEndian, t)
		_ = binary.Write(&h, binary.LittleEndian, d)
		h.WriteByte(0x20) // attribute: archive
		h.WriteByte(0)    // header level 0
		h.WriteByte(byte(len(name)))
		h.Write(name)
		_ = binary.Write(&h, binary.LittleEndian, crc16ARC(f.Data))
		hb := h.Bytes()
		var sum byte
		for _, c := range hb {
			sum += c
		}
		buf.WriteByte(byte(len(hb)))
		buf.WriteByte(sum)
		buf.Write(hb)
		buf.Write(f.Data)
	}
	buf.WriteByte(0)
	return buf.Bytes()
}

// compressLZW implements Unix compress(1) output (.Z): magic 1F 9D, block
// mode, 16-bit maximum codes, codes written LSB-first in groups of n_bits
// bytes that are padded out whenever the code width grows — exactly as
// compress 4.x does, so uncompress / gzip -d can read it.
func compressLZW(in []byte) []byte {
	const maxBits = 16
	const maxMaxCode = 1 << maxBits
	out := []byte{0x1F, 0x9D, 0x80 | maxBits}
	if len(in) == 0 {
		return out
	}
	nBits := 9
	maxCode := 1<<nBits - 1
	freeEnt := 257
	group := make([]byte, maxBits)
	offset := 0 // bit offset within group
	output := func(code int) {
		if code >= 0 {
			for i := 0; i < nBits; i++ {
				if code&(1<<i) != 0 {
					group[(offset+i)/8] |= 1 << ((offset + i) % 8)
				}
			}
			offset += nBits
			if offset == nBits*8 {
				out = append(out, group[:nBits]...)
				for i := range group {
					group[i] = 0
				}
				offset = 0
			}
			if freeEnt > maxCode {
				if offset > 0 {
					out = append(out, group[:nBits]...)
					for i := range group {
						group[i] = 0
					}
				}
				offset = 0
				nBits++
				if nBits == maxBits {
					maxCode = maxMaxCode
				} else {
					maxCode = 1<<nBits - 1
				}
			}
			return
		}
		if offset > 0 {
			out = append(out, group[:(offset+7)/8]...)
		}
	}
	dict := make(map[int]int, 1<<14)
	ent := int(in[0])
	for _, c := range in[1:] {
		key := int(c)<<16 | ent
		if code, ok := dict[key]; ok {
			ent = code
			continue
		}
		output(ent)
		ent = int(c)
		if freeEnt < maxMaxCode {
			dict[key] = freeEnt
			freeEnt++
		}
	}
	output(ent)
	output(-1)
	return out
}
