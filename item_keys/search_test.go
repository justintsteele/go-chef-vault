package item_keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildClientSearchPlan_WithQuery(t *testing.T) {
	query := "name:testhost*"
	plan := BuildClientSearchPlan(&query)
	if plan == nil {
		t.Fatalf("No Plan")
	}
	assert.Equal(t, query, plan.Query)
}

func TestNormalizeSearchQuery_WithNil(t *testing.T) {
	assert.Nil(t, NormalizeSearchQuery(nil))
}

func TestNormalizeSearchQuery_WithEmpty(t *testing.T) {
	assert.Nil(t, NormalizeSearchQuery(""))
}

func TestNormalizeSearchQuery_WithQuery(t *testing.T) {
	query := "name:testhost*"
	normalized := NormalizeSearchQuery(query)

	require.NotNil(t, normalized)
	assert.Equal(t, query, *normalized)
}

func TestNormalizeSearchQuery_WithStringValueNil(t *testing.T) {
	var query *string = nil
	assert.Nil(t, NormalizeSearchQuery(query))
}

func TestNormalizeSearchQuery_WithTypedNilInterface(t *testing.T) {
	var s *string = nil
	var v any = s
	assert.Nil(t, NormalizeSearchQuery(v))
}

func TestNormalizeSearchQuery_WithUnsupportedType(t *testing.T) {
	assert.Nil(t, NormalizeSearchQuery(123))
}
