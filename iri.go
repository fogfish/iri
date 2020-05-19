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
	Identity() IRI
}

/*

IRI is Internationalized Resource Identifier
https://en.wikipedia.org/wiki/Internationalized_Resource_Identifier
*/
type IRI struct {
	ID Compact `dynamodbav:"id" json:"id"`
}

/*

New parses a compact IRI string
*/
func New(iri string, args ...interface{}) IRI {
	if len(args) > 0 {
		return IRI{ID: NewCompact(fmt.Sprintf(iri, args...))}
	}

	return IRI{ID: NewCompact(iri)}
}

/*

Prefix return IRI prefix
*/
func (iri IRI) Prefix(rank ...int) string {
	return iri.ID.Prefix(rank...)
}

/*

Suffix return IRI suffix
*/
func (iri IRI) Suffix(rank ...int) string {
	return iri.ID.Suffix(rank...)
}

/*

Parent returns a IRI that is prefix of this one
*/
func (iri IRI) Parent(rank ...int) IRI {
	return IRI{ID: iri.ID.Parent(rank...)}
}

/*

Heir returns a IRI that descendant of this one.
*/
func (iri IRI) Heir(segment string) IRI {
	return IRI{ID: iri.ID.Heir(segment)}
}

/*

Path converts IRI to the path, joins IRI segments
*/
func (iri IRI) Path() string {
	return path.Join(iri.ID.Seq...)
}

/*

Compact converts IRI to the string (compact representation)
*/
func (iri IRI) Compact() Compact {
	return iri.ID
}

/*

Identity return unique identity
*/
func (iri IRI) Identity() IRI {
	return iri
}

/*

Eq return true if IRI equals
*/
func (iri IRI) Eq(x IRI) bool {
	return iri.ID.Eq(x.ID)
}

/*

Segments returns segments of IRI
*/
func (iri IRI) Segments() []string {
	return iri.ID.Segments()
}

/*

Compact is a compact (prefix:suffix) representation of IRI
*/
type Compact struct {
	Seq []string
}

/*

NewCompact builds compact IRI from string
*/
func NewCompact(iri string) Compact {
	return Compact{
		Seq: strings.Split(iri, ":"),
	}
}

/*

Prefix return IRI prefix
*/
func (iri Compact) Prefix(rank ...int) string {
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
func (iri Compact) Suffix(rank ...int) string {
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
func (iri Compact) Parent(rank ...int) Compact {
	r := 1
	if len(rank) > 0 {
		r = rank[0]
	}

	n := len(iri.Seq) - r
	if n <= 0 {
		return Compact{Seq: []string{""}}
	}

	return Compact{Seq: append([]string{}, iri.Seq[:n]...)}
}

/*

Heir returns a IRI that descendant of this one.
*/
func (iri Compact) Heir(segment string) Compact {
	if len(iri.Seq) == 1 && iri.Seq[0] == "" {
		return Compact{Seq: []string{segment}}
	}

	return Compact{Seq: append(append([]string{}, iri.Seq...), segment)}
}

/*

String ...
*/
func (iri Compact) String() string {
	return strings.Join(iri.Seq, ":")
}

/*

Eq return true if two IRI equals
*/
func (iri Compact) Eq(x Compact) bool {
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
func (iri Compact) Segments() []string {
	return iri.Seq
}

/*

MarshalJSON `IRI ⟼ "prefix:suffix"`
*/
func (iri Compact) MarshalJSON() ([]byte, error) {
	if len(iri.Seq) == 0 {
		return json.Marshal("")
	}

	return json.Marshal(iri.String())
}

/*

UnmarshalJSON `"prefix:suffix" ⟼ IRI`
*/
func (iri *Compact) UnmarshalJSON(b []byte) error {
	var path string
	err := json.Unmarshal(b, &path)
	if err != nil {
		return err
	}

	*iri = New(path).ID
	return nil
}

/*

MarshalDynamoDBAttributeValue `IRI ⟼ "prefix/suffix"`
*/
func (iri Compact) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
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
func (iri *Compact) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	*iri = NewCompact(aws.StringValue(av.S))
	return nil
}
