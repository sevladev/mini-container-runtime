package units

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseMemory(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty memory value")
	}

	s = strings.ToLower(s)

	multipliers := map[string]int64{
		"k": 1024,
		"m": 1024 * 1024,
		"g": 1024 * 1024 * 1024,
	}

	for suffix, mul := range multipliers {
		if strings.HasSuffix(s, suffix) {
			val, err := strconv.ParseFloat(s[:len(s)-len(suffix)], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid memory value: %s", s)
			}
			return int64(val * float64(mul)), nil
		}
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid memory value: %s", s)
	}
	return val, nil
}
