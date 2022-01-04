package main

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

// Test_main tests that main executes run using standard IO.
func Test_main(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		args     []string
		out, err string
		code     int
		run      func(*cobra.Command, []string)
	}{
		{"no args", []string{}, "ddns", "^Error:", 0, nil},
		{"bad args", []string{"blurp.wibble"}, "^$", "no records updated", 2, nil},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			code := 0
			bout, berr := new(bytes.Buffer), new(bytes.Buffer)

			defer func(f func(int)) { exit = f }(exit)
			exit = func(i int) { code = i }

			defer func(out, err io.Writer) { stdout, stderr = out, err }(stdout, stderr)
			stdout, stderr = bout, berr

			defer func(a []string) { args = a }(args)
			args = tc.args

			defer func(f func(*cobra.Command, []string)) { run = f }(run)
			if tc.run != nil {
				run = tc.run
			}

			main()
			assert.Equal(t, tc.code, code)
			assert.Assert(t, regexp.MustCompile(tc.out).MatchString(strings.TrimSpace(bout.String())))
			assert.Assert(t, regexp.MustCompile(tc.err).MatchString(strings.TrimSpace(berr.String())))
		})
	}
}
