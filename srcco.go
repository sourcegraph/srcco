// srcco (pronounced "source-co") is a literate-programming-style
// documentation generator that makes the code clickable so you can jump
// to the definition of any function, type, or variable.
//
// Built on top of srclib (https://srclib.org).
//
// Inspired by Docco (http://jashkenas.github.io/docco/), Groc
// (http://nevir.github.io/groc/), and Gocco
// (http://nikhilm.github.io/gocco/).
//
// Installation:
//
//   $ go get sourcegraph.com/sourcegraph/srcco
//
// And install srclib:
//
//   $ go get sourcegraph.com/sourcegraph/srclib/cmd/src
//   # This will only pull down the Go toolchain.
//   $ src toolchain install-std --skip-ruby --skip-javascript --skip-python
//
// Then call srcco like this in the directory you want to build:
//   $ srcco .
//
//   Usage: srcco [FLAGS] DIR
//
//   Generate documentation for the project at DIR.
//     -github-pages=false: create docs in gh-pages branch
//     -out="docs": The directory name for the output files
//     -v=false: show verbose output
//
// I extended the Go srclib toolchain
// (https://sourcegraph.com/sourcegraph/srclib-go) to add start and end ranges
// to comments. None of the other toolchains output this information
// currently, but it shouldn't be that hard to add.
//
// Languages that will be supported soon: (if you're interested in
// hacking on a srclib toolchain, get in touch with the author of
// srcco and I can help you get spun up)
//
// - Python
//
// - Ruby
//
// - JavaScript
//
// - Java
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

// We define our option flags here.
var (
	// verboseOpt tells srcco to print out debugging logs.
	verboseOpt bool
	// outDirOpt is the output directory for the generated
	// documentation.
	outDirOpt string
	// gitHubPagesOpt tells srcco to generate the docs in the
	// repository's "gh-pages" branch and push it to GitHub. If
	// gitHubPagesOpt is true, outDirOpt is ignored.
	gitHubPagesOpt bool
)

func init() {
	flag.BoolVar(&verboseOpt, "v", false, "show verbose output")
	flag.StringVar(&outDirOpt, "out", "docs", "The directory name for the output files")
	flag.BoolVar(&gitHubPagesOpt, "github-pages", false, "create docs in gh-pages branch and push to GitHub")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: srcco [FLAGS] DIR\n")
		fmt.Fprintf(os.Stderr, "Generate documentation for the project at DIR.\n")
		fmt.Fprintf(os.Stderr, "For more information, see:\n")
		fmt.Fprintf(os.Stderr, "\tsourcegraph.github.io/srcco\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
}

// The vLogger is used for verbose logging.
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

// failedCmd is an error type for failed shell commands.
type failedCmd struct {
	cmd interface{}
	err interface{}
}

func (f failedCmd) Error() string {
	return fmt.Sprintf("command %v failed: %s", f.cmd, f.err)
}

// ensureSrclibExists is a hack to make sure that "src" is accessible
// from the PATH.
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

// command takes a set of command line arguments and returns the cmd
// object, stdout, and stderr for that command.
func command(argv []string) (cmd *exec.Cmd, stdout *bytes.Buffer, stderr *bytes.Buffer) {
	if len(argv) == 0 {
		panic("command: argv must have at least one item")
	}
	cmd = exec.Command(argv[0], argv[1:]...)
	stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	return cmd, stdout, stderr
}

// execute is the function that does all of the work. It takes the
// project directory as dir. If dir is the empty string, then the
// current working directory is used.
func execute(dir string) error {
	// First, we check to make sure that srclib exists.
	if err := ensureSrclibExists(); err != nil {
		log.Fatal(err)
	}

	// We need to get a list of all of the files that we want to
	// generate. First, we need to turn dir into an absolute path.
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
	// We could import sourcegraph.com/sourcegraph/srclib/src and
	// call src.APIUnitsCmd.Execute, but I want to demonstrate how
	// to use src's command line interface. Plus, the user needs
	// to set up srclib with their toolchains after installing it,
	// so it might confuse them if go get'ing srcco also
	// downloaded srclib's repo.
	argv := []string{"src", "api", "units", dir}
	cmd, stdout, stderr := command(argv)
	vLogf("Running %v", argv)
	if err := cmd.Run(); err != nil {
		log.Fatal(failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}})
	}
	if stdout.Len() == 0 {
		log.Fatal(failedCmd{argv, "no output"})
	}
	// Get all of the file names associated with this project.
	var us units
	if err := json.Unmarshal(stdout.Bytes(), &us); err != nil {
		log.Fatal(err)
	}
	// If we haven't found any files, that means the user probably
	// hasn't installed any srclib language toolchains.
	// Short-circuit here if that's the case.
	noFiles := true
	for _, u := range us {
		if len(u.Files) != 0 {
			noFiles = false
			break
		}
	}
	if noFiles {
		fmt.Fprintf(os.Stderr, "srclib could not find any files for this project.\n")
		fmt.Fprintf(os.Stderr, "Have you installed any language toolchains?\n")
		fmt.Fprintf(os.Stderr, "If not, run 'src toolchain install-std'.\n")
		os.Exit(1)
	}
	if gitHubPagesOpt {
		out := ".git/srcco-tmp"
		if err := genDocs(dir, out, us.collateFiles()); err != nil {
			return err
		}
		// We need to remove all srclib build data from the
		// project directory in order to add the correct files
		// in our gh-pages script (in "data/publish-gh-pages.sh").
		argv := []string{"src", "build-data", "rm", "--all", "--local"}
		cmd, stdout, stderr := command(argv)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}}
		}
		// We pipe the gh-pages script into bash. Bash's "-s"
		// option tells it to read from stdin.
		argv = []string{"bash", "-s"}
		cmd, stdout, stderr = command(argv)
		cmd.Stdin = bytes.NewReader(ghPagesScript)
		if err := cmd.Run(); err != nil {
			return failedCmd{argv, []interface{}{err, stdout.String(), stderr.String()}}
		}
		return nil
	}
	// If we aren't generating a gh-pages site, generate the docs normally.
	return genDocs(dir, outDirOpt, us.collateFiles())
}

// doc represents a comment. srclib also gives us the definition a
// comment is attached to, but we don't care about that.
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

// A ref represents a reference to a definition. A definition is also a
// reference, it is a reference to itself. We can identify a ref's
// definition by joining DefUnit and DefPath.
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

// A def is a definition, and it includes things like functions,
// variables, and types.
type def struct {
	defKey
	Name     string
	File     string
	DefStart uint32
	DefEnd   uint32
	// We only care about TreePath to create a structured table of
	// contents.
	TreePath string
}

type defKey struct {
	Unit string
	Path string
}

func (d def) path() string {
	return d.TreePath
}

type defs []def

func (d defs) Len() int           { return len(d) }
func (d defs) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d defs) Less(i, j int) bool { return d[i].TreePath < d[j].TreePath }

var _ sort.Interface = defs{}

// genDocs generates a set of docs for the project at root for the
// code in files, and it outputs the docs in the directory siteName
// (which must be relative to root).
func genDocs(root, siteName string, files []string) error {
	vLog("Generating Docs")
	sitePath := filepath.Join(root, siteName)
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		log.Fatal(err)
	}
	// structuredTOCs is a map from file name to html-formatted
	// structured table of contents. This is created but ignored
	// for now (the ui is in the works! Contact the author of
	// srcco if you're interested in helping!)
	structuredTOCs := map[string]string{}
	// defsMap is a map from defKeys to defs. We use it to store
	// all of the defs that exist in this project so we can
	// quickly look them up. Ideally, we would use "src api
	// describe", but that call is too slow right now because it
	// doesn't hit the new, faster srclib backend... yet :)
	defsMap := map[defKey]def{}
	for _, f := range files {
		// Grab all the defs.
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
		// We create the table of contents for the defs here.
		// We wrap the defs in an interface that exposes their
		// TreePaths as Path(), so we can use
		// createTableOfContents on files too. See the
		// documentation on createTableOfContents for more
		// info.
		sort.Sort(defs(out.Defs))
		structuredTOCs[f] = createTableOfContents(defsWrapPathers(defsTOCFilter(out.Defs)))
	}
	// The files are wrapped as Pathers (which have the method
	// Path()) so that createTableOfContents can be used with defs
	// too. See the documentation on createTableOfContents for more info.
	fileTOC := createTableOfContents(filesWrapPathers(files))

	// Okay, this is where the real work gets done! We process the
	// refs for each file and generate the HTML for the code views
	// in this loop.
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

		// We filter out nonunique comments here, and comments
		// that don't have the format "text/html". I fixed a
		// bug in the Go toolchain that was generating
		// overlapping comments, so that may not be needed.
		// We're adding more powerful API commands to take
		// advantage of the new srclib backend, and I want to
		// replace this logic with a srclib call when that's
		// done.
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
		// We turn the refs into HTML annotations that can be
		// applied to the source code.
		sort.Sort(refs(out.Refs))
		anns, err := ann(src, out.Refs, f, defsMap)
		if err != nil {
			return err
		}
		htmlFile := htmlFilename(f)
		vLogf("Creating dir %s", filepath.Dir(filepath.Join(sitePath, htmlFile)))
		if err := os.MkdirAll(filepath.Dir(filepath.Join(sitePath, htmlFile)), 0755); err != nil {
			log.Fatal(err)
		}
		// Sort everything *again* just to be sure!
		sort.Sort(docs(htmlDocs))
		sort.Sort(annotations(anns))
		// Now we create the segments, which have the type
		// "segment". They are fed into the template.
		s, err := createSegments(src, anns, htmlDocs)
		if err != nil {
			return err
		}
		vLogf("Creating file %s", filepath.Join(sitePath, htmlFile))
		w, err := os.Create(filepath.Join(sitePath, htmlFile))
		if err != nil {
			return err
		}
		// After gathering all that data, we feed it into our template!
		if err := codeTemplate.Execute(w, HTMLOutput{f, resourcePrefix(f), fileTOC, structuredTOCs[f], s}); err != nil {
			return err
		}
	}
	// We copy our resource files at the end.
	if err := copyBytes(cssData, filepath.Join(sitePath, "srcco.css")); err != nil {
		return err
	}
	if err := copyBytes(jsData, filepath.Join(sitePath, "srcco.js")); err != nil {
		return err
	}
	return nil
}

// copyBytes is a helper function that copies b to file.
func copyBytes(b []byte, file string) error {
	w, err := os.Create(file)
	if err != nil {
		return err
	}
	defer w.Close()

	// do the actual work
	_, err = io.Copy(w, bytes.NewReader(b))
	return err
}

// HTMLOutput is fed into our code view template.
type HTMLOutput struct {
	Title                     string
	ResourcePrefix            string
	FileTableOfContents       string
	StructuredTableOfContents string
	Segments                  []segment
}

// These files are read from a really clever Go library, go-bindata,
// that takes resource files and packs them into the binary to make
// them easier to distribute.
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

type annotations []annotate.Annotation

func (a annotations) Len() int      { return len(a) }
func (a annotations) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a annotations) Less(i, j int) bool {
	return (a[i].Start < a[j].Start) || ((a[i].Start == a[j].Start) && a[i].End < a[j].End)
}

type htmlAnnotator syntaxhighlight.HTMLConfig

// class is copied from sourcegraph.com/sourcegraph/annotate, because
// it isn't exported in that package..
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

// Annotate is also copied from sourcegraph.com/sourcegraph/annotate
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

// resourcePrefix takes a file path and gives you the number of
// "../"'s needed to get to the "root" of that file. It's used so that
// we don't need to know the explicit root of our generated docs.
func resourcePrefix(file string) string {
	file = filepath.Clean(file)
	count := strings.Count(file, "/")
	var prefix string
	for i := 0; i < count; i++ {
		prefix += "../"
	}
	return prefix
}

// htmlFilename takes a file, prepends a resource prefix to it and
// appends ".html".
func htmlFilename(filename string) string {
	return filepath.Join(resourcePrefix(filename), filename+".html")
}

// ann is a function that takes a source file, a set of refs for that
// source file, the file name, and a map of all the defs in the
// repository, and creates a set of annotations that can be applied to
// the source file.
func ann(src []byte, refs []ref, filename string, defs map[defKey]def) ([]annotate.Annotation, error) {
	vLog("Annotating", filename)
	// Run the source code through a generic code syntax
	// highlighter to identify language units (vars, functions,
	// etc) and give them classes.
	annotations, err := syntaxhighlight.Annotate(src, htmlAnnotator(syntaxhighlight.DefaultHTMLConfig))
	if err != nil {
		return nil, err
	}
	sort.Sort(annotations)

	// refAt is a helper function that tells us the ref at a
	// certain point.
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

	// Now we go through all of the annotations, and we add links
	// to annotations that are sitting on top of refs, and that we
	// have definitions for in our def map.
	anns := make([]annotate.Annotation, 0, len(annotations))
	for _, a := range annotations {
		r, found := refAt(uint32(a.Start))
		if !found {
			a.Left = []byte(fmt.Sprintf(`<span class="%s">`, string(a.Left)))
			a.Right = []byte(`</span>`)
			anns = append(anns, *a)
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
		anns = append(anns, *a)
	}

	// Now we go through all of the defs and mark them up with
	// "invisible" anchor tags which only have an id associated
	// with them so that we can jump to them.
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
		anns = append(anns, a)
	}
	return anns, nil
}

// A segment represents a row in the final output.
type segment struct {
	DocHTML  string
	CodeHTML string
}

// createSegments takes the source code, all of the annotations, and
// the docs, and it interleaves them into segments, where docs only
// appear in the DocHTML bits, and code in the CodeHTML parts. anns
// and docs must be sorted.
func createSegments(src []byte, anns []annotate.Annotation, docs []doc) ([]segment, error) {
	vLog("Creating segments")
	var segments []segment
	var s segment
	var lineComment bool
	// addSegment is a wrapper function for appending a new
	// segment and creating a new one at 's'. It may be an abuse
	// of closures :)
	addSegment := func() {
		segments = append(segments, s)
		s = segment{}
	}
	for i := 0; i < len(src); {
		// If we're on a doc, add it to DocHTML and advance i
		// to the end of the doc.
		for len(docs) != 0 && docs[0].Start == uint32(i) {
			// This is a bit tricky, but if lineComment is
			// true, that means we've already added that
			// doc to DocHTML, so we shouldn't add it again.
			if !lineComment {
				s.DocHTML = docs[0].Data
			}
			// After ignoring the first line
			// comment, we are no longer in a line
			// comment (because they expand to the
			// rest of the line).

			i = int(docs[0].End)
			docs = docs[1:]
		}
		// Now, we need to determine how long the CodeHTML
		// block should be. It will extend to either the
		// beginning of the next comment or the end of the
		// document, whichever comes sooner.
		var runTo int
		if len(docs) != 0 {
			runTo = int(docs[0].Start)
		} else {
			runTo = len(src)
		}
		// Throw away annotations that overlapped with the
		// docs we just added.
		for len(anns) != 0 && i > anns[0].Start {
			anns = anns[1:]
		}
		// Skip any newlines (in reality, we should be
		// skipping any lines that are *all* whitespace, but
		// that requires backtracking or looking forward,
		// which I didn't want to put in v0.1.)
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
		// In this loop, we add all of the annotations to the
		// CodeHTML part of our segment.
		for i < runTo {
			// If there are no annotations left, we can
			// short-circuit this process by stuffing the
			// rest of the source code into the CodeHTML
			// block.
			if len(anns) == 0 {
				s.CodeHTML += template.HTMLEscapeString(string(src[i:runTo]))
				i = runTo
				break
			}
			// We work on one annotation at a time.
			a := anns[0]
			// Add all the space between i and a.Start to the CodeHTML block
			if i < a.Start {
				s.CodeHTML += template.HTMLEscapeString(string(src[i:a.Start]))
				i = a.Start
				// We continue so that the 'i < runTo'
				// check happens again, because we may
				// have reached runTo.
				continue
			}
			// If the annotation extends past the end of
			// our run, the state of our program is messed
			// up (usually means the srclib-cache hasn't
			// been refreshed.)
			if a.End > runTo {
				log.Println(a, string(src[a.Start:a.End]), runTo)
				log.Fatal("createSegment: illegal state: a.End > runTo")
			}
			// Now we add the annotation in full to the CodeHTML block.
			s.CodeHTML += string(a.Left) +
				template.HTMLEscapeString(string(src[a.Start:a.End])) +
				string(a.Right)
			// Advance i and anns.
			i = a.End
			anns = anns[1:]
		}
		// At the end of our loop, we add a segment.
		addSegment()
	}
	return segments, nil
}

// And that's it! We set up the flags and start the program in our
// main function.
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

// Everything below is my work in progress table of contents stuff,
// some of which you saw above. I think it's pretty interesting, but
// it's loosely commented for now. If you want to help out, email the
// author of srcco :)
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
