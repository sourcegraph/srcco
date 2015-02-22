// srclib-docco is a docco-like static documentation generator.
// TODO: write this in a literal style.
package srclib_docco

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/jessevdk/go-flags"
)

var CLI = flags.NewNamedParser("src-docco", flags.Default)

// TODO: add a help command.

func init() {
	CLI.LongDescription = "TODO"
	//CLI.AddGroup("Global options", "", &GlobalOpt)

	_, err := CLI.AddCommand("gen",
		"generate documentation",
		"Generate docco-like documentation for a thing",
		&genCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: add limitations help output.

type GenCmd struct {
	Dir string `long:"dir" description:"The root directory for the project"`
}

var genCmd GenCmd

type unit struct {
	Files []string
}

type failedCmd struct {
	cmd interface{}
	err interface{}
}

// TODO: Is this correct?
func (f failedCmd) Error() string {
	return fmt.Sprintf("command %v failed: %s", f.cmd, f.err)
}

func ensureSrclibExists() error {
	cmd := exec.Command("src", "version")
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"error with srclib: %v, %s, %s",
			err,
			stdout.String(),
			stderr.String(),
		)
	}
	return nil
}

func (c *GenCmd) Execute(args []string) error {
	if err := ensureSrclibExists(); err != nil {
		log.Fatal(err)
	}

	// First, we need to get a list of all of the files that we
	// want to generate.
	//
	// We could import sourcegraph.com/sourcegraph/srclib, but I
	// want to demonstrate how to use its command line interface.
	var dir string
	if c.Dir == "" {
		d, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dir = d
	} else {
		dir = c.Dir
	}
	argv := []string{"src", "api", "units", dir}
	cmd := exec.Command(argv[0], argv[1:]...)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	// TODO: remove failedCmd
	if err := cmd.Run(); err != nil {
		log.Fatal(failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}})
	}
	if stderr.Len() != 0 {
		log.Fatal(failedCmd{argv, stderr})
	}
	if stdout.Len() == 0 {
		log.Fatal(failedCmd{argv, "no output"})
	}
	var units []unit
	if err := json.Unmarshal(stdout.Bytes(), &units); err != nil {
		log.Fatal(err)
	}
	fmt.Println(units)
	return nil
}

func Main() error {
	//log.SetFlags(0)
	//log.SetPrefix("")
	log.SetFlags(log.Lshortfile)
	_, err := CLI.Parse()
	return err
}
