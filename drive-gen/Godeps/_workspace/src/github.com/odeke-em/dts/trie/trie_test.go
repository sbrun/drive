package trie

import (
	"fmt"
	"reflect"
	"testing"
)

func TestInit(t *testing.T) {
	tr := New(AsciiAlphabet)
	if tr == nil {
		t.Errorf("Expecting non nil creation")
	}
}

func TestGetAndAdd(t *testing.T) {
	tr := New(AsciiAlphabet)
	value, ok := tr.Get("flux")
	if ok {
		t.Errorf("Not expecting a successful retrieval from this empty map")
	}
	if value != nil {
		t.Errorf("Expecting a nil value as the retrieved result!")
	}

	key, saved := "こんにいちは", "konichiwa"
	evicted := tr.Set(key, saved)
	if evicted != nil {
		t.Errorf("Evicted value should be nil since nothing was set in there before!")
	}

	retr, rOk := tr.Get(key)
	if !rOk {
		t.Errorf("A successful retrieval expected!")
	}
	if retr == nil {
		t.Errorf("Retrieved value cannot be nil!")
	}
	if retr != saved {
		t.Errorf("Expected %v got %v", saved, retr)
	}

	popd, pOk := tr.Pop(key)
	if !pOk {
		t.Errorf("Expected a successful pop!")
	}
	if popd != saved {
		t.Errorf("Pop: Expected %v instead got %v", saved, popd)
	}

	retr2, retr2Ok := tr.Get(key)
	if retr2Ok {
		t.Errorf("Already popped that key was not expecting a successful retrieval!")
	}
	if retr2 != nil {
		t.Errorf("Expected nil, instead got %v", retr2)
	}

	setValue := "кофе"
	evicted1 := tr.Set(key, setValue)
	if evicted1 != nil {
		t.Errorf("Expected no eviction instead got %v", evicted1)
	}

	retr3, retr3Ok := tr.Get(key)
	if !retr3Ok {
		t.Errorf("Expected a successful retrieval!")
	}
	if retr3 == nil {
		t.Errorf("Expected %v instead got nil", setValue)
	}
	if retr3 != setValue {
		t.Errorf("Expected %v instead got %s", setValue, retr3)
	}
}

func TestExpectedWalkOrder(t *testing.T) {
	contentMap := map[string]interface{}{
		"mnt":   t.Errorf,
		"ghost": New,
		"break": TestInit,
	}

	tr := New(AsciiAlphabet)
	for key, value := range contentMap {
		evicted := tr.Set(key, value)
		if evicted != nil {
			t.Errorf("Clash detected after inserting %v, expected nil as eviction result, got %v",
				value, evicted)
		}
	}
	resultsChan := tr.Walk()
	results := make([]interface{}, 0)
	for res := range resultsChan {
		results = append(results, res)
	}
	resLen, expectedLen := len(results), len(contentMap)
	if resLen != expectedLen {
		t.Errorf("Expected full traversal of the map, hence %d values expected got %d",
			resLen, expectedLen)
	}

	ptr := func(v interface{}) uintptr {
		return reflect.ValueOf(v).Pointer()
	}
	ptrEqual := func(a, b interface{}) bool {
		return ptr(a) == ptr(b)
	}

	// Checking the ordering manually! TODO: Organize the expectations and value as kv struct list
	// ie break, ghost, mnt
	if !ptrEqual(results[2], t.Errorf) {
		t.Errorf("Expected %v got %v", t.Errorf, results[2])
	}
	if !ptrEqual(results[1], New) {
		t.Errorf("Expected %v got %v", New, results[1])
	}
	if !ptrEqual(results[0], TestInit) {
		t.Errorf("Expected %v got %v", TestInit, results[0])
	}
}

func TestTagging(t *testing.T) {
	targets := []string{
		"/mnt/", "/mnt/ch/gm", "/mnt/ch/px",
		"/usr/bin", "/usr/lib", "/etc/ssh/",
		"/etc/ssl/certs/own", "/usr/sbin", "/usr/exec",
	}
	tr := New(AsciiAlphabet)
	for _, path := range targets {
		tr.Set(path, path)
	}

	dir := "dir"

	divergentPaths := tr.Tag(PotentialDir, dir)
	if divergentPaths != 4 {
		t.Errorf("Expected 4 divergent paths instead got: %d", divergentPaths)
	}

	matchesChan := tr.Match(HasEOS)
	for match := range matchesChan {
		fmt.Println(match.Tag, match.Data)
	}

	markedDir := func(tn *TrieNode) bool {
		if tn == nil {
			return false
		}
		cast, ok := tn.Tag.(string)
		return ok && cast == dir
	}

	dsp := tr.Match(markedDir)
	fmt.Println("dsp", len(dsp))

	extracts := tr.MatchAndHarvest(PotentialDir)
	for extract := range extracts {
		fmt.Println("extracted", extract.Data)
	}
}

func TestBreadthFirstApply(t *testing.T) {
	tr := New(AsciiAlphabet)
	targets := []string{
		"/mnt", "/opt/X11/bin/xlogo", "/opt/X11/include/GL", "/mnt/ch/gm",
		"/mnt/ch/px", "/usr/bin", "/usr/lib", "/usr/sbin", "/usr/exec",
	}

	for _, path := range targets {
		tr.Set(path, path)
	}

	f := func(d interface{}) {
		fmt.Println("d", d)
	}

	tr.BreadthFirstApply(f)
}
