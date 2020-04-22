package override

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/influxdb/v2/kit/feature"
)

const (
	// Comma should delimit key-value pairs in the value of the configuration string.
	Comma = ","
	// Colon is the assignment operator in the key-value pairs in the configuration string, i.e. 'k:v'.
	Colon = ":"
	// These are specifically named to be documented as part of the public API, as opposed to e.g. "Delimiter"
)

// Flagger computes any flag values from a string formatted as a list of key-value pairs,
// i.e. 'k1:v1,k2:v2,...' and uses them to override their corresponding defaults.
type Flagger struct {
	flags map[string]string
}

// Make a Flagger that returns defaults with any overrides parsed from the string.
func Make(s string) (Flagger, error) {
	flags, err := parse(s)
	if err != nil {
		return Flagger{}, err
	}

	return Flagger{
		flags: flags,
	}, nil
}

func parse(s string) (map[string]string, error) {
	var (
		pairs = strings.Split(s, Comma)
		m     = make(map[string]string, len(pairs))
	)
	if len(pairs) < 1 {
		return nil, errMalformed(s)
	}
	for _, pair := range pairs {
		split := strings.Split(pair, Colon)
		if len(split) != 2 {
			return nil, errMalformed(s)
		}
		m[split[0]] = split[1]
	}

	return m, nil
}

func errMalformed(s string) error {
	return fmt.Errorf("malformed configuration string %q must match format \"k1:v1,k2:v2,...\"", s)
}

// Flags returns a map of default values. It never returns an error.
func (f Flagger) Flags(_ context.Context, flags ...feature.Flag) (map[string]interface{}, error) {
	if len(flags) == 0 {
		flags = feature.Flags()
	}

	m := make(map[string]interface{}, len(flags))
	for _, flag := range flags {
		if s, overridden := f.flags[flag.Key()]; overridden {
			iface, err := f.coerce(s, flag)
			if err != nil {
				return nil, err
			}
			m[flag.Key()] = iface
		} else {
			m[flag.Key()] = flag.Default()
		}
	}

	return m, nil
}

func (Flagger) coerce(s string, flag feature.Flag) (iface interface{}, err error) {
	switch flag.(type) {
	case feature.BoolFlag:
		iface, err = strconv.ParseBool(s)
	case feature.IntFlag:
		iface, err = strconv.Atoi(s)
	case feature.FloatFlag:
		iface, err = strconv.ParseFloat(s, 64)
	default:
		iface = s
	}

	if err != nil {
		return nil, fmt.Errorf("coercing string %q based on flag type %T: %v", s, flag, err)
	}
	return
}
