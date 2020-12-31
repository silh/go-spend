package log_test

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"go-spend/log"
	"io/ioutil"
	inner "log"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	tests := []struct {
		name             string
		level            int
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "trace",
			level:            log.TraceLvl,
			shouldContain:    []string{"[TRACE] 1", "[DEBUG] 2", "[INFO] 3", "[WARN] 4", "[ERROR] 5"},
			shouldNotContain: []string{},
		},
		{
			name:             "debug",
			level:            log.DebugLvl,
			shouldContain:    []string{"[DEBUG] 2", "[INFO] 3", "[WARN] 4", "[ERROR] 5"},
			shouldNotContain: []string{"[TRACE] 1"},
		},
		{
			name:             "info",
			level:            log.InfoLvl,
			shouldContain:    []string{"[INFO] 3", "[WARN] 4", "[ERROR] 5"},
			shouldNotContain: []string{"[TRACE] 1", "[DEBUG] 2"},
		},
		{
			name:             "warn",
			level:            log.WarnLvl,
			shouldContain:    []string{"[WARN] 4", "[ERROR] 5"},
			shouldNotContain: []string{"[TRACE] 1", "[DEBUG] 2", "[INFO] 3"},
		},
		{
			name:             "error",
			level:            log.ErrorLvl,
			shouldContain:    []string{"[ERROR] 5"},
			shouldNotContain: []string{"[TRACE] 1", "[DEBUG] 2", "[INFO] 3", "[WARN] 4"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result []byte
			buffer := bytes.NewBuffer(result)
			inner.SetOutput(buffer)

			log.Level = test.level

			log.Trace("1")
			log.Debug("2")
			log.Info("3")
			log.Warn("4")
			log.Error("5")

			result, err := ioutil.ReadAll(buffer)
			require.NoError(t, err)
			all := string(result)
			for _, s := range test.shouldContain {
				require.True(t, strings.Contains(all, s))
			}
			for _, s := range test.shouldNotContain {
				require.False(t, strings.Contains(all, s))
			}
		})
	}
}
