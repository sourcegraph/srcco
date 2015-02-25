// srcco is a Docco-like static documentation generator.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/sourcegraph/annotate"
	"github.com/sourcegraph/syntaxhighlight"
)

var (
	verboseOpt     bool
	outDirOpt      string
	gitHubPagesOpt bool
)

func init() {
	flag.BoolVar(&verboseOpt, "v", false, "show verbose output")
	flag.StringVar(&outDirOpt, "out", "docs", "The directory name for the output files")
	flag.BoolVar(&gitHubPagesOpt, "github-pages", false, "create docs in gh-pages branch")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: srcco [FLAGS] DIR\n")
		fmt.Fprintf(os.Stderr, "Generate documentation for the project at DIR.\n")
		fmt.Fprintf(os.Stderr, "For more information, see:\n")
		fmt.Fprintf(os.Stderr, "\tsourcegraph.github.io/srcco\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
}

var vLogger = log.New(os.Stderr, "", 0)

func vLogf(format string, v ...interface{}) {
	if !verboseOpt {
		return
	}
	vLogger.Printf(format, v...)
}

func vLog(v ...interface{}) {
	if !verboseOpt {
		return
	}
	vLogger.Println(v...)
}

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

func execute(dir string) error {
	if err := ensureSrclibExists(); err != nil {
		log.Fatal(err)
	}

	// First, we need to get a list of all of the files that we
	// want to generate.
	//
	// We could import sourcegraph.com/sourcegraph/srclib, but I
	// want to demonstrate how to use its command line interface.
	if dir == "" {
		d, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dir = d
	} else {
		d, err := filepath.Abs(dir)
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
	if gitHubPagesOpt {
		out := ".git/srcco-tmp"
		if err := genSite(dir, out, us.collateFiles()); err != nil {
			return err
		}
		argv := []string{"src", "build-data", "rm", "--all", "--local"}
		cmd, stdout, stderr := command(argv)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}}
		}
		argv = []string{"bash", "-s"}
		cmd, stdout, stderr = command(argv)
		cmd.Stdin = bytes.NewReader(ghPagesScript)
		if err := cmd.Run(); err != nil {
			return failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}}
		}
		return nil
	}
	return genSite(dir, outDirOpt, us.collateFiles())
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

type defKey struct {
	Unit string
	Path string
}

type def struct {
	defKey
	Name     string
	File     string
	TreePath string
	DefStart uint32
	DefEnd   uint32
}

func (d def) path() string {
	return d.TreePath
}

type defs []def

func (d defs) Len() int           { return len(d) }
func (d defs) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d defs) Less(i, j int) bool { return d[i].TreePath < d[j].TreePath }

var _ sort.Interface = defs{}

func genSite(root, siteName string, files []string) error {
	vLog("Generating Site")
	sitePath := filepath.Join(root, siteName)
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		log.Fatal(err)
	}
	structuredTOCs := map[string]string{}
	defsMap := map[defKey]def{}
	for _, f := range files {
		argv := []string{"src", "api", "list", "--file", filepath.Join(root, f), "--no-refs", "--no-docs"}
		cmd, stdout, stderr := command(argv)
		vLog("Running", argv)
		if err := cmd.Run(); err != nil {
			return failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}}
		}
		var out struct{ Defs []def }
		if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
			log.Fatal(err)
		}
		for _, d := range out.Defs {
			defsMap[d.defKey] = d
		}
		sort.Sort(defs(out.Defs))
		structuredTOCs[f] = createTableOfContents(defsWrapPathers(defsTOCFilter(out.Defs)))
	}
	fileTOC := createTableOfContents(filesWrapPathers(files))

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

		seenHTMLDoc := map[struct{ start, end uint32 }]bool{}
		var htmlDocs []doc
		for _, d := range out.Docs {
			if d.Format == "text/html" {
				if !seenHTMLDoc[struct{ start, end uint32 }{d.Start, d.End}] {
					htmlDocs = append(htmlDocs, d)
					seenHTMLDoc[struct{ start, end uint32 }{d.Start, d.End}] = true
				}
			}
		}
		sort.Sort(docs(htmlDocs))
		sort.Sort(refs(out.Refs)) // May be redundant.
		anns, err := ann(src, out.Refs, f, defsMap)
		if err != nil {
			return err
		}
		htmlFile := htmlFilename(f)
		vLogf("Creating dir %s", filepath.Dir(filepath.Join(sitePath, htmlFile)))
		if err := os.MkdirAll(filepath.Dir(filepath.Join(sitePath, htmlFile)), 0755); err != nil {
			log.Fatal(err)
		}
		sort.Sort(annotations(anns))
		s, err := createSegments(src, anns, htmlDocs)
		if err != nil {
			return err
		}
		vLogf("Creating file %s", filepath.Join(sitePath, htmlFile))
		w, err := os.Create(filepath.Join(sitePath, htmlFile))
		if err != nil {
			return err
		}

		if err := codeTemplate.Execute(w, HTMLOutput{f, resourcePrefix(f), fileTOC, structuredTOCs[f], s}); err != nil {
			return err
		}
	}
	if err := copyBytes(cssData, filepath.Join(sitePath, "srcco.css")); err != nil {
		return err
	}
	if err := copyBytes(jsData, filepath.Join(sitePath, "srcco.js")); err != nil {
		return err
	}
	return nil
}

func resourcePrefix(file string) string {
	file = filepath.Clean(file)
	count := strings.Count(file, "/")
	var prefix string
	for i := 0; i < count; i++ {
		prefix += "../"
	}
	return prefix
}

func copyBytes(b []byte, there string) error {
	w, err := os.Create(there)
	if err != nil {
		return err
	}
	defer w.Close()

	// do the actual work
	_, err = io.Copy(w, bytes.NewReader(b))
	return err
}

type tocNode struct {
	name    string
	nodes   []*tocNode
	pathers []pather
}

type pather interface {
	path() string
}

func defsTOCFilter(defs []def) []def {
	f := defs[:0]
	m := map[def]bool{}
	for _, d := range defs {
		if !m[d] {
			m[d] = true
			d.TreePath = strings.TrimPrefix(d.TreePath, "./")
			f = append(f, d)
		}
	}
	return f
}

func defsWrapPathers(defs []def) []pather {
	ps := make([]pather, 0, len(defs))
	for _, d := range defs {
		ps = append(ps, pather(d))
	}
	return ps
}

type file string

func (f file) path() string { return string(f) }

func filesWrapPathers(files []string) []pather {
	ps := make([]pather, 0, len(files))
	for _, f := range files {
		ps = append(ps, pather(file(f)))
	}
	return ps
}

func createTableOfContents(pathers []pather) string {
	nodes := map[string]*tocNode{}
	nodes["/"] = &tocNode{name: "/"}

	getParent := func(i int, parts []string) *tocNode {
		if i == 0 {
			return nodes["/"]
		}
		parent := nodes[strings.Join(parts[0:i], "/")]
		if parent == nil {
			log.Fatal("createTOC: illegal state: parent == nil")
		}
		return parent
	}
	// We've created the head node and added it to our map. Now,
	// we need to go through all of the pathers and add them to the
	// correct node. To do this, we get the "name" for each
	// section of the tree path. For instance, for the tree path
	// "a/b/c/d", the nodes associated with it are "a", "a/b", and
	// "a/b/c", and the pather, "d", should be added to "a/b/c".
	for _, pather := range pathers {
		// First, we split the path into parts separated by
		// "/".
		parts := strings.Split(pather.path(), "/")
		// Now we walk through the parts.
		for i, name := range parts {
			// If "i" is the last index of "parts", that
			// means it represents the pather and we need to
			// add it to a node.
			if i == len(parts)-1 {
				parent := getParent(i, parts)
				// Now we add this pather to the parent
				// and break out of the loop.
				parent.pathers = append(parent.pathers, pather)
				break
			}
			// "i" is not the last index of "parts", which
			// means that it is a node. First, we check to
			// see if it exists.
			path := strings.Join(parts[0:i+1], "/")
			// If "n" is non-nil, then the node has
			// already been created.
			if n := nodes[path]; n != nil {
				continue
			}
			// The node does not exist. First we add it to
			// the nodes map, and then we add it to its
			// parent.
			nodes[path] = &tocNode{name: name}
			parent := getParent(i, parts)
			parent.nodes = append(parent.nodes, nodes[path])
		}
	}
	var patherToHTML func(p pather) string
	if len(pathers) != 0 {
		switch pathers[0].(type) {
		case def:
			patherToHTML = func(p pather) string {
				d := p.(def)
				return fmt.Sprintf(`<a class="def" href="%s">%s</a>`,
					htmlFilename(d.File)+"#"+filepath.Join(d.Unit, d.Path),
					d.Name,
				)
			}
		case file:
			patherToHTML = func(p pather) string {
				f := string(p.(file))
				return fmt.Sprintf(`<a class="file" href="%s">%s</a>`,
					htmlFilename(f),
					filepath.Base(f),
				)
			}
		default:
			log.Fatal("createStructuredTOC: illegal state, pather's concrete type unknown")
		}
	}
	var nodeToHTML func(n tocNode) string
	nodeToHTML = func(n tocNode) string {
		title := fmt.Sprintf(`<div class="node"><div class="node-title">%s</div>`, n.name)
		body := `<div class="node-body">`
		for _, c := range n.nodes {
			body += nodeToHTML(*c)
		}
		for _, p := range n.pathers {
			body += patherToHTML(p)
		}
		body += "</div></div>"
		return title + "\n" + body
	}
	return nodeToHTML(*nodes["/"])
}

type HTMLOutput struct {
	Title                     string
	ResourcePrefix            string
	FileTableOfContents       string
	StructuredTableOfContents string
	Segments                  []segment
}

var codeTemplate *template.Template
var cssData []byte
var jsData []byte
var ghPagesScript []byte

func init() {
	r, err := Asset("data/view.html")
	if err != nil {
		log.Fatal(err)
	}
	codeTemplate = template.Must(template.New("view.html").Parse(string(r)))
	r, err = Asset("data/srcco.css")
	if err != nil {
		log.Fatal(err)
	}
	cssData = r
	r, err = Asset("data/srcco.js")
	if err != nil {
		log.Fatal(err)
	}
	jsData = r
	r, err = Asset("data/publish-gh-pages.sh")
	if err != nil {
		log.Fatal(err)
	}
	ghPagesScript = r
}

type annotation struct {
	annotate.Annotation
	Comment bool
}

type annotations []annotation

func (a annotations) Len() int      { return len(a) }
func (a annotations) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a annotations) Less(i, j int) bool {
	return (a[i].Start < a[j].Start) || ((a[i].Start == a[j].Start) && a[i].End < a[j].End)
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
	return filepath.Join(resourcePrefix(filename), filename+".html")
}

func ann(src []byte, refs []ref, filename string, defs map[defKey]def) ([]annotation, error) {
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
		if d, ok := defs[defKey{r.DefUnit, r.DefPath}]; ok {
			id := filepath.Join(d.Unit, d.Path)
			href := htmlFilename(d.File) + "#" + id
			a.Left = []byte(fmt.Sprintf(
				`<span class="%s"><a href="%s">`,
				string(a.Left),
				href,
			))
			a.Right = []byte(`</span></a>`)
		} else {
			a.Left = []byte(fmt.Sprintf(`<span class="%s">`, string(a.Left)))
			a.Right = []byte(`</span>`)
		}
		anns = append(anns, annotation{*a, false})
	}

	for _, d := range defs {
		if d.File != filename {
			continue
		}
		a := annotate.Annotation{
			Left:  []byte(fmt.Sprintf(`<span class="def" id="%s">`, filepath.Join(d.Unit, d.Path))),
			Right: []byte("</span>"),
			Start: int(d.DefStart),
			End:   int(d.DefStart),
		}
		anns = append(anns, annotation{a, false})
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
	var lineComment bool
	addSegment := func() {
		segments = append(segments, s)
		s = segment{}
	}
	for i := 0; i < len(src); {
		for len(docs) != 0 && docs[0].Start == uint32(i) {
			// Add doc
			if !lineComment {
				s.DocHTML = docs[0].Data
				lineComment = false
			}
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
		for i < len(src) && src[i] == '\n' {
			i++
		}
		// Special case: check to see if there's a newline
		// between i and runTo. If there isn't, that means
		// there's a line comment on the next line, and it
		// should begin a new section.
		if len(docs) != 0 {
			lineComment = true
			for j := i; j < runTo; j++ {
				if src[j] == '\n' {
					lineComment = false
					break
				}
			}
			if lineComment {
				addSegment()
				s.DocHTML = docs[0].Data
			}
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
		addSegment()
	}
	return segments, nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: must provide a root directory\n")
		flag.Usage()
	} else if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "error: too many args\n")
		flag.Usage()
	}
	if err := execute(args[0]); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
