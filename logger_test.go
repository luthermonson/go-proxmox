package proxmox

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeveledLogger(t *testing.T) {
	tests := []struct {
		level  int
		input  string
		stdout string
		stderr string
	}{
		{
			level:  LevelError,
			input:  "log",
			stderr: "[ERROR] error log\n",
			stdout: "",
		},
		{
			level:  LevelWarn,
			input:  "log",
			stderr: "[ERROR] error log\n[WARN] warn log\n",
			stdout: "",
		},
		{
			level:  LevelInfo,
			input:  "log",
			stderr: "[ERROR] error log\n[WARN] warn log\n",
			stdout: "[INFO] info log\n",
		},
		{
			level:  LevelDebug,
			input:  "log",
			stderr: "[ERROR] error log\n[WARN] warn log\n",
			stdout: "[INFO] info log\n[DEBUG] debug log\n",
		},
	}

	for _, test := range tests {
		err := &bytes.Buffer{}
		out := &bytes.Buffer{}
		log := &LeveledLogger{Level: test.level, stderrOverride: err, stdoutOverride: out}

		log.Errorf("error %s", test.input)
		log.Warnf("warn %s", test.input)
		log.Infof("info %s", test.input)
		log.Debugf("debug %s", test.input)
		assert.Equal(t, test.stdout, out.String())
		assert.Equal(t, test.stderr, err.String())
	}

	log := &LeveledLogger{Level: LevelDebug}
	assert.Equal(t, os.Stderr, log.stderr())
	assert.Equal(t, os.Stdout, log.stdout())
}
