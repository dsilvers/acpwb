package compliance

import (
	"fmt"
	"strings"
)

// Result is a scored (possibly partial) attempt.
type Result struct {
	Course   *Course
	Attempt  int
	Answers  string
	Right    []bool // per answered question
	Correct  int
	Percent  int // of all N questions
	Passed   bool
	Complete bool
}

// Score grades an answer string. ok is false when the string isn't a valid
// answer path for this attempt (too long, or a letter past a question's
// last choice).
func Score(c *Course, attempt int, answers string) (Result, bool) {
	res := Result{Course: c, Attempt: attempt, Answers: answers}
	if len(answers) > c.N {
		return res, false
	}
	qs := Questions(c, attempt)
	for i := 0; i < len(answers); i++ {
		a := int(answers[i]) - 'a'
		if a < 0 || a >= len(qs[i].Choices) {
			return res, false
		}
		right := a == qs[i].Correct
		res.Right = append(res.Right, right)
		if right {
			res.Correct++
		}
	}
	res.Complete = len(answers) == c.N
	res.Percent = res.Correct * 100 / c.N
	res.Passed = res.Complete && res.Percent >= c.Pass
	return res, true
}

// PassingAnswers is the all-correct answer string for an attempt (used by
// the transcript to link straight to earned certificates).
func PassingAnswers(c *Course, attempt int) string {
	b := make([]byte, c.N)
	for i, q := range Questions(c, attempt) {
		b[i] = letters[q.Correct]
	}
	return string(b)
}

// CertID is the certificate number and watermark of a completed attempt.
func CertID(c *Course, attempt int, answers string) string {
	return md5Hex(fmt.Sprintf("acpwb_compliance_%s_%d_%s", c.ID, attempt, answers))[:8]
}

// DecodePath scores the answer trail in a /compliance/ path
// (…/<agency>/<course>/attempt-<k>/<answers>/…), for offline analysis of
// logged CrawlerVisit paths.
func DecodePath(path string) (Result, bool) {
	seg := strings.Split(strings.Trim(strings.TrimPrefix(path, "/compliance"), "/"), "/")
	if len(seg) < 4 {
		return Result{}, false
	}
	c, ok := ParseCourse(seg[0], seg[1])
	if !ok {
		return Result{}, false
	}
	k, ok := parseAttempt(seg[2])
	if !ok || !validAnswers(seg[3]) {
		return Result{}, false
	}
	return Score(c, k, seg[3])
}
