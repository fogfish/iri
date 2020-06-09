package iri

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Thing is "interface tag" allows usage of IRI abstraction in other interfaces
type Thing interface {
	Identity() ID
}

/*

ID is type tagging built over IRI type, unique identity of a thing.

  type MyStruct struct {
		iri.ID
	}
*/
type ID struct {
	IRI IRI `dynamodbav:"id" json:"id"`
}

/*

New parses a compact IRI string
*/
func New(iri string, args ...interface{}) ID {
	if len(args) > 0 {
		return ID{IRI: NewIRI(fmt.Sprintf(iri, args...))}
	}

	return ID{IRI: NewIRI(iri)}
}

/*

Prefix return IRI prefix
*/
func (iri ID) Prefix(rank ...int) string {
	return iri.IRI.Prefix(rank...)
}

/*

Suffix return IRI suffix
*/
func (iri ID) Suffix(rank ...int) string {
	return iri.IRI.Suffix(rank...)
}

/*

Parent returns a IRI that is prefix of this one
*/
func (iri ID) Parent(rank ...int) ID {
	return ID{IRI: iri.IRI.Parent(rank...)}
}

/*

Heir returns a IRI that descendant of this one.
*/
func (iri ID) Heir(segment string) ID {
	return ID{IRI: iri.IRI.Heir(segment)}
}

/*

Path converts IRI to the path, joins IRI segments
*/
func (iri ID) Path() string {
	return path.Join(iri.IRI.Seq...)
}

/*

ToIRI converts ID to IRI type
*/
func (iri ID) ToIRI() *IRI {
	return &iri.IRI
}

/*

Identity return unique identity, required by Thing interface
*/
func (iri ID) Identity() ID {
	return iri
}

/*

Eq return true if IRI equals
*/
func (iri ID) Eq(x ID) bool {
	return iri.IRI.Eq(x.IRI)
}

/*

Segments returns segments of IRI
*/
func (iri ID) Segments() []string {
	return iri.IRI.Segments()
}

/*

IRI is Internationalized Resource Identifier
https://en.wikipedia.org/wiki/Internationalized_Resource_Identifier
*/
type IRI struct {
	Seq []string
}

/*

NewIRI builds compact IRI from string
*/
func NewIRI(iri string) IRI {
	return IRI{
		Seq: strings.Split(iri, ":"),
	}
}

/*

Prefix return IRI prefix
*/
func (iri IRI) Prefix(rank ...int) string {
	r := 1
	if len(rank) > 0 {
		r = rank[0]
	}

	if r == 1 && len(iri.Seq) == 1 {
		return strings.Join(iri.Seq, ":")
	}

	n := len(iri.Seq) - r
	if n < 0 {
		return ""
	}

	return strings.Join(iri.Seq[:n], ":")
}

/*

Suffix return IRI suffix
*/
func (iri IRI) Suffix(rank ...int) string {
	r := 1
	if len(rank) > 0 {
		r = rank[0]
	}

	if len(iri.Seq) == 1 {
		return ""
	}

	n := len(iri.Seq) - r
	if n < 0 {
		n = 0
	}

	return strings.Join(iri.Seq[n:len(iri.Seq)], ":")
}

/*

Parent returns a IRI that is prefix of this one
*/
func (iri IRI) Parent(rank ...int) IRI {
	r := 1
	if len(rank) > 0 {
		r = rank[0]
	}

	n := len(iri.Seq) - r
	if n <= 0 {
		return IRI{Seq: []string{""}}
	}

	return IRI{Seq: append([]string{}, iri.Seq[:n]...)}
}

/*

Heir returns a IRI that descendant of this one.
*/
func (iri IRI) Heir(segment string) IRI {
	if len(iri.Seq) == 1 && iri.Seq[0] == "" {
		return IRI{Seq: []string{segment}}
	}

	return IRI{Seq: append(append([]string{}, iri.Seq...), segment)}
}

/*

String ...
*/
func (iri IRI) String() string {
	return strings.Join(iri.Seq, ":")
}

/*

Eq return true if two IRI equals
*/
func (iri IRI) Eq(x IRI) bool {
	if len(iri.Seq) != len(x.Seq) {
		return false
	}

	for i, v := range iri.Seq {
		if x.Seq[i] != v {
			return false
		}
	}

	return true
}

/*

Segments return elements
*/
func (iri IRI) Segments() []string {
	return iri.Seq
}

/*

MarshalJSON `IRI ⟼ "prefix:suffix"`
*/
func (iri IRI) MarshalJSON() ([]byte, error) {
	if len(iri.Seq) == 0 {
		return json.Marshal("")
	}

	return json.Marshal(iri.String())
}

/*

UnmarshalJSON `"prefix:suffix" ⟼ IRI`
*/
func (iri *IRI) UnmarshalJSON(b []byte) error {
	var path string
	err := json.Unmarshal(b, &path)
	if err != nil {
		return err
	}

	*iri = New(path).IRI
	return nil
}

/*

MarshalDynamoDBAttributeValue `IRI ⟼ "prefix/suffix"`
*/
func (iri IRI) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	if len(iri.Seq) == 0 {
		av.NULL = aws.Bool(true)
		return nil
	}

	// Note: we are using string representation to allow linked data in dynamo tables
	val, err := dynamodbattribute.Marshal(iri.String())
	if err != nil {
		return err
	}

	av.S = val.S
	return nil
}

/*

UnmarshalDynamoDBAttributeValue `"prefix/suffix" ⟼ IRI`
*/
func (iri *IRI) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	*iri = NewIRI(aws.StringValue(av.S))
	return nil
}
