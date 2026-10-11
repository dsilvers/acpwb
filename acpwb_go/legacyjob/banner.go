package legacyjob

import (
	"fmt"
	"strings"
)

// 5x7 dot-matrix font for the JES/print separator page.
var font5x7 = map[rune][7]string{
	'A': {" ### ", "#   #", "#   #", "#####", "#   #", "#   #", "#   #"},
	'B': {"#### ", "#   #", "#   #", "#### ", "#   #", "#   #", "#### "},
	'C': {" ### ", "#   #", "#    ", "#    ", "#    ", "#   #", " ### "},
	'D': {"#### ", "#   #", "#   #", "#   #", "#   #", "#   #", "#### "},
	'E': {"#####", "#    ", "#    ", "#### ", "#    ", "#    ", "#####"},
	'F': {"#####", "#    ", "#    ", "#### ", "#    ", "#    ", "#    "},
	'G': {" ### ", "#   #", "#    ", "# ###", "#   #", "#   #", " ####"},
	'H': {"#   #", "#   #", "#   #", "#####", "#   #", "#   #", "#   #"},
	'I': {"#####", "  #  ", "  #  ", "  #  ", "  #  ", "  #  ", "#####"},
	'J': {"  ###", "   # ", "   # ", "   # ", "   # ", "#  # ", " ##  "},
	'K': {"#   #", "#  # ", "# #  ", "##   ", "# #  ", "#  # ", "#   #"},
	'L': {"#    ", "#    ", "#    ", "#    ", "#    ", "#    ", "#####"},
	'M': {"#   #", "## ##", "# # #", "# # #", "#   #", "#   #", "#   #"},
	'N': {"#   #", "##  #", "# # #", "#  ##", "#   #", "#   #", "#   #"},
	'O': {" ### ", "#   #", "#   #", "#   #", "#   #", "#   #", " ### "},
	'P': {"#### ", "#   #", "#   #", "#### ", "#    ", "#    ", "#    "},
	'Q': {" ### ", "#   #", "#   #", "#   #", "# # #", "#  # ", " ## #"},
	'R': {"#### ", "#   #", "#   #", "#### ", "# #  ", "#  # ", "#   #"},
	'S': {" ####", "#    ", "#    ", " ### ", "    #", "    #", "#### "},
	'T': {"#####", "  #  ", "  #  ", "  #  ", "  #  ", "  #  ", "  #  "},
	'U': {"#   #", "#   #", "#   #", "#   #", "#   #", "#   #", " ### "},
	'V': {"#   #", "#   #", "#   #", "#   #", "#   #", " # # ", "  #  "},
	'W': {"#   #", "#   #", "#   #", "# # #", "# # #", "## ##", "#   #"},
	'X': {"#   #", "#   #", " # # ", "  #  ", " # # ", "#   #", "#   #"},
	'Y': {"#   #", "#   #", " # # ", "  #  ", "  #  ", "  #  ", "  #  "},
	'Z': {"#####", "    #", "   # ", "  #  ", " #   ", "#    ", "#####"},
	'0': {" ### ", "#   #", "#  ##", "# # #", "##  #", "#   #", " ### "},
	'1': {"  #  ", " ##  ", "  #  ", "  #  ", "  #  ", "  #  ", " ### "},
	'2': {" ### ", "#   #", "    #", "   # ", "  #  ", " #   ", "#####"},
	'3': {"#####", "   # ", "  #  ", "   # ", "    #", "#   #", " ### "},
	'4': {"   # ", "  ## ", " # # ", "#  # ", "#####", "   # ", "   # "},
	'5': {"#####", "#    ", "#### ", "    #", "    #", "#   #", " ### "},
	'6': {"  ## ", " #   ", "#    ", "#### ", "#   #", "#   #", " ### "},
	'7': {"#####", "    #", "   # ", "  #  ", " #   ", " #   ", " #   "},
	'8': {" ### ", "#   #", "#   #", " ### ", "#   #", "#   #", " ### "},
	'9': {" ### ", "#   #", "#   #", " ####", "    #", "   # ", " ##  "},
	'-': {"     ", "     ", "     ", "#####", "     ", "     ", "     "},
	'.': {"     ", "     ", "     ", "     ", "     ", " ##  ", " ##  "},
	'/': {"    #", "    #", "   # ", "  #  ", " #   ", "#    ", "#    "},
	' ': {"     ", "     ", "     ", "     ", "     ", "     ", "     "},
}

// blockLetters renders text in the 5x7 font, each glyph drawn with its own
// character (the classic separator-page look). Unknown runes are blanks.
func blockLetters(text string, ink rune) []string {
	rows := make([]string, 7)
	for _, r := range strings.ToUpper(text) {
		g, ok := font5x7[r]
		if !ok {
			g = font5x7[' ']
		}
		ch := r
		if ink != 0 || r == ' ' || r == '-' || r == '.' || r == '/' {
			ch = ink
			if ch == 0 {
				ch = '*'
			}
		}
		for i := 0; i < 7; i++ {
			rows[i] += strings.ReplaceAll(g[i], "#", string(ch)) + "  "
		}
	}
	return rows
}

func bannerPage(j *job, l *Listing, jobName, userID string, end bool) string {
	x := j.x
	word := "START"
	if end {
		word = " END "
	}
	ink := rune(0)
	if x.chance(0.4) {
		ink = pick(x, []rune{'*', 'X', '#', '@', 'H'})
	}
	edge := fmt.Sprintf("****%s****  %s  %s  ROOM %d  BIN %02d  %s  %s  %s  ****%s****",
		word, l.ID.JobNumber, jobName, x.between(100, 999), x.between(1, 40), userID,
		fmt.Sprintf("%02d %s %04d", j.in.Day, monthAbbr[j.in.Month-1], j.in.Year), strings.ToUpper(j.in.RecordID), word)
	var out []string
	for i := 0; i < 3; i++ {
		out = append(out, edge)
	}
	out = append(out, "")
	for _, w := range []string{jobName, userID, l.ID.JobNumber[3:]} {
		out = append(out, blockLetters(w, ink)...)
		out = append(out, "")
	}
	out = append(out, fmt.Sprintf("     PROGRAMMER: %-30s  DEPT: %-24s  PRINTER: %s", cardText(j.g.roster[len(j.g.roster)-1].Name()), cardText(pick(x, getLibrary().names["depts"])), pick(x, []string{"PRT3", "RMT12", "LPA0", "PRT1"})),
		fmt.Sprintf("     FORMS: %-8s  COPIES: %d  CLASS: %s  DEST: %s", pick(x, []string{"STD", "GRNB", "1PRT", "WIDE"}), x.between(1, 3), string(rune('A'+x.intn(6))), pick(x, []string{"LOCAL", "RMT12", "R14", "BLDGC"})), "")
	for i := 0; i < 3; i++ {
		out = append(out, edge)
	}
	return strings.Join(out, "\n")
}
