package p

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"rogchap.com/v8go"
)

//go:embed automerge.js
var automergeLib string

var (
	v8isolateOnce sync.Once
	v8isolate     *v8go.Isolate
	v8isoErr      error
	v8isoMu       sync.Mutex // guards method calls on v8isolate
)

func newV8Context() (*v8go.Context, error) {
	v8isolateOnce.Do(func() {
		v8isolate, v8isoErr = v8go.NewIsolate()
	})
	if v8isoErr != nil {
		return nil, v8isoErr
	}
	v8isoMu.Lock()
	defer v8isoMu.Unlock()
	return v8go.NewContext(v8isolate)
}

// A Doc is an automerge document.
type Doc struct {
	v8 *v8go.Context
	mu sync.Mutex // protects javascript execution
}

// NewDoc creates a new automerge document.
func NewDoc() (*Doc, error) {
	d := new(Doc)
	var err error
	d.v8, err = newV8Context()
	if err != nil {
		return nil, err
	}
	_, err = d.v8.RunScript(automergeLib, "automerge.js")
	if err != nil {
		return nil, err
	}
	err = d.exec(`let doc = Automerge.init()`)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Doc) exec(js string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	// fmt.Println()
	// fmt.Println(js)
	// fmt.Println()

	_, err := d.v8.RunScript(js, "exec.js")
	if err != nil {
		return fmt.Errorf("failed to execute\n---\n%s\n---\n%w", js, err)
	}
	return nil
}

func (d *Doc) Change(js string) error {
	const head = "doc = Automerge.change(doc, doc => {\n"
	const tail = "\n})\n"
	return d.exec(head + js + tail)
}

func Do() error {
	doc, err := NewDoc()
	if err != nil {
		return err
	}
	err = doc.Change("doc.text = {}")
	if err != nil {
		return err
	}

	buf := new(strings.Builder)
	buf.WriteString("doc.text['ABC'] = new Automerge.Text()\n")
	printf := func(msg string, args ...interface{}) {
		fmt.Fprintf(buf, msg, args...)
	}
	writeString := func(s string) {
		printf("%s", s)
	}

	off := 0
	const chunkSize = 333

	for lo := 0; lo < 80000; lo += chunkSize {
		chunk := 80000 - lo
		if chunk > chunkSize {
			chunk = chunkSize
		}
		printf("doc.text['ABC'].insertAt(%d", off)
		for i := lo; i < lo+chunk; i++ {
			writeString(",'a'")
		}
		off += chunk
		writeString(")\n")
	}
	return doc.Change(buf.String())
}
