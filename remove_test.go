package vault

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-chef/chef"
)

func TestService_pruneData(t *testing.T) {
	type data chef.DataBagItem
	var existing data
	var remove data
	var want data
	rawExist := `{
				"foo": "foo-value-1",
				"bar": "bar-value-1",
				"baz": {
					"fuz": "fuz-value-1",
					"buz": "buz-value-1"
				}
			}`
	rawRemove := `{ "baz": { "fuz": "fuz-value-1" } }`
	rawWant := `{
				"foo": "foo-value-1",
				"bar": "bar-value-1",
				"baz": {
					"buz": "buz-value-1"
				}
			}`

	if err := json.Unmarshal([]byte(rawExist), &existing); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(rawRemove), &remove); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(rawWant), &want); err != nil {
		t.Fatal(err)
	}

	got, ok := pruneData(existing, remove)
	if !ok {
		t.Errorf("pruneData did not work")
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("pruneData did not work")
	}

}
