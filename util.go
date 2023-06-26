package cborquery

import (
	"fmt"
	"strconv"
)

func toNodeName(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int:
		return "i" + strconv.FormatInt(int64(v), 10), nil
	case int8:
		return "i" + strconv.FormatInt(int64(v), 10), nil
	case int16:
		return "i" + strconv.FormatInt(int64(v), 10), nil
	case int32:
		return "i" + strconv.FormatInt(int64(v), 10), nil
	case int64:
		return "i" + strconv.FormatInt(v, 10), nil
	case uint:
		return "u" + strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return "u" + strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return "u" + strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return "u" + strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return "u" + strconv.FormatUint(v, 10), nil
	case float32:
		return "f" + strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return "f" + strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	case fmt.Stringer:
		return v.String(), nil
	}
	return "", fmt.Errorf("type \"%T\" unsupported", value)
}
