package lex

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

const (
	picPrefix = "PIC "
)

type parser func(t *Tree, l line, root *Record) *Record

func noOp(_ *Tree, _ line, _ *Record) *Record {
	return nil
}

// parsePIC is a parser that is used to build records
// for lines that are determined to be PIC definitions
func parsePIC(_ *Tree, l line, _ *Record) *Record {
	picNumDef := strings.TrimPrefix(l.items[6].val, picPrefix)
	length, err := parsePICCount(picNumDef)
	if err != nil {
		log.Fatalln(err)
	}

	return &Record{
		Name:   l.items[4].val,
		Length: length,
		depth:  l.items[2].val,
		Typ:    parsePICType(picNumDef),
	}
}

// parseRedefines is a parser that is used to build records
// for lines that are determined to be REDEFINES definitions
//
// It will build the new Record and replace the the redefinition
// target
func parseRedefines(_ *Tree, l line, root *Record) *Record {
	picNumDef := strings.TrimPrefix(l.items[10].val, picPrefix)
	length, err := parsePICCount(picNumDef)
	if err != nil {
		log.Fatalln(err)
	}

	r := &Record{
		Name:   l.items[4].val,
		Length: length,
		Typ:    parsePICType(picNumDef),
	}

	target := l.items[8].val
	dst, i := root.fromCache(target)
	if dst == nil {
		log.Fatalln(fmt.Sprintf("redefinition target %s does not exist", target))
	}

	return root.redefine(i, dst, r)
}

// parseGroupRedefines is a parser that is used to build records
// for lines that are determined to be REDEFINES definitions
// for groups, instead of PICs
//
// It will build the new Record and replace the the redefinition
// target
func parseGroupRedefines(t *Tree, l line, root *Record) *Record {
	target := strings.TrimSuffix(l.items[8].val, ".")
	dst, _ := root.fromCache(target)
	if dst == nil {
		log.Fatalln(fmt.Sprintf("redefinition target %s does not exist", target))
	}

	if dst.depthMap == nil {
		// then check whether a node has children at this depth
		parent, seenGroup := root.depthMap[dst.depth]
		if seenGroup {
			copyDepthMap(parent, dst)
			root = parent
		}
	}

	r := &Record{
		Name:     l.items[4].val,
		Typ:      reflect.Struct,
		depth:    l.items[2].val,
		depthMap: dst.depthMap,
	}

	dst, i := root.fromCache(target)
	if dst == nil {
		log.Fatalln(fmt.Sprintf("redefinition target %s does not exist", target))
	}
	root.Length -= dst.Length

	r = root.redefine(i, dst, r)
	t.parseLines(r)
	root.Length += r.Length
	return r
}

func parseOccurs(_ *Tree, l line, _ *Record) *Record {
	picNumDef := strings.TrimPrefix(strings.TrimSpace(l.items[6].val), picPrefix)
	length, err := parsePICCount(picNumDef)
	if err != nil {
		log.Fatalln(err)
	}

	n, err := parseOccursCount(l.items[8])
	if err != nil {
		log.Fatalln(err)
	}

	return &Record{
		Name:   l.items[4].val,
		Length: length,
		Occurs: n,
		depth:  l.items[2].val,
		Typ:    parsePICType(picNumDef),
	}
}

// TODO: (pgmitche) tweak godoc here
// isMultilineRedefinition is a fingerprinting function that validates whether a
// line is an indicator for, and sibling of a multi-line redefinition
//
// it first checks the fingerprint of the line against the first fingerprint of
// a multi-line REDEFINES definition
// then uses nx() to get the next line, and validate that against the second
// fingerprint of a multi-line REDEFINES definition
//
// when this is successful, the parseRedefines parserfn is returned, along with
// a new, single, line object built from the two fingerprinted line objects.
func parseRedefinesMulti(t *Tree, l line, _ *Record) *Record {
	if err := t.nextLine(); err != nil {
		panic(err)
	}

	_, i, ok := basicParserGet(t.line.items)
	if !ok || !equalFingerprints(getFingerprint(i), multiRedefinesPartFp) {
		log.Fatalln("parser indicated multi-line redefinition, but failed to verify next line")
	}

	l.items = lineFromMultiRedefines(l.items, i)
	nl := l

	return parseRedefines(nil, nl, nil)
}

// TODO: (pgmitche) tweak godoc here
// isMultilineOccurrence is a fingerprinting function that validates whether a
// line is an indicator for, and sibling of a multi-line occurrence definition
//
// it first checks the fingerprint of the line against the first fingerprint of
// a multi-line OCCURS definition
// then uses nx() to get the next line, and validate that against the second
// fingerprint of a multi-line OCCURS definition
//
// when this is successful, the parseOccurs parserfn is returned, along with a
// new, single, line object built from the two fingerprinted line objects.
func parseOccursMulti(t *Tree, l line, _ *Record) *Record {
	if err := t.nextLine(); err != nil {
		panic(err)
	}

	_, i, ok := basicParserGet(t.line.items)
	if !ok || !equalFingerprints(getFingerprint(i), multiOccursPartFp) {
		log.Fatalln()
	}

	l.items = lineFromMultiOccurs(l.items, i)
	nl := l

	return parseOccurs(nil, nl, nil)
}

// parseNumDelimitedStruct is a parser wrapper used to build records
// for lines that are determined to be new group definitions that are built with
// number delimiter tokens at the start and end of the source line
//
// It will call parseStruct to handle logic for new groups
func parseNumDelimitedStruct(t *Tree, l line, root *Record) *Record {
	// if the level number is 01, ignore this object.
	// refer to README.md Level Number section
	if l.items[2].val == recordDescriptionIndicator {
		return noOp(t, l, root)
	}

	return parseStruct(t, l, root, 4, 2)
}

// parseNonNumDelimitedStruct is a parser wrapper used to build records
// for lines that are determined to be new group definitions that are built
// without number delimiter tokens at the start and end of the source line
//
// It will call parseStruct to handle logic for new groups
func parseNonNumDelimitedStruct(t *Tree, l line, root *Record) *Record {
	// if the level number is 01, ignore this object.
	// refer to README.md Level Number section
	if l.items[1].val == recordDescriptionIndicator {
		return noOp(t, l, root)
	}

	return parseStruct(t, l, root, 3, 1)
}

// parseStruct will build a new Record of type struct, store itself under the
// parent Record as a child. It will also add an entry to the parent struct
// indicating a new group.
//
// new groups are determined by a number token preceding an identifier token
// if this number has been seen before, then this struct is not a new group, but
// a sibling of other structs/pics in the root Record
// if this number is a new unseen value, parsing will continue for each line
// but each new line will treat this struct as it's root.
//
// In this way we may build a tree like
// root
//  |-group1
//  |   |-picA
//  |-group2
//  |	|-picA
func parseStruct(t *Tree, l line, root *Record, nameIdx, groupIdx int) *Record {
	newNode := &Record{
		Name:  l.items[nameIdx].val,
		Typ:   reflect.Struct,
		depth: l.items[groupIdx].val,
	}

	return delve(t, root, newNode)
}

func delve(t *Tree, root *Record, newRecord *Record) *Record {
	parent, seenGroup := root.depthMap[newRecord.depth]
	if seenGroup {
		parent.Children = append(parent.Children, newRecord)
		t.parseLines(newRecord)
		l := newRecord.Length
		if newRecord.Occurs > 0 {
			l *= newRecord.Occurs
		}

		parent.Length += l
		// then the parent is the target for continued addition of children
		return parent
	}

	root.depthMap[newRecord.depth] = root
	copyDepthMap(root, newRecord)

	root.Children = append(root.Children, newRecord)
	t.parseLines(newRecord)

	l := newRecord.Length
	if newRecord.Occurs > 0 {
		l *= newRecord.Occurs
	}

	root.Length += l
	// else the newRecord is the target for continued addition of children
	return newRecord
}

func copyDepthMap(src, dst *Record) {
	if dst.depthMap == nil {
		dst.depthMap = make(map[string]*Record)
	}

	for key, value := range src.depthMap {
		dst.depthMap[key] = value
	}
}
