package dto

import (
	"encoding/json"
	"strconv"
)

// FlexUint unmarshals a JSON field that may be either a number (2) or a
// quoted string ("2"). Use *FlexUint in request structs where clients may
// send either form.
type FlexUint struct{ V uint }

func (f *FlexUint) UnmarshalJSON(b []byte) error {
	// Try number first.
	var n uint
	if err := json.Unmarshal(b, &n); err == nil {
		f.V = n
		return nil
	}
	// Fall back to quoted string.
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	n64, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	f.V = uint(n64)
	return nil
}
