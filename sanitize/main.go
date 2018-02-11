package sanitize

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

func Username(str string, withoutSeperators ...bool) string {
	if withoutSeperators != nil && len(withoutSeperators) > 0 {
		if withoutSeperators[0] {
			return String(str, "", regexp.MustCompile(`[^A-Za-z0-9]`))
		}
	}
	return String(str, "_", regexp.MustCompile(`[^A-Za-z0-9\_\.\-]`))
}

func URLPath(str string) string {
	return String(str, "-", regexp.MustCompile(`[?!":;$@\.,/()\[\]{}#%^*|~<>€£¥•]`))
}

// A very limited list of transliterations to catch common european names translated to urls.
// This set could be expanded with at least caps and many more characters.
var transliterations = map[rune]string{
	'À': "A",
	'Á': "A",
	'Â': "A",
	'Ã': "A",
	'Ä': "A",
	'Å': "AA",
	'Æ': "AE",
	'Ç': "C",
	'È': "E",
	'É': "E",
	'Ê': "E",
	'Ë': "E",
	'Ì': "I",
	'Í': "I",
	'Î': "I",
	'Ï': "I",
	'Ð': "D",
	'Ł': "L",
	'Ñ': "N",
	'Ò': "O",
	'Ó': "O",
	'Ô': "O",
	'Õ': "O",
	'Ö': "O",
	'Ø': "OE",
	'Ù': "U",
	'Ú': "U",
	'Ü': "U",
	'Û': "U",
	'Ý': "Y",
	'Þ': "Th",
	'ß': "ss",
	'à': "a",
	'á': "a",
	'â': "a",
	'ã': "a",
	'ä': "a",
	'å': "aa",
	'æ': "ae",
	'ç': "c",
	'è': "e",
	'é': "e",
	'ê': "e",
	'ë': "e",
	'ì': "i",
	'í': "i",
	'î': "i",
	'ï': "i",
	'ð': "d",
	'ł': "l",
	'ñ': "n",
	'ń': "n",
	'ò': "o",
	'ó': "o",
	'ô': "o",
	'õ': "o",
	'ō': "o",
	'ö': "o",
	'ø': "oe",
	'ś': "s",
	'ù': "u",
	'ú': "u",
	'û': "u",
	'ū': "u",
	'ü': "u",
	'ý': "y",
	'þ': "th",
	'ÿ': "y",
	'ż': "z",
	'Œ': "OE",
	'œ': "oe",
}

// Accents replaces a set of accented characters with ascii equivalents.
func Accents(s string) string {
	// Replace some common accent characters
	b := bytes.NewBufferString("")
	for _, c := range s {
		// Check transliterations first
		if val, ok := transliterations[c]; ok {
			b.WriteString(val)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// A list of characters we consider separators in normal strings and replace with our canonical separator - rather than removing.
var (
	separators = regexp.MustCompile(`[ &=+:]`)

	dashes = regexp.MustCompile(`[\-]+`)

	underscores = regexp.MustCompile(`[\_]+`)

	dots = regexp.MustCompile(`[\.]+`)

	spaces = regexp.MustCompile(`[ ]+`)

	illegalName = regexp.MustCompile(`[^\p{L}\d\_\.\-\s&=+':!?,]`)
)

// cleanString replaces separators with - and removes characters listed in the regexp provided from string.
// Accents, spaces, and all characters not in A-Za-z0-9 are replaced.
func String(s string, separator string, r *regexp.Regexp) string {

	// Remove any trailing space to avoid ending on -
	s = strings.Trim(s, " ")

	// Flatten accents first so that if we remove non-ascii we still get a legible name
	s = Accents(s)

	// Replace certain joining characters with a separator
	s = separators.ReplaceAllString(s, separator)

	// Remove all other unrecognised characters - NB we do allow any printable characters
	s = r.ReplaceAllString(s, "")

	// Remove any multiple separator caused by replacements above
	s = dashes.ReplaceAllString(s, "-")
	s = underscores.ReplaceAllString(s, "_")
	s = dots.ReplaceAllString(s, ".")

	return s
}

func Name(s string) string {
	s = StripTags(s)
	// Remove any trailing space to avoid ending on -
	s = strings.Trim(s, " ")

	// Remove all other unrecognised characters - NB we do allow any printable characters
	s = illegalName.ReplaceAllString(s, "")

	// Remove any multiple separator caused by replacements above
	s = spaces.ReplaceAllString(s, " ")
	return s
}

func Time(h int, m int, pmOrAm string) (bool, float64) {
	if h != 0 && pmOrAm != "" {
		if h < 1 || h > 12 {
			return false, 0
		}

		if m < 0 || m > 59 {
			return false, 0
		}

		if pmOrAm == "pm" {
			if h != 12 {
				h += 12
			}
		} else {
			if h == 12 {
				h = 0
			}
		}

		timeString := strconv.Itoa(h) + "."
		if m < 10 {
			timeString += "0" + strconv.Itoa(m)
		} else {
			timeString += strconv.Itoa(m)
		}
		time, err := strconv.ParseFloat(timeString, 64)
		if err == nil {
			return true, time
		}
		return false, 0
	}
	return false, 0
}

func YoutubeVideoID(url string) string {
	if url != "" {
		r, _ := regexp.Compile(`youtube\.com\/watch\?v=([^\&\?\/]+)`)
		for i, v := range r.FindStringSubmatch(url) {
			if i == 1 {
				return v
			}
		}
		r, _ = regexp.Compile(`youtube\.com\/embed\/([^\&\?\/]+)`)
		for i, v := range r.FindStringSubmatch(url) {
			if i == 1 {
				return v
			}
		}
		r, _ = regexp.Compile(`youtube\.com\/v\/([^\&\?\/]+)`)
		for i, v := range r.FindStringSubmatch(url) {
			if i == 1 {
				return v
			}
		}
		r, _ = regexp.Compile(`youtu\.be\/([^\&\?\/]+)`)
		for i, v := range r.FindStringSubmatch(url) {
			if i == 1 {
				return v
			}
		}
		r, _ = regexp.Compile(`youtube\.com\/verify_age\?next_url=\/watch%3Fv%3D([^\&\?\/]+)`)
		for i, v := range r.FindStringSubmatch(url) {
			if i == 1 {
				return v
			}
		}
	}
	return ""
}
