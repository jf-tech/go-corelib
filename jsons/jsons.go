package jsons

import (
	"bytes"
	"encoding/json"
	"strings"
)

const (
	prettyIndent = "\t"
)

// PrettyMarshal does a JSON marshaling of 'v' with human readable output.
func PrettyMarshal(v interface{}) (string, error) {
	valueBuf := new(bytes.Buffer)
	enc := json.NewEncoder(valueBuf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", prettyIndent)
	err := enc.Encode(v)
	if err != nil {
		return "", err
	}
	lines := strings.Split(valueBuf.String(), "\n")
	noEmptyLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(strings.TrimSpace(line)) > 0 {
			noEmptyLines = append(noEmptyLines, line)
		}
	}
	return strings.Join(noEmptyLines, "\n"), nil
}

// BestEffortPrettyMarshal does a best effort JSON marshaling of 'v' with human
// readable output. '{}' will be produced if there is any JSON marshal error. This
// function never fails.
func BestEffortPrettyMarshal(v interface{}) string {
	jsonStr, err := PrettyMarshal(v)
	if err != nil {
		return "{}"
	}
	return jsonStr
}

// BPM is a shortcut (mostly used in tests) to BestEffortPrettyMarshal.
var BPM = BestEffortPrettyMarshal

// PrettyJSON reformats a json string to be pretty
func PrettyJSON(jsonStr string) (string, error) {
	var v interface{}
	err := json.Unmarshal([]byte(jsonStr), &v)
	if err != nil {
		return "", err
	}
	return PrettyMarshal(v)
}

// BestEffortPrettyJSON reformats a json string to be pretty, ignoring any error.
func BestEffortPrettyJSON(jsonStr string) string {
	s, err := PrettyJSON(jsonStr)
	if err != nil {
		return "{}"
	}
	return s
}

// BPJ is a shortcut (mostly used in tests) to BestEffortPrettyJSON.
var BPJ = BestEffortPrettyJSON

// BestEffortMinifyMarshal returns a mninified JSON encoding of v. Any error countered will cause "{}" be returned.
func BestEffortMinifyMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// BMM is a shortcut to BestEffortMinifyMarshal
var BMM = BestEffortMinifyMarshal

// BestEffortMinifyJSON reformats a json string to be minimal, ignoring any error.
func BestEffortMinifyJSON(jsonStr string) string {
	var v interface{}
	err := json.Unmarshal([]byte(jsonStr), &v)
	if err != nil {
		return "{}"
	}
	return BestEffortMinifyMarshal(v)
}

// BMJ is a shortcut to BestEffortMinifyJSON
var BMJ = BestEffortMinifyJSON
