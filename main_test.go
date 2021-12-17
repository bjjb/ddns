package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"gotest.tools/assert"
)

type testExecutor func() error

func (f testExecutor) Execute() error { return f() }

func Test_main(t *testing.T) {
	called := false
	defer func(f func(_ []string, _, _ io.Writer) executor) { cmd = f }(cmd)
	cmd = func(args []string, out, err io.Writer) executor {
		return testExecutor(func() error { called = true; return nil })
	}
	main()
	assert.Assert(t, called)
}

func Test_cmd(t *testing.T) {
	args := []string{"-v"}
	bout, berr := new(bytes.Buffer), new(bytes.Buffer)
	c := cmd(args, bout, berr)
	assert.NilError(t, c.Execute())
	assert.Equal(t, fmt.Sprintf("%s version %s\n", name, version), bout.String())
	assert.Equal(t, "", berr.String())
}
