package output

import (
	"encoding/json"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func PrettyString(v interface{}) string {
	log := zapr.NewLogger(zap.L())

	empJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Error(err, "Failed to marshal value")
		return ""
	}
	return string(empJSON)
}
