package times

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tkuchiki/go-timezone"

	"github.com/carterpeel/go-corelib/strs"
)

type trieEntry struct {
	Pattern string
	Layout  string
	TZ      bool
}

var dateEntries = []trieEntry{
	// all date formats connected with '-'
	{Pattern: "0000-00-00", Layout: "2006-01-02"},
	{Pattern: "00-00-0000", Layout: "01-02-2006"},

	// all date formats connected with '/'
	{Pattern: "0000/00/00", Layout: "2006/01/02"},
	{Pattern: "00/00/0000", Layout: "01/02/2006"},
	{Pattern: "0/00/0000", Layout: "1/02/2006"},
	{Pattern: "0/0/0000", Layout: "1/2/2006"},
	{Pattern: "00/0/0000", Layout: "01/2/2006"},
	{Pattern: "00/00/00", Layout: "01/02/06"},

	// all date formats with no delimiter
	{Pattern: "00000000", Layout: "20060102"},
}

// The delim char between date and time.
var dateTimeDelims = []string{"T", " ", ""}

var timeEntries = []trieEntry{
	// hh:mm[:ss[.sssssssss]]
	{Pattern: "00:00:00", Layout: "15:04:05"},
	{Pattern: "00:00:00.0", Layout: "15:04:05"},
	{Pattern: "00:00:00.00", Layout: "15:04:05"},
	{Pattern: "00:00:00.000", Layout: "15:04:05"},
	{Pattern: "00:00:00.0000", Layout: "15:04:05"},
	{Pattern: "00:00:00.00000", Layout: "15:04:05"},
	{Pattern: "00:00:00.000000", Layout: "15:04:05"},
	{Pattern: "00:00:00.0000000", Layout: "15:04:05"},
	{Pattern: "00:00:00.00000000", Layout: "15:04:05"},
	{Pattern: "00:00:00.000000000", Layout: "15:04:05"},
	{Pattern: "00:00", Layout: "15:04"},

	// hhmm[ss]
	{Pattern: "000000", Layout: "150405"},
	{Pattern: "0000", Layout: "1504"},

	// hh:mm[:ss[.sssssssss]] AM
	{Pattern: "00:00:00 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.0 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.00 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.000 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.0000 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.00000 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.000000 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.0000000 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.00000000 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.000000000 AM", Layout: "03:04:05 PM"},
	{Pattern: "00:00 AM", Layout: "03:04 PM"},

	// hh:mm[:ss[.sssssssss]] PM
	{Pattern: "00:00:00 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.0 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.00 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.000 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.00000 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.000000 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.0000000 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.00000000 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.000000000 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00:00.0000000000 PM", Layout: "03:04:05 PM"},
	{Pattern: "00:00 PM", Layout: "03:04 PM"},

	// hh:mm[:ss[.sssssssss]]AM
	{Pattern: "00:00:00AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.0AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.00AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.000AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.0000AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.00000AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.000000AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.0000000AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.00000000AM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.000000000AM", Layout: "03:04:05PM"},
	{Pattern: "00:00AM", Layout: "03:04PM"},

	// hh:mm[:ss[.sssssssss]]PM
	{Pattern: "00:00:00PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.0PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.00PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.000PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.0000PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.00000PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.000000PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.0000000PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.00000000PM", Layout: "03:04:05PM"},
	{Pattern: "00:00:00.000000000PM", Layout: "03:04:05PM"},
	{Pattern: "00:00PM", Layout: "03:04PM"},

	// hhmm[ss] AM
	{Pattern: "000000 AM", Layout: "030405 PM"},
	{Pattern: "0000 AM", Layout: "0304 PM"},

	// hhmm[ss] PM
	{Pattern: "000000 PM", Layout: "030405 PM"},
	{Pattern: "0000 PM", Layout: "0304 PM"},

	// hhmm[ss]AM
	{Pattern: "000000AM", Layout: "030405PM"},
	{Pattern: "0000AM", Layout: "0304PM"},

	// hhmm[ss]PM
	{Pattern: "000000PM", Layout: "030405PM"},
	{Pattern: "0000PM", Layout: "0304PM"},
}

// The delim char between time and tz offset.
var timeTZOffsetDelims = []string{"+", "-", " +", " -"}

var tzOffsetEntries = []trieEntry{
	{Pattern: "00", Layout: "07"},
	{Pattern: "0000", Layout: "0700"},
	{Pattern: "00:00", Layout: "07:00"},
}

func digitKey(count int) uint64 {
	return uint64('d'<<32) | uint64(uint32(count))
}

const (
	spaceKey = uint64('s' << 32)
)

func keyMapper(s string, index int) (advance int, key uint64) {
	r, size := utf8.DecodeRuneInString(s[index:])
	switch {
	case unicode.IsDigit(r):
		count := 1
		for advance = index + size; advance < len(s); {
			r, size = utf8.DecodeRuneInString(s[advance:])
			if !unicode.IsDigit(r) {
				break
			}
			advance += size
			count++
		}
		return advance - index, digitKey(count)
	case unicode.IsSpace(r):
		for advance = index + size; advance < len(s); {
			r, size = utf8.DecodeRuneInString(s[advance:])
			if !unicode.IsSpace(r) {
				break
			}
			advance += size
		}
		return advance - index, spaceKey
	default:
		return size, uint64(r)
	}
}

func addToTrie(trie *strs.RuneTrie, e trieEntry) {
	if !trie.Add(e.Pattern, e) {
		panic(fmt.Sprintf("pattern '%s' caused a collision", e.Pattern))
	}
}

func initDateTimeTrie() *strs.RuneTrie {
	trie := strs.NewRuneTrie(keyMapper)
	for _, de := range dateEntries {
		// date only
		addToTrie(trie, de)
		for _, dateTimeDelim := range dateTimeDelims {
			if dateTimeDelim == "" && de.Pattern != "00000000" {
				// This is a ugly special case:
				// We have use case where the input is 202009201234 (basically 2020/09/20 12:34)
				// However we can't blindly apply "" as a delim to all date/time patterns as it
				// will cause ambiguity and collisions, such as:
				//  mm/dd/yyhhmmss vs mm/dd/yyyyhhmm
				// This using "" as delim only applies to dates being all digits.
				continue
			}
			for _, te := range timeEntries {
				// date + time
				addToTrie(
					trie,
					trieEntry{
						Pattern: de.Pattern + dateTimeDelim + te.Pattern,
						Layout:  de.Layout + dateTimeDelim + te.Layout,
					})
				// date + time + "Z"
				addToTrie(
					trie,
					trieEntry{
						Pattern: de.Pattern + dateTimeDelim + te.Pattern + "Z",
						Layout:  de.Layout + dateTimeDelim + te.Layout + "Z",
						TZ:      true,
					})
				for _, timeTZOffsetDelim := range timeTZOffsetDelims {
					for _, offset := range tzOffsetEntries {
						// date + time + tz-offset
						addToTrie(
							trie,
							trieEntry{
								Pattern: de.Pattern + dateTimeDelim + te.Pattern + timeTZOffsetDelim + offset.Pattern,
								// while in trie pattern we need '+' or '-', in actual golang time.Parse/ParseInLocation
								// call, the layout always uses '-' for tz offset. So need to replace '+' with '-'.
								Layout: de.Layout + dateTimeDelim + te.Layout +
									strings.ReplaceAll(timeTZOffsetDelim, "+", "-") + offset.Layout,
								TZ: true,
							})
					}
				}
			}
		}
	}
	return trie
}

func initTimezones() map[string]bool {
	all := timezone.New().Timezones()
	delete(all, "-00")
	tzMap := make(map[string]bool)
	for _, tzs := range all {
		for _, tz := range tzs {
			tzMap[tz] = true
		}
	}
	return tzMap
}

var dateTimeTrie *strs.RuneTrie
var allTimezones map[string]bool

func init() {
	// These two initialization routes take in total at about 10s of milliseconds. See benchmark in test.
	dateTimeTrie = initDateTimeTrie()
	allTimezones = initTimezones()
}
