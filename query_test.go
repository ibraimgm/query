package query_test

import (
	"testing"

	"github.com/ibraimgm/query"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		original string
		values   []interface{}
		expected string
	}{
		{name: "Single substitution", original: "SELECT 1 FROM foo WHERE id=?", values: []interface{}{1}, expected: "SELECT 1 FROM foo WHERE id=$1"},
		{name: "Multiple substitution", original: "SELECT 1 FROM foo WHERE id=? AND status=?", values: []interface{}{1, 2}, expected: "SELECT 1 FROM foo WHERE id=$1 AND status=$2"},
		{name: "Missing arg", original: "SELECT 1 FROM foo WHERE id=? AND status=? AND type=?", values: []interface{}{1, 2}, expected: "SELECT 1 FROM foo WHERE id=$1 AND status=$2 AND type=?"},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			var b query.Builder

			b.Add(test.original, test.values...)
			actual := b.String()

			if actual != test.expected {
				t.Fatalf("expected '%s', but got '%s'", test.expected, actual)
			}

			if len(b.Params()) != len(test.values) {
				t.Fatalf("expected %d parameters, but found %d", len(test.values), len(b.Params()))
			}
		})
	}
}

func TestAddIf(t *testing.T) {
	value := "foo"
	var ptr *string = nil

	tests := []struct {
		name     string
		original string
		value    interface{}
		expected string
	}{
		{name: "With value", original: "SELECT ?", value: 1, expected: "SELECT $1"},
		{name: "Nil value", original: "SELECT ?", value: nil, expected: ""},
		{name: "Pointer with value", original: "SELECT ?", value: &value, expected: "SELECT $1"},
		{name: "Pointer to nil", original: "SELECT ?", value: ptr, expected: ""},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			var b query.Builder

			b.AddIf(test.original, test.value)
			actual := b.String()

			if actual != test.expected {
				t.Fatalf("expected '%s', but got '%s'", test.expected, actual)
			}
		})
	}
}

func TestMultipleBuffers(t *testing.T) {
	const expected = "SELECT id,name,age,dept FROM employees WHERE 1=1 AND dept = $1 AND age > $2 ORDER BY id"

	var b query.Builder
	b.Add("SELECT id,name,age,dept")
	b.From(" FROM employees ")
	b.Where("WHERE 1=1")
	b.WhereIf(" AND dept = ?", "HR")
	b.WhereIf(" AND name = ?", nil)
	b.WhereIf(" AND age > ?", 30)
	b.Order(" ORDER BY id")

	actual := b.String()
	if actual != expected {
		t.Fatalf("expected '%s', but got '%s'", expected, actual)
	}
}

func TestSetParamReplace(t *testing.T) {
	const originalParam int = 10
	const newParam int = 20

	var b query.Builder
	b.Add("SELECT 1 FROM t WHERE id = ?", originalParam)

	// make sure the original parameter is ok
	params := b.Params()
	if len(params) != 1 {
		t.Fatalf("wrong parameter lenght: %d", len(params))
	}

	if v, ok := params[0].(int); !ok || v != originalParam {
		t.Fatalf("wrong parameter value: %v (%v)", v, ok)
	}

	// change it manually and try again
	b.SetParam(1, newParam)
	params = b.Params()
	if len(params) != 1 {
		t.Fatalf("wrong parameter lenght: %d", len(params))
	}

	if v, ok := params[0].(int); !ok || v != newParam {
		t.Fatalf("wrong parameter value: %v (%v)", v, ok)
	}
}

func TestSetParamSize(t *testing.T) {
	const firstParam int = 10
	const lastParam int = 20
	const otherParam int = 30

	var b query.Builder
	b.Add("SELECT 1 FROM t WHERE id = ?", firstParam)
	b.SetParam(4, lastParam)

	params := b.Params()
	if len(params) != 4 {
		t.Fatalf("wrong parameter lenght: %d", len(params))
	}

	for i := range params {
		switch i {
		case 0:
			if v, ok := params[i].(int); !ok || v != firstParam {
				t.Fatalf("wrong parameter value at %d: %v (%v)", i, v, ok)
			}
		case 3:
			if v, ok := params[i].(int); !ok || v != lastParam {
				t.Fatalf("wrong parameter value at %d: %v (%v)", i, v, ok)
			}
		default:
			if params[i] != nil {
				t.Fatalf("parameter at position %d should be nil", i)
			}
		}
	}

	// add a new parameter, to an existing index.
	// this should not change the size
	b.SetParam(2, otherParam)

	params = b.Params()
	if len(params) != 4 {
		t.Fatalf("wrong parameter lenght: %d", len(params))
	}

	if v, ok := params[1].(int); !ok || v != otherParam {
		t.Fatalf("wrong parameter value: %v (%v)", v, ok)
	}
}
