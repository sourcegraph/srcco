// srclib-docco is a docco-like static documentation generator.
// TODO: write this in a literal style.
package srclib_docco

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"text/template"

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

type docs []doc

func (d docs) Len() int      { return len(d) }
func (d docs) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d docs) Less(i, j int) bool {
	return d[i].Start < d[j].Start || (d[i].Start == d[j].Start && d[i].End < d[j].End)
}

var _ sort.Interface = docs{}

type ref struct {
	DefUnit string
	DefPath string
	File    string
	Start   uint32
}

type refs []ref

func (r refs) Len() int           { return len(r) }
func (r refs) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r refs) Less(i, j int) bool { return r[i].Start < r[j].Start }

var _ sort.Interface = refs{}

func genSite(root, siteName string, files []string) error {
	vLog("Generating Site")
	sitePath := filepath.Join(root, siteName)
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		log.Fatal(err)
	}
	filesMap := map[string]struct{}{}
	for _, f := range files {
		filesMap[f] = struct{}{}
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
		sort.Sort(docs(htmlDocs))
		sort.Sort(refs(out.Refs))
		anns, err := ann(src, out.Refs, f, filesMap)
		if err != nil {
			return err
		}
		htmlFile := htmlFilename(f)
		vLogf("Creating dir %s", filepath.Dir(filepath.Join(sitePath, htmlFile)))
		if err := os.MkdirAll(filepath.Dir(filepath.Join(sitePath, htmlFile)), 0755); err != nil {
			log.Fatal(err)
		}
		s, err := createSegments(src, anns, htmlDocs)
		if err != nil {
			return err
		}
		vLogf("Creating file %s", filepath.Join(sitePath, htmlFile))
		w, err := os.Create(filepath.Join(sitePath, htmlFile))
		if err != nil {
			return err
		}
		if err := codeTemplate.Execute(w, HTMLOutput{f, s}); err != nil {
			return err
		}
	}
	return nil
}

type HTMLOutput struct {
	Title    string
	Segments []segment
}

var codeTemplate = template.Must(template.New("code").Parse(codeText))

var codeText = `
<!DOCTYPE html>
<html>
  <head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap.min.css">
    <style>
.code {
    white-space: pre-wrap;
}
    </style>
  </head>
  <body>
    <div class="container">
      {{ range .Segments}}
      <div class="row">
        <div class="left col-xs-4">{{.DocHTML}}</div>
        <div class="right col-xs-8 code">{{.CodeHTML}}</div>
      </div>
      {{ end }}
    </div>
  </body>
</html>
`

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

func htmlFilename(filename string) string {
	return filepath.Join("/", filename+".html")
}

func ann(src []byte, refs []ref, filename string, filesMap map[string]struct{}) ([]annotation, error) {
	vLog("Annotating", filename)
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
		id := filepath.Join(r.DefUnit, r.DefPath)
		var href string
		if _, in := filesMap[r.File]; in {
			href = htmlFilename(r.File) + "#" + id
			a.Left = []byte(fmt.Sprintf(
				`<span class="%s" id="%s"><a href="%s">`,
				string(a.Left),
				id,
				href,
			))
			a.Right = []byte(`</span></a>`)
		} else {
			a.Left = []byte(fmt.Sprintf(
				`<span class="%s" id="%s">`,
				string(a.Left),
				id,
			))
			a.Right = []byte(`</span>`)
		}
		anns = append(anns, annotation{*a, false})
	}

	return anns, nil
}

type segment struct {
	DocHTML  string
	CodeHTML string
}

// anns and docs must be sorted.
func createSegments(src []byte, anns []annotation, docs []doc) ([]segment, error) {
	vLog("Creating segments")
	var segments []segment
	var s segment
	for i := 0; i < len(src); {
		for len(docs) != 0 && docs[0].Start == uint32(i) {
			// Add doc
			s.DocHTML = docs[0].Data
			i = int(docs[0].End)
			docs = docs[1:]
		}
		var runTo int
		if len(docs) != 0 {
			runTo = int(docs[0].Start)
		} else {
			runTo = len(src)
		}
		for len(anns) != 0 && i > anns[0].Start {
			anns = anns[1:]
		}
		for src[i] == '\n' {
			i++
		}
		for i < runTo {
			if len(anns) == 0 {
				s.CodeHTML += template.HTMLEscapeString(string(src[i:runTo]))
				i = runTo
				break
			}
			a := anns[0]
			if i < a.Start {
				s.CodeHTML += template.HTMLEscapeString(string(src[i:a.Start]))
				i = a.Start
				continue
			}
			if a.End > runTo {
				log.Fatal("createSegment: illegal state: a.End > runTo")
			}
			s.CodeHTML += string(a.Left) +
				template.HTMLEscapeString(string(src[a.Start:a.End])) +
				string(a.Right)
			i = a.End
			anns = anns[1:]
		}
		segments = append(segments, s)
		s = segment{}
	}
	return segments, nil
}

func Main() error {
	log.SetFlags(log.Lshortfile)
	_, err := CLI.Parse()
	return err
}
