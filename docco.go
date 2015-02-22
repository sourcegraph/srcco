// srclib-docco is a docco-like static documentation generator.
// TODO: write this in a literal style.
package srclib_docco

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/sourcegraph/annotate"
	"github.com/sourcegraph/syntaxhighlight"
)

var CLI = flags.NewNamedParser("src-docco", flags.Default)

var GlobalOpt struct {
	Verbose bool `short:"v" description:"show verbose output"`
}

var vLogger = log.New(os.Stderr, "", 0)

func vLogf(format string, v ...interface{}) {
	if !GlobalOpt.Verbose {
		return
	}
	vLogger.Printf(format, v...)
}

func vLog(v ...interface{}) {
	if !GlobalOpt.Verbose {
		return
	}
	vLogger.Println(v...)
}

func init() {
	CLI.LongDescription = "TODO"
	CLI.AddGroup("Global options", "", &GlobalOpt)

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
	// TODO: must be relative to Dir.
	SiteDirName string `long:"site-dir" description:"The directory name for the output files" default:"site"`
}

var genCmd GenCmd

// Source units have lots of information associated with them, but we
// only care about the files.
type unit struct {
	Files []string
}

type units []unit

func (us units) collateFiles() []string {
	var fs []string
	for _, u := range us {
		for _, f := range u.Files {
			fs = append(fs, f)
		}
	}
	return fs
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

func command(argv []string) (cmd *exec.Cmd, stdout *bytes.Buffer, stderr *bytes.Buffer) {
	if len(argv) == 0 {
		panic("command: argv must have at least one item")
	}
	cmd = exec.Command(argv[0], argv[1:]...)
	stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	return cmd, stdout, stderr
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
		d, err := filepath.Abs(c.Dir)
		if err != nil {
			log.Fatal(err)
		}
		dir = d
	}
	argv := []string{"src", "api", "units", dir}
	cmd, stdout, stderr := command(argv)
	vLogf("Running %v", argv)
	// TODO: remove failedCmd
	if err := cmd.Run(); err != nil {
		log.Fatal(failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}})
	}
	if stdout.Len() == 0 {
		log.Fatal(failedCmd{argv, "no output"})
	}
	var us units
	if err := json.Unmarshal(stdout.Bytes(), &us); err != nil {
		log.Fatal(err)
	}
	return genSite(dir, c.SiteDirName, us.collateFiles())
}

func genSite(root, siteName string, files []string) error {
	vLog("Generating Site")
	// Here's the plan:
	// * Make sitePath directory
	// * Annotate one file at a time
	// * Thow it into a file
	// * Ignore srclib shit for now
	sitePath := filepath.Join(root, siteName)
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		vLog("Processing", f)
		src, err := ioutil.ReadFile(filepath.Join(root, f))
		if err != nil {
			return err
		}
		htmlFile := "/" + f + ".html"
		as, err := ann(src, root, f, htmlFile)
		if err != nil {
			return err
		}
		vLog("Never get here")
		if err := os.MkdirAll(filepath.Dir(filepath.Join(sitePath, htmlFile)), 0755); err != nil {
			log.Fatal(err)
		}
		w, err := os.Create(filepath.Join(sitePath, htmlFile))
		if err != nil {
			return err
		}
		if err := writeAnns(w, src, as); err != nil {
			return err
		}
	}
	return nil
}

type annotation struct {
	annotate.Annotation
	Comment bool
}

type htmlAnnotator syntaxhighlight.HTMLConfig

func (c htmlAnnotator) class(kind syntaxhighlight.Kind) string {
	switch kind {
	case syntaxhighlight.String:
		return c.String
	case syntaxhighlight.Keyword:
		return c.Keyword
	case syntaxhighlight.Comment:
		return c.Comment
	case syntaxhighlight.Type:
		return c.Type
	case syntaxhighlight.Literal:
		return c.Literal
	case syntaxhighlight.Punctuation:
		return c.Punctuation
	case syntaxhighlight.Plaintext:
		return c.Plaintext
	case syntaxhighlight.Tag:
		return c.Tag
	case syntaxhighlight.HTMLTag:
		return c.HTMLTag
	case syntaxhighlight.HTMLAttrName:
		return c.HTMLAttrName
	case syntaxhighlight.HTMLAttrValue:
		return c.HTMLAttrValue
	case syntaxhighlight.Decimal:
		return c.Decimal
	}
	return ""
}

func (a htmlAnnotator) Annotate(start int, kind syntaxhighlight.Kind, tokText string) (*annotate.Annotation, error) {
	class := a.class(kind)
	if class == "" {
		return nil, nil
	}
	return &annotate.Annotation{
		Start: start,
		End:   start + len(tokText),
		Left:  []byte(class),
		Right: nil,
	}, nil
}

func ann(src []byte, root, file, htmlFile string) ([]annotation, error) {
	vLog("Annotating", file)
	annotations, err := syntaxhighlight.Annotate(src, htmlAnnotator(syntaxhighlight.DefaultHTMLConfig))
	if err != nil {
		return nil, err
	}
	sort.Sort(annotations)
	as := make([]annotation, 0, len(annotations))
	argv := []string{"src", "api", "list", "--file", filepath.Join(root, file)}
	cmd, stdout, stderr := command(argv)
	vLog("Running", argv)
	if err := cmd.Run(); err != nil {
		return nil, failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}}
	}
	type ref struct {
		DefUnit string
		DefPath string
		Start   uint32
	}
	refs := []ref{}
	if err := json.Unmarshal(stdout.Bytes(), &refs); err != nil {
		log.Fatal(err)
	}
	var refAtIndex int
	refAt := func(start uint32) (r ref, found bool) {
		for refAtIndex < len(refs) {
			if refs[refAtIndex].Start == start {
				refAtIndex++
				return r, true
			} else if refs[refAtIndex].Start < start {
				refAtIndex++
			} else { // refs[refAtIndex].Start > start
				return ref{}, false
			}
		}
		return ref{}, false
	}
	for _, a := range annotations {
		if string(a.Left) == "com" {
			as = append(as, annotation{*a, true})
			continue
		}
		r, found := refAt(uint32(a.Start))
		if !found {
			a.Left = []byte(fmt.Sprintf(`<span class="%s">`, string(a.Left)))
			a.Right = []byte(`</span>`)
			as = append(as, annotation{*a, false})
			continue
		}
		a.Left = []byte(fmt.Sprintf(
			`<span class="%s" id="%s"><a href="%s">`,
			string(a.Left),
			filepath.Join(r.DefUnit, r.DefPath),
			htmlFile,
		))
		a.Right = []byte(`</span></a>`)
		as = append(as, annotation{*a, false})
	}
	return as, nil
}

func writeAnns(w io.Writer, src []byte, as []annotation) error {
	vLog("Writing annotations")
	line := 0
	if _, err := w.Write([]byte(fmt.Sprintf(`<pre><div line=%d>`, line))); err != nil {
		return err
	}
	addDivs := func(src []byte) []byte {
		buf := &bytes.Buffer{}
		buf.Grow(len(src))
		for _, b := range src {
			if b == '\n' {
				line++
				buf.Write([]byte(fmt.Sprintf("</div>\n<div line=%d>", line)))
				continue
			}
			buf.WriteByte(b)
		}
		return buf.Bytes()
	}
	defer func() {
		w.Write([]byte("</div>\n</pre>"))
	}()
	for i := 0; i < len(src); {
		log.Println(i)
		if len(as) == 0 {
			_, err := w.Write(addDivs(src[i:]))
			return err
		}
		a := as[0]
		if i > a.Start {
			log.Fatal("writeAnns: illegal state: i > a.Start")
		}
		if i < a.Start {
			if _, err := w.Write(addDivs(src[i : i+1])); err != nil {
				return err
			}
			i++
			continue
		}
		// i == a.Start
		for _, b := range src[a.Start:a.End] {
			if b == '\n' {
				log.Fatal(`writeAnns: illegal state: \n in annotation`)
			}
		}
		if _, err := w.Write([]byte(strings.Join([]string{
			string(a.Left),
			string(src[a.Start:a.End]),
			string(a.Right),
		}, ""))); err != nil {
			return err
		}
		i = a.End
		as = as[1:]
	}
	return nil
}

func Main() error {
	//log.SetPrefix("")
	log.SetFlags(log.Lshortfile)
	_, err := CLI.Parse()
	return err
}
