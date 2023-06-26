package cborquery

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/antchfx/xpath"
	"github.com/fxamacker/cbor"
	"github.com/stretchr/testify/require"
)

type keyValueList []keyValue

func BenchmarkSelectorCache(b *testing.B) {
	_, err := getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	require.NoError(b, err)

	DisableSelectorCache = false
	for i := 0; i < b.N; i++ {
		_, _ = getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	}
}

func BenchmarkDisableSelectorCache(b *testing.B) {
	_, err := getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	require.NoError(b, err)

	DisableSelectorCache = true
	for i := 0; i < b.N; i++ {
		_, _ = getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	}
}

func TestNavigator(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)

	nav := CreateXPathNavigator(doc)
	nav.MoveToRoot()
	require.Equal(t, nav.NodeType(), xpath.RootNode, "node type is not RootNode")

	expectedPeople := []keyValueList{
		[]keyValue{
			{key: "age", value: "42"},
			{key: "email", value: "john@example.com"},
			{key: "id", value: "101"},
			{key: "name", value: "John Doe"},
		},
		[]keyValue{
			{key: "age", value: "40"},
			{key: "id", value: "102"},
			{key: "name", value: "Jane Doe"},
		},
	}
	require.True(t, nav.MoveToChild())
	for _, keyvalues := range expectedPeople {
		require.Equal(t, "people", nav.Current().Name)
		require.True(t, nav.MoveToChild())

		for i, v := range keyvalues {
			if i > 0 {
				require.True(t, nav.MoveToNext())
			}
			require.Equal(t, v.key, nav.Current().Name)
			require.Equal(t, v.value, nav.Value())
		}
		require.True(t, nav.MoveToParent())
		require.True(t, nav.MoveToNext())
	}

	require.True(t, nav.MoveToParent())
	require.Equal(t, nav.Current().Type, DocumentNode, "expected 'DocumentNode'")
}

func TestQueryNames(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)

	expected := []string{
		"John Doe",
		"Jane Doe",
		"Jack Doe",
		"Jack Buck",
		"Janet Doe",
	}

	nodes, err := QueryAll(doc, "/descendant::*[name() = 'people']/name")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "name", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}

	nodes, err = QueryAll(doc, "//name")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "name", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}

	nodes, err = QueryAll(doc, "/people[3]/name")
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	require.Equal(t, "name", nodes[0].Name)
	require.EqualValues(t, expected[2], nodes[0].Value())
}

func TestQueryPhoneNumberFirst(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)

	expected := []string{
		"555-555-5555",
		"555-555-0000",
		"555-777-0000",
	}
	nodes, err := QueryAll(doc, "//phones[1]/number")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "number", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}
}

func TestQueryPhoneNumberLast(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)

	expected := []string{
		"555-555-5555",
		"555-555-0002",
		"555-777-0001",
	}
	nodes, err := QueryAll(doc, "//phones[last()]/number")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "number", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}
}

func TestQueryPhoneNoEmail(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)

	nodes, err := QueryAll(doc, "/people[not(email)]/id")
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	require.Equal(t, "id", nodes[0].Name)
	require.EqualValues(t, 102, nodes[0].Value())
}

func TestQueryPhoneAge(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)

	expected := []string{
		"John Doe",
		"Jane Doe",
		"Jack Buck",
	}
	nodes, err := QueryAll(doc, "/people[age > 18]/name")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))
	for i, name := range expected {
		require.Equal(t, "name", nodes[i].Name)
		require.EqualValues(t, name, nodes[i].Value())
	}
}

func TestQueryJack(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)

	nodes, err := QueryAll(doc, "//people[contains(name, 'Jack')]/id")
	require.NoError(t, err)

	require.Len(t, nodes, 2)
	require.Equal(t, "id", nodes[0].Name)
	require.EqualValues(t, 201, nodes[0].Value())
	require.Equal(t, "id", nodes[1].Name)
	require.EqualValues(t, 301, nodes[1].Value())
}

func TestQueryExample(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)

	expected := []string{
		"John Doe: ",
		"Jane Doe: ",
		"Jack Doe: 555-555-5555",
		"Jack Buck: 555-555-0000,555-555-0001,555-555-0002",
		"Janet Doe: 555-777-0000,555-777-0001",
	}

	nodes, err := QueryAll(doc, "//people")
	require.NoError(t, err)
	require.Len(t, nodes, len(expected))

	for i, person := range nodes {
		name := FindOne(person, "name").InnerText()
		numbers := make([]string, 0)
		for _, node := range Find(person, "phones/number") {
			numbers = append(numbers, node.InnerText())
		}
		v := fmt.Sprintf("%s: %s", name, strings.Join(numbers, ","))
		require.EqualValues(t, expected[i], v)
	}
}

func TestQueryNode(t *testing.T) {
	test := []map[interface{}]interface{}{
		{
			258: "002-2.1.x",
			259: "14ca85ed9",
			260: 1687787189711304960,
			261: true,
			263: 3,
			264: 23.76,
			265: 68.934,
			266: false,
		},
	}
	msg, err := cbor.Marshal(test, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)
	selected, err := QueryAll(doc, "//element")
	require.NoError(t, err)
	require.Len(t, selected, 1)
	nodes, err := QueryAll(selected[0], "n260")
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	require.Equal(t, uint64(1687787189711304960), nodes[0].value)
}
