package botclassify

import (
	"encoding/json"
	"os"
	"testing"
)

type classifyCase struct {
	UA             string `json:"ua"`
	IP             string `json:"ip"`
	ClassifyUA     string `json:"classify_ua"`
	ClassifyIP     string `json:"classify_ip"`
	ClassifyUAOrIP string `json:"classify_ua_or_ip"`
}

type groupCase struct {
	BotType string `json:"bot_type"`
	Group   string `json:"group"`
}

type fixtures struct {
	ClassifyCases []classifyCase `json:"classify_cases"`
	GroupCases    []groupCase    `json:"group_cases"`
}

func loadFixtures(t *testing.T) fixtures {
	t.Helper()
	data, err := os.ReadFile("testdata/fixtures.json")
	if err != nil {
		t.Fatalf("reading fixtures: %v", err)
	}
	var fx fixtures
	if err := json.Unmarshal(data, &fx); err != nil {
		t.Fatalf("unmarshaling fixtures: %v", err)
	}
	return fx
}

func TestClassifyParityWithPython(t *testing.T) {
	fx := loadFixtures(t)
	for _, c := range fx.ClassifyCases {
		c := c
		t.Run(c.UA+"|"+c.IP, func(t *testing.T) {
			if got := ClassifyUA(c.UA); got != c.ClassifyUA {
				t.Errorf("ClassifyUA(%q) = %q, want %q", c.UA, got, c.ClassifyUA)
			}
			if got := ClassifyIP(c.IP); got != c.ClassifyIP {
				t.Errorf("ClassifyIP(%q) = %q, want %q", c.IP, got, c.ClassifyIP)
			}
			if got := ClassifyUAOrIP(c.UA, c.IP); got != c.ClassifyUAOrIP {
				t.Errorf("ClassifyUAOrIP(%q, %q) = %q, want %q", c.UA, c.IP, got, c.ClassifyUAOrIP)
			}
		})
	}
}

func TestBotTypeToGroupParityWithPython(t *testing.T) {
	fx := loadFixtures(t)
	for _, c := range fx.GroupCases {
		c := c
		t.Run(c.BotType, func(t *testing.T) {
			if got := BotTypeToGroup(c.BotType); got != c.Group {
				t.Errorf("BotTypeToGroup(%q) = %q, want %q", c.BotType, got, c.Group)
			}
		})
	}
}

// Every BOT_PATTERNS entry must actually be present in the Go port with the
// exact same name — catches a silently-dropped or renamed pattern that the
// fixture cases above wouldn't necessarily surface if both sides dropped it.
func TestAllPatternsPresent(t *testing.T) {
	if len(botPatternDefs) != 63 {
		t.Errorf("botPatternDefs has %d entries, want 63 (matching BOT_PATTERNS in bot_classify.py)", len(botPatternDefs))
	}
	if len(ipBotRangeDefs) != 11 {
		t.Errorf("ipBotRangeDefs has %d entries, want 11 (matching _IP_BOT_RANGE_DEFS)", len(ipBotRangeDefs))
	}
}
