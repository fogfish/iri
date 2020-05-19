package iri_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/fogfish/iri"
	"github.com/fogfish/it"
)

var (
	r0 iri.IRI = iri.New("")
	r1 iri.IRI = iri.New("a")
	r2 iri.IRI = iri.New("a:b")
	r3 iri.IRI = iri.New("a:b:c")
	r4 iri.IRI = iri.New("a:b:c:d")
	r5 iri.IRI = iri.New("a:b:c:d:e")
)

func TestIRI(t *testing.T) {
	test := map[*iri.IRI][]string{
		&r0: []string{""},
		&r1: []string{"a"},
		&r2: []string{"a", "b"},
		&r3: []string{"a", "b", "c"},
		&r4: []string{"a", "b", "c", "d"},
		&r5: []string{"a", "b", "c", "d", "e"},
	}

	for k, v := range test {
		it.Ok(t).
			If(*k).Should().Equal(iri.IRI{iri.Compact{v}}).
			If(k.Segments()).Should().Equal(v)
	}
}

func TestPrefix(t *testing.T) {
	test := map[*iri.IRI][]string{
		&r0: []string{"", "", "", "", "", ""},
		&r1: []string{"a", "a", "", "", "", ""},
		&r2: []string{"a", "a", "", "", "", ""},
		&r3: []string{"a:b", "a:b", "a", "", "", ""},
		&r4: []string{"a:b:c", "a:b:c", "a:b", "a", "", ""},
		&r5: []string{"a:b:c:d", "a:b:c:d", "a:b:c", "a:b", "a", ""},
	}

	for k, v := range test {
		it.Ok(t).
			If(k.Prefix()).Should().Equal(v[0]).
			If(k.Prefix(1)).Should().Equal(v[1]).
			If(k.Prefix(2)).Should().Equal(v[2]).
			If(k.Prefix(3)).Should().Equal(v[3]).
			If(k.Prefix(4)).Should().Equal(v[4]).
			If(k.Prefix(5)).Should().Equal(v[5])
	}
}

func TestSuffix(t *testing.T) {
	test := map[*iri.IRI][]string{
		&r0: []string{"", "", "", "", "", ""},
		&r1: []string{"", "", "", "", "", ""},
		&r2: []string{"b", "b", "a:b", "a:b", "a:b", "a:b"},
		&r3: []string{"c", "c", "b:c", "a:b:c", "a:b:c", "a:b:c"},
		&r4: []string{"d", "d", "c:d", "b:c:d", "a:b:c:d", "a:b:c:d"},
		&r5: []string{"e", "e", "d:e", "c:d:e", "b:c:d:e", "a:b:c:d:e"},
	}

	for k, v := range test {
		it.Ok(t).
			If(k.Suffix()).Should().Equal(v[0]).
			If(k.Suffix(1)).Should().Equal(v[1]).
			If(k.Suffix(2)).Should().Equal(v[2]).
			If(k.Suffix(3)).Should().Equal(v[3]).
			If(k.Suffix(4)).Should().Equal(v[4]).
			If(k.Suffix(5)).Should().Equal(v[5])
	}
}

func TestParent(t *testing.T) {
	test := map[*iri.IRI][]iri.IRI{
		&r0: []iri.IRI{r0, r0, r0, r0, r0, r0},
		&r1: []iri.IRI{r0, r0, r0, r0, r0, r0},
		&r2: []iri.IRI{r1, r1, r0, r0, r0, r0},
		&r3: []iri.IRI{r2, r2, r1, r0, r0, r0},
		&r4: []iri.IRI{r3, r3, r2, r1, r0, r0},
		&r5: []iri.IRI{r4, r4, r3, r2, r1, r0},
	}

	for k, v := range test {
		it.Ok(t).
			If(k.Parent()).Should().Equal(v[0]).
			If(k.Parent(1)).Should().Equal(v[1]).
			If(k.Parent(2)).Should().Equal(v[2]).
			If(k.Parent(3)).Should().Equal(v[3]).
			If(k.Parent(4)).Should().Equal(v[4]).
			If(k.Parent(5)).Should().Equal(v[5])
	}
}

func TestHeir(t *testing.T) {
	it.Ok(t).
		If(r0.Heir("a")).Should().Equal(r1).
		If(r1.Heir("b")).Should().Equal(r2).
		If(r2.Heir("c")).Should().Equal(r3).
		If(r3.Heir("d")).Should().Equal(r4).
		If(r4.Heir("e")).Should().Equal(r5)
}

func TestPath(t *testing.T) {
	test := map[*iri.IRI]string{
		&r0: "",
		&r1: "a",
		&r2: "a/b",
		&r3: "a/b/c",
		&r4: "a/b/c/d",
		&r5: "a/b/c/d/e",
	}

	for k, v := range test {
		it.Ok(t).
			If(k.Path()).Should().Equal(v)
	}
}

func TestImmutable(t *testing.T) {
	rN := r3.Parent().Heir("t")

	it.Ok(t).
		If(r3.Path()).Should().Equal("a/b/c").
		If(rN.Path()).Should().Equal("a/b/t")
}

func TestEq(t *testing.T) {
	test := []iri.IRI{r0, r1, r2, r3, r4, r5}

	for _, v := range test {
		it.Ok(t).If(v.Eq(v)).Should().Equal(true)
	}
}

func TestNotEq(t *testing.T) {
	r6 := iri.New("1:2:3:4:5:6")
	test := []iri.IRI{r0, r1, r2, r3, r4, r5}

	for _, v := range test {
		it.Ok(t).
			If(v.Eq(r6)).Should().Equal(false).
			If(v.Eq(v.Parent().Heir("t"))).Should().Equal(false)
	}
}

func TestJSON(t *testing.T) {
	type Struct struct {
		iri.IRI
		Title string `json:"title"`
	}

	test := map[*Struct]string{
		&Struct{IRI: iri.New(""), Title: "t"}:      "{\"id\":\"\",\"title\":\"t\"}",
		&Struct{IRI: iri.New("a"), Title: "t"}:     "{\"id\":\"a\",\"title\":\"t\"}",
		&Struct{IRI: iri.New("a:b"), Title: "t"}:   "{\"id\":\"a:b\",\"title\":\"t\"}",
		&Struct{IRI: iri.New("a:b:c"), Title: "t"}: "{\"id\":\"a:b:c\",\"title\":\"t\"}",
	}

	for eg, expect := range test {
		in := Struct{}

		bytes, err1 := json.Marshal(eg)
		err2 := json.Unmarshal(bytes, &in)

		it.Ok(t).
			If(err1).Should().Equal(nil).
			If(err2).Should().Equal(nil).
			If(*eg).Should().Equal(in).
			If(string(bytes)).Should().Equal(expect)
	}
}

func TestDynamo(t *testing.T) {
	type Struct struct {
		iri.IRI
		Title string `dynamodbav:"title"`
	}

	test := []Struct{
		Struct{IRI: iri.New(""), Title: "t"},
		Struct{IRI: iri.New("a"), Title: "t"},
		Struct{IRI: iri.New("a:b"), Title: "t"},
		Struct{IRI: iri.New("a:b:c"), Title: "t"},
	}

	for _, eg := range test {
		in := Struct{}

		gen, err1 := dynamodbattribute.MarshalMap(eg)
		err2 := dynamodbattribute.UnmarshalMap(gen, &in)

		it.Ok(t).
			If(err1).Should().Equal(nil).
			If(err2).Should().Equal(nil).
			If(eg).Should().Equal(in)
	}
}

func TestTypeSafe(t *testing.T) {
	type A struct{ iri.IRI }
	type B struct{ iri.IRI }
	type C struct{ iri.IRI }

	a := A{iri.New("a")}
	b := B{iri.New("a:b")}
	c := C{b.Heir("c")}

	it.Ok(t).
		If(a.IRI).Should().Equal(r1).
		If(b.IRI).Should().Equal(r2).
		If(c.IRI).Should().Equal(r3)
}
