package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func Test_main(t *testing.T) {
	ok := false
	defer func(f func(
		[]string,
		io.Writer,
		io.Writer,
		func(string) string,
		func(int),
	) *cobra.Command) {
		command = f
	}(command)
	command = func(
		args []string,
		out, err io.Writer,
		env func(string) string,
		exit func(int),
	) *cobra.Command {
		return &cobra.Command{
			Run: func(_ *cobra.Command, _ []string) { ok = true },
		}
	}
	main()
	if !ok {
		t.Fatal("should have executed")
	}
}

func Test_command(t *testing.T) {
	args := []string{"-v"}
	bout, berr := new(bytes.Buffer), new(bytes.Buffer)
	env := getenv(env)
	exitCode := 0
	exit := func(i int) { exitCode = i }
	c := command(args, bout, berr, env, exit)
	if x, y := c.ValidArgsFunction(c, []string{}, ""); x != nil && y != 0 {
		t.Fatalf("wrong result from ValidArgsFunction: %s %d", x, y)
	}
	c.Execute()
	if berr.String() != "" {
		t.Fatalf("didn't expect anything in stderr; got %q", berr.String())
	}
	if bout.String() != fmt.Sprintf("%s version %s\n", name, version) {
		t.Fatalf("expected name and version, got %q", bout.String())
	}
	if exitCode != 0 {
		t.Fatalf("expected a 0 exit-code; got %d", exitCode)
	}
}

func Example() {
	defer func(a []string) { os.Args = a }(os.Args)
	ipProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "10.0.0.1")
	}))
	defer ipProvider.Close()
	os.Args = []string{"ddns", "-I", ipProvider.URL}
	main()
	// output:
	// 10.0.0.1
}
