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

type doc struct {
	Format string
	Data   string
	Start  uint32
	End    uint32
}

type ref struct {
	DefUnit string
	DefPath string
	Start   uint32
}

func genSite(root, siteName string, files []string) error {
	vLog("Generating Site")
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
		argv := []string{"src", "api", "list", "--file", filepath.Join(root, f), "--no-defs"}
		cmd, stdout, stderr := command(argv)
		vLog("Running", argv)
		if err := cmd.Run(); err != nil {
			return failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}}
		}
		out := struct {
			Refs []ref
			Docs []doc
		}{}
		if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
			log.Fatal(err)
		}

		seenHTMLDoc := make(map[struct{ start, end uint32 }]struct{})
		var htmlDocs []doc
		for _, d := range out.Docs {
			if d.Format == "text/html" {
				if _, seen := seenHTMLDoc[struct{ start, end uint32 }{d.Start, d.End}]; !seen {
					htmlDocs = append(htmlDocs, d)
					seenHTMLDoc[struct{ start, end uint32 }{d.Start, d.End}] = struct{}{}
				}
			}
		}
		htmlFile := "/" + f + ".html"
		anns, err := ann(src, out.Refs, htmlFile)
		if err != nil {
			return err
		}
		vLogf("Creating dir %s", filepath.Dir(filepath.Join(sitePath, htmlFile)))
		if err := os.MkdirAll(filepath.Dir(filepath.Join(sitePath, htmlFile)), 0755); err != nil {
			log.Fatal(err)
		}
		vLogf("Creating file %s", filepath.Join(sitePath, htmlFile))
		w, err := os.Create(filepath.Join(sitePath, htmlFile))
		if err != nil {
			return err
		}
		if err := writeHTML(w, src, anns, htmlDocs); err != nil {
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

func ann(src []byte, refs []ref, htmlFile string) ([]annotation, error) {
	vLog("Annotating", htmlFile)
	annotations, err := syntaxhighlight.Annotate(src, htmlAnnotator(syntaxhighlight.DefaultHTMLConfig))
	if err != nil {
		return nil, err
	}
	sort.Sort(annotations)

	var refAtIndex int
	refAt := func(start uint32) (r ref, found bool) {
		for refAtIndex < len(refs) {
			if refs[refAtIndex].Start == start {
				defer func() { refAtIndex++ }()
				return refs[refAtIndex], true
			} else if refs[refAtIndex].Start < start {
				refAtIndex++
			} else { // refs[refAtIndex].Start > start
				return ref{}, false
			}
		}
		return ref{}, false
	}

	anns := make([]annotation, 0, len(annotations))
	for _, a := range annotations {
		r, found := refAt(uint32(a.Start))
		if !found {
			a.Left = []byte(fmt.Sprintf(`<span class="%s">`, string(a.Left)))
			a.Right = []byte(`</span>`)
			anns = append(anns, annotation{*a, false})
			continue
		}
		a.Left = []byte(fmt.Sprintf(
			`<span class="%s" id="%s"><a href="%s">`,
			string(a.Left),
			filepath.Join(r.DefUnit, r.DefPath),
			htmlFile,
		))
		a.Right = []byte(`</span></a>`)
		anns = append(anns, annotation{*a, false})
	}

	return anns, nil
}

func writeHTML(w io.Writer, src []byte, anns []annotation, docs []doc) error {
	vLog("Writing docs")
	var leftLine int
	if _, err := w.Write([]byte(fmt.Sprintf(`<div class="left"><div leftline=%d>`, leftLine))); err != nil {
		return err
	}
	var docIndex int
	for i, b := range src {
		if b == '\n' {
			leftLine++
			if _, err := w.Write([]byte(fmt.Sprintf("</div>\n<div leftline=%d>", leftLine))); err != nil {
				return err
			}
		}
		if docIndex < len(docs) && i == int(docs[docIndex].Start) {
			if _, err := w.Write([]byte(docs[docIndex].Data)); err != nil {
				return err
			}
			docIndex++
		}
	}
	w.Write([]byte(`</div></div>`))
	inDocIndex := 0
	inDoc := func(i uint32) bool {
		// Move inDocIndex forward. TODO: explain '>='.
		for inDocIndex < len(docs) && i >= docs[inDocIndex].End {
			inDocIndex++
		}
		if inDocIndex >= len(docs) {
			return false
		}
		// i is less than the End of the doc at the doc index.
		// If it's greater than the doc's start, then we have
		// an intersection.
		if i >= docs[inDocIndex].Start {
			return true
		}
		return false
	}

	vLog("Writing annotations")
	rightLine := 0
	if _, err := w.Write([]byte(fmt.Sprintf(
		`<pre><div class="right"><div rightline=%d>`,
		rightLine,
	))); err != nil {
		return err
	}
	addDivs := func(src []byte) []byte {
		buf := &bytes.Buffer{}
		buf.Grow(len(src))
		for i, b := range src {
			if b == '\n' {
				rightLine++
				buf.Write([]byte(fmt.Sprintf("</div>\n<div rightline=%d>", rightLine)))
				continue
			}
			if !inDoc(uint32(i)) {
				buf.WriteByte(b)
			}
		}
		return buf.Bytes()
	}
	defer func() { w.Write([]byte("</div>\n</pre>")) }()
	for i := 0; i < len(src); {
		for inDoc(uint32(i)) {
			// Throw away annotations that i has passed.
			for len(anns) != 0 && i >= anns[0].Start {
				anns = anns[1:]
			}
			i++
		}
		if len(anns) == 0 {
			_, err := w.Write(addDivs(src[i:]))
			return err
		}
		a := anns[0]
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
		anns = anns[1:]
	}
	return nil
}

func Main() error {
	log.SetFlags(log.Lshortfile)
	_, err := CLI.Parse()
	return err
}
