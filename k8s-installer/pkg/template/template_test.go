package template

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

const (
	Base64EncodedTemplate     = `{{b64enc .}}`
	Base64EncodedPre          = "123456"
	Base64EncodedExpected     = "MTIzNDU2"
	StrSplitTemplate          = `{{split "." .}}`
	StrSplitPre               = "foo.bar.baz"
	TestName                  = "test"
	Concurrency               = 1000
	StrArrLocateTemplate      = `{{index %d .}}`
	MultiPipelineTestTemplate = `{{split "." . | index %d}}`
)

var StrArr = []string{"foo", "bar", "baz"}

func TestSingleRender(t *testing.T) {
	renderred, err := New(TestName).Render(Base64EncodedTemplate, Base64EncodedPre)
	if err != nil {
		t.Fatalf("template render error: %v", err)
	}
	if renderred != Base64EncodedExpected {
		t.Fatalf("%s BASE64 encoded to %s, expected: %s", Base64EncodedPre, renderred,
			Base64EncodedExpected)
	}
	t.Logf("render out: %s", renderred)
}

func TestConcurrentRender(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(Concurrency)
	for i := 0; i < Concurrency; i++ {
		go func() {
			dt := New(TestName)
			renderred, err := dt.Render(Base64EncodedTemplate, Base64EncodedPre)
			if err != nil {
				t.Fatalf("template render error: %v", err)
			}
			if renderred != Base64EncodedExpected {
				t.Fatalf("%s BASE64 encoded to %s, expected: %s", Base64EncodedPre,
					renderred, Base64EncodedExpected)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestSplitPipeline(t *testing.T) {
	renderred, err := New(TestName).Render(StrSplitTemplate, StrSplitPre)
	if err != nil {
		t.Fatalf("template render error: %v", err)
	}
	t.Logf("render out: %s", renderred)
}

func TestStrArrIndexPipeline(t *testing.T) {
	for i, str := range StrArr {
		renderred, err := New(TestName).Render(fmt.Sprintf(StrArrLocateTemplate, i), StrArr)
		if err != nil {
			t.Fatalf("template render error: %v", err)
		}
		if renderred != str {
			t.Fatalf("arr[%d] supposed to be %s, now: %s", i, str, renderred)
		}
	}
}

func TestMultiPipeline(t *testing.T) {
	for i, str := range strings.Split(StrSplitPre, ".") {
		renderred, err := New(TestName).Render(fmt.Sprintf(MultiPipelineTestTemplate, i), StrSplitPre)
		if err != nil {
			t.Fatalf("template render error: %v", err)
		}
		if renderred != str {
			t.Fatalf("arr[%d] supposed to be %s, now: %s", i, str, renderred)
		}
	}
}
