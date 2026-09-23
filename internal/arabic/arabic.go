// Package arabic prepares Arabic-script text for terminals without bidi
// support: letters are replaced by their contextual presentation forms and
// each line is reordered for left-to-right display.
package arabic

import (
	"slices"
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
)

type joining int

const (
	nonJoining joining = iota
	rightJoining
	dualJoining
	transparent
)

// forms holds isolated, final, initial and medial presentation forms; right
// joining letters only have the first two.
var forms = map[rune][]rune{
	'ء': {0xFE80},
	'آ': {0xFE81, 0xFE82},
	'أ': {0xFE83, 0xFE84},
	'ؤ': {0xFE85, 0xFE86},
	'إ': {0xFE87, 0xFE88},
	'ئ': {0xFE89, 0xFE8A, 0xFE8B, 0xFE8C},
	'ا': {0xFE8D, 0xFE8E},
	'ب': {0xFE8F, 0xFE90, 0xFE91, 0xFE92},
	'ة': {0xFE93, 0xFE94},
	'ت': {0xFE95, 0xFE96, 0xFE97, 0xFE98},
	'ث': {0xFE99, 0xFE9A, 0xFE9B, 0xFE9C},
	'ج': {0xFE9D, 0xFE9E, 0xFE9F, 0xFEA0},
	'ح': {0xFEA1, 0xFEA2, 0xFEA3, 0xFEA4},
	'خ': {0xFEA5, 0xFEA6, 0xFEA7, 0xFEA8},
	'د': {0xFEA9, 0xFEAA},
	'ذ': {0xFEAB, 0xFEAC},
	'ر': {0xFEAD, 0xFEAE},
	'ز': {0xFEAF, 0xFEB0},
	'س': {0xFEB1, 0xFEB2, 0xFEB3, 0xFEB4},
	'ش': {0xFEB5, 0xFEB6, 0xFEB7, 0xFEB8},
	'ص': {0xFEB9, 0xFEBA, 0xFEBB, 0xFEBC},
	'ض': {0xFEBD, 0xFEBE, 0xFEBF, 0xFEC0},
	'ط': {0xFEC1, 0xFEC2, 0xFEC3, 0xFEC4},
	'ظ': {0xFEC5, 0xFEC6, 0xFEC7, 0xFEC8},
	'ع': {0xFEC9, 0xFECA, 0xFECB, 0xFECC},
	'غ': {0xFECD, 0xFECE, 0xFECF, 0xFED0},
	'ـ': {0x0640, 0x0640, 0x0640, 0x0640},
	'ف': {0xFED1, 0xFED2, 0xFED3, 0xFED4},
	'ق': {0xFED5, 0xFED6, 0xFED7, 0xFED8},
	'ك': {0xFED9, 0xFEDA, 0xFEDB, 0xFEDC},
	'ل': {0xFEDD, 0xFEDE, 0xFEDF, 0xFEE0},
	'م': {0xFEE1, 0xFEE2, 0xFEE3, 0xFEE4},
	'ن': {0xFEE5, 0xFEE6, 0xFEE7, 0xFEE8},
	'ه': {0xFEE9, 0xFEEA, 0xFEEB, 0xFEEC},
	'و': {0xFEED, 0xFEEE},
	'ى': {0xFEEF, 0xFEF0, 0xFBE8, 0xFBE9},
	'ي': {0xFEF1, 0xFEF2, 0xFEF3, 0xFEF4},
	'ٱ': {0xFB50, 0xFB51},
	'ٹ': {0xFB66, 0xFB67, 0xFB68, 0xFB69},
	'پ': {0xFB56, 0xFB57, 0xFB58, 0xFB59},
	'چ': {0xFB7A, 0xFB7B, 0xFB7C, 0xFB7D},
	'ڈ': {0xFB88, 0xFB89},
	'ڑ': {0xFB8C, 0xFB8D},
	'ژ': {0xFB8A, 0xFB8B},
	'ک': {0xFB8E, 0xFB8F, 0xFB90, 0xFB91},
	'گ': {0xFB92, 0xFB93, 0xFB94, 0xFB95},
	'ں': {0xFB9E, 0xFB9F},
	'ھ': {0xFBAA, 0xFBAB, 0xFBAC, 0xFBAD},
	'ۀ': {0xFBA4, 0xFBA5},
	'ہ': {0xFBA6, 0xFBA7, 0xFBA8, 0xFBA9},
	'ی': {0xFBFC, 0xFBFD, 0xFBFE, 0xFBFF},
	'ے': {0xFBAE, 0xFBAF},
	'ۓ': {0xFBB0, 0xFBB1},
}

// lamAlef maps the alef that follows a lam to the isolated and final forms
// of their ligature.
var lamAlef = map[rune][2]rune{
	'آ': {0xFEF5, 0xFEF6},
	'أ': {0xFEF7, 0xFEF8},
	'إ': {0xFEF9, 0xFEFA},
	'ا': {0xFEFB, 0xFEFC},
}

const (
	lam       = 'ل'
	isolated  = 0
	final     = 1
	initial   = 2
	medial    = 3
)

func joiningOf(r rune) joining {
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || r == '‍' {
		return transparent
	}
	f, ok := forms[r]
	if !ok || r == 'ء' {
		return nonJoining
	}
	if len(f) == 2 {
		return rightJoining
	}
	return dualJoining
}

// Shape replaces Arabic letters by their contextual presentation forms,
// keeping logical order.
func Shape(text string) string {
	runes := []rune(text)
	out := make([]rune, 0, len(runes))
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		kind := joiningOf(r)
		if kind == nonJoining || kind == transparent {
			out = append(out, r)
			continue
		}
		joinsBefore := canJoinForward(runes, previousBase(runes, i))
		next := nextBase(runes, i)
		if r == lam && next >= 0 {
			ligature, ok := lamAlef[runes[next]]
			if ok {
				out = append(out, ligature[boolIndex(joinsBefore)])
				out = append(out, runes[i+1:next]...)
				i = next
				continue
			}
		}
		joinsAfter := kind == dualJoining && next >= 0 && joiningOf(runes[next]) != nonJoining
		out = append(out, forms[r][formIndex(kind, joinsBefore, joinsAfter)])
	}
	return string(out)
}

func formIndex(kind joining, joinsBefore bool, joinsAfter bool) int {
	if kind == rightJoining {
		return boolIndex(joinsBefore)
	}
	switch {
	case joinsBefore && joinsAfter:
		return medial
	case joinsAfter:
		return initial
	case joinsBefore:
		return final
	}
	return isolated
}

func boolIndex(b bool) int {
	if !b {
		return 0
	}
	return 1
}

func previousBase(runes []rune, i int) int {
	for j := i - 1; j >= 0; j-- {
		if joiningOf(runes[j]) != transparent {
			return j
		}
	}
	return -1
}

func nextBase(runes []rune, i int) int {
	for j := i + 1; j < len(runes); j++ {
		if joiningOf(runes[j]) != transparent {
			return j
		}
	}
	return -1
}

func canJoinForward(runes []rune, i int) bool {
	return i >= 0 && joiningOf(runes[i]) == dualJoining
}

// Visual shapes text and reorders one line for a left-to-right terminal:
// right-to-left runs are reversed by grapheme cluster, while runs of Latin
// letters and digits keep their internal order.
func Visual(line string) string {
	runs := splitRuns(Shape(line))
	slices.Reverse(runs)
	var b strings.Builder
	for _, run := range runs {
		if run.ltr {
			b.WriteString(run.text)
			continue
		}
		b.WriteString(reverseClusters(run.text))
	}
	return b.String()
}

type run struct {
	text string
	ltr  bool
}

func splitRuns(text string) []run {
	runs := []run{}
	var current strings.Builder
	currentLTR := false
	for _, r := range text {
		ltr := isLTR(r)
		if current.Len() > 0 && ltr != currentLTR {
			runs = append(runs, run{text: current.String(), ltr: currentLTR})
			current.Reset()
		}
		currentLTR = ltr
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		runs = append(runs, run{text: current.String(), ltr: currentLTR})
	}
	return runs
}

func isLTR(r rune) bool {
	return unicode.IsDigit(r) || r < 0x0590 && unicode.IsLetter(r)
}

var mirrored = map[string]string{"(": ")", ")": "(", "[": "]", "]": "[", "{": "}", "}": "{", "«": "»", "»": "«", "<": ">", ">": "<"}

func reverseClusters(text string) string {
	clusters := []string{}
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		cluster := graphemes.Str()
		if swapped, ok := mirrored[cluster]; ok {
			cluster = swapped
		}
		clusters = append(clusters, cluster)
	}
	slices.Reverse(clusters)
	return strings.Join(clusters, "")
}

// IsRTL reports whether text starts with a right-to-left letter.
func IsRTL(text string) bool {
	for _, r := range text {
		if !unicode.IsLetter(r) {
			continue
		}
		return unicode.In(r, unicode.Arabic, unicode.Hebrew)
	}
	return false
}
