package cborquery

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor"
	"github.com/stretchr/testify/require"

	"github.com/srebhan/cborquery/testcases/addressbook"
)

var addressbookSample = &addressbook.AddressBook{
	People: []*addressbook.Person{
		{
			Name:  "John Doe",
			Id:    101,
			Email: "john@example.com",
			Age:   42,
		},
		{
			Name: "Jane Doe",
			Id:   102,
			Age:  40,
		},
		{
			Name:  "Jack Doe",
			Id:    201,
			Email: "jack@example.com",
			Age:   12,
			Phones: []*addressbook.Person_PhoneNumber{
				{Number: "555-555-5555", Type: addressbook.Person_WORK},
			},
		},
		{
			Name:  "Jack Buck",
			Id:    301,
			Email: "buck@example.com",
			Age:   19,
			Phones: []*addressbook.Person_PhoneNumber{
				{Number: "555-555-0000", Type: addressbook.Person_HOME},
				{Number: "555-555-0001", Type: addressbook.Person_MOBILE},
				{Number: "555-555-0002", Type: addressbook.Person_WORK},
			},
		},
		{
			Name:  "Janet Doe",
			Id:    1001,
			Email: "janet@example.com",
			Age:   16,
			Phones: []*addressbook.Person_PhoneNumber{
				{Number: "555-777-0000"},
				{Number: "555-777-0001", Type: addressbook.Person_HOME},
			},
		},
	},
	Tags: []string{"home", "private", "friends"},
}

func TestParseAddressBookXML(t *testing.T) {
	msg, err := cbor.Marshal(addressbookSample, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)
	require.Len(t, doc.ChildNodes(), 8)

	actual := doc.OutputXML()
	expected := `<?xml version="1.0"?><root><people><age>42</age><email>john@example.com</email><id>101</id><name>John Doe</name></people><people><age>40</age><id>102</id><name>Jane Doe</name></people><people><age>12</age><email>jack@example.com</email><id>201</id><name>Jack Doe</name><phones><number>555-555-5555</number><type>2</type></phones></people><people><age>19</age><email>buck@example.com</email><id>301</id><name>Jack Buck</name><phones><number>555-555-0000</number><type>1</type></phones><phones><number>555-555-0001</number></phones><phones><number>555-555-0002</number><type>2</type></phones></people><people><age>16</age><email>janet@example.com</email><id>1001</id><name>Janet Doe</name><phones><number>555-777-0000</number></phones><phones><number>555-777-0001</number><type>1</type></phones></people><tags>home</tags><tags>private</tags><tags>friends</tags></root>`
	require.Equal(t, expected, actual)
}

func TestNumericKeys(t *testing.T) {
	test := map[interface{}]interface{}{
		1:      "foo",
		2:      true,
		3.14:   42.3,
		"test": 23,
	}
	msg, err := cbor.Marshal(test, cbor.EncOptions{})
	require.NoError(t, err)

	doc, err := Parse(bytes.NewBuffer(msg))
	require.NoError(t, err)
	require.Len(t, doc.ChildNodes(), 4)

	expected := []keyValue{
		{"1", "foo"},
		{"2", true},
		{"3.14", float64(42.3)},
		{"test", uint64(23)},
	}
	actual := make([]keyValue, 0, len(doc.ChildNodes()))
	for _, n := range doc.ChildNodes() {
		actual = append(actual, keyValue{n.Name, n.Value()})
	}
	require.ElementsMatch(t, actual, expected)
}
