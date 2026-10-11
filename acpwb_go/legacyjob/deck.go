package legacyjob

import (
	"fmt"
	"strings"
)

// report is the exact printed output (SYSPRINT) and input deck (SYSIN) for a
// page. Every emitter's FORMAT / printf / PICTURE layout targets these exact
// strings, and the page's "Output" section shows them — computed from the
// article's own published values, so the listing always matches the tables.
type report struct {
	Julian   string // 22.096
	PageHdr  string // full first line, including carriage control and page 1
	HdrLeft  string // "1ACPWB RESEARCH DIVISION     BNCHMK REL 2.4A     RECORD "
	HdrMid   string // "     22.096     PAGE"
	Title1   string // "0ENGAGEMENT METRICS -- "
	ColHdr1  string
	Title2   string // "0MARKET BENCHMARK -- "
	ColHdr2  string
	Trailer  []string
	Checksum int

	Org, Peer, RecID string
	MetNames         []string // 40-col card text
	BenNames         []string

	Deck    []string // SYSIN card images
	Printed []string // SYSPRINT lines (right-trimmed)
}

func julian(y, m, d int) string {
	cum := []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	doy := cum[m-1] + d
	if m > 2 && (y%4 == 0 && (y%100 != 0 || y%400 == 0)) {
		doy++
	}
	return fmt.Sprintf("%02d.%03d", y%100, doy)
}

func buildReport(in *Input, id ID, b *builder) *report {
	r := &report{Julian: julian(in.Year, in.Month, in.Day)}
	r.Org = padRight(cardText(in.Org), 40)
	r.Peer = padRight(cardText(in.PeerGroup), 60)
	r.RecID = padRight(strings.ToUpper(in.RecordID), 8)
	r.HdrLeft = fmt.Sprintf("1ACPWB RESEARCH DIVISION     %s REL %-5s     RECORD ", id.Program, id.Release)
	r.HdrMid = fmt.Sprintf("     %s     PAGE", r.Julian)
	r.PageHdr = r.HdrLeft + r.RecID + r.HdrMid + fmt.Sprintf("%5d", 1)
	r.Title1 = "0ENGAGEMENT METRICS -- "
	r.Title2 = "0MARKET BENCHMARK -- "
	r.ColHdr1 = " " + padRight("METRIC", 40) + padLeft("BASELINE", 9) + padLeft("CURRENT", 9) + padLeft("CHANGE", 9)
	r.ColHdr2 = " " + padRight("METRIC", 40)
	for _, c := range ladderCol {
		r.ColHdr2 += padLeft(c, 9)
	}

	for _, m := range in.Metrics {
		r.MetNames = append(r.MetNames, padRight(cardText(m.Name), 40))
	}
	for _, bn := range in.Bench {
		r.BenNames = append(r.BenNames, padRight(cardText(bn.Name), 40))
	}

	// Checksum, exactly as the generated program computes it.
	ck := b.sealH
	for _, m := range in.Metrics {
		ck = (ck*b.ckMul + m.Baseline + m.Current) % b.ckMod
	}
	for _, bn := range in.Bench {
		ck = (ck*b.ckMul + bn.Median) % b.ckMod
	}
	ck = (ck + b.sealL) % b.ckMod
	r.Checksum = ck

	r.Trailer = []string{
		fmt.Sprintf("0RECORDS PROCESSED%6d METRICS%6d BENCHMARKS   SEAL %6d", len(in.Metrics), len(in.Bench), ck),
	}

	// SYSIN deck.
	r.Deck = append(r.Deck, r.Org+r.RecID)
	r.Deck = append(r.Deck, strings.TrimRight(r.Peer, " "))
	r.Deck = append(r.Deck, fmt.Sprintf("%3d%3d", len(in.Metrics), len(in.Bench)))
	for i, m := range in.Metrics {
		r.Deck = append(r.Deck, r.MetNames[i]+fmt.Sprintf("%6d%6d", m.Baseline, m.Current))
	}
	for i, bn := range in.Bench {
		r.Deck = append(r.Deck, r.BenNames[i]+fmt.Sprintf("%7d", bn.Median))
	}

	// SYSPRINT.
	pr := []string{r.PageHdr, r.Title1 + r.Org, r.ColHdr1}
	for i, m := range in.Metrics {
		pr = append(pr, " "+r.MetNames[i]+fmt.Sprintf("%9d%9d%+9d", m.Baseline, m.Current, m.Current-m.Baseline))
	}
	pr = append(pr, r.Title2+r.Peer, r.ColHdr2)
	for i, bn := range in.Bench {
		line := " " + r.BenNames[i]
		for _, v := range bn.P {
			line += fmt.Sprintf("%9d", v)
		}
		pr = append(pr, line)
	}
	pr = append(pr, r.Trailer...)
	for _, l := range pr {
		r.Printed = append(r.Printed, strings.TrimRight(l, " "))
	}
	return r
}
