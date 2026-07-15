package adapt

import "github.com/ml444/gkit/log"

func ShouldLog(syncEnabled bool, msgLevel log.LogLevel) bool {
	if !syncEnabled {
		return true
	}
	return msgLevel >= log.CurrentLevel()
}
