package tgbotapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddNonEmpty(t *testing.T) {
	params := make(Params)
	params.AddNonEmpty("value", "value")
	assert.Len(t, params, 1)
	assert.Equal(t, "value", params["value"])
	params.AddNonEmpty("test", "")
	assert.Len(t, params, 1)
	assert.Equal(t, "", params["test"])
}

func TestAddNonZero(t *testing.T) {
	params := make(Params)
	params.AddNonZero("value", 1)
	assert.Len(t, params, 1)
	assert.Equal(t, "1", params["value"])
	params.AddNonZero("test", 0)
	assert.Len(t, params, 1)
	assert.Equal(t, "", params["test"])
}

func TestAddNonZero64(t *testing.T) {
	params := make(Params)
	params.AddNonZero64("value", 1)
	assert.Len(t, params, 1)
	assert.Equal(t, "1", params["value"])
	params.AddNonZero64("test", 0)
	assert.Len(t, params, 1)
	assert.Equal(t, "", params["test"])
}

func TestAddBool(t *testing.T) {
	params := make(Params)
	params.AddBool("value", true)
	assert.Len(t, params, 1)
	assert.Equal(t, "true", params["value"])
	params.AddBool("test", false)
	assert.Len(t, params, 1)
	assert.Equal(t, "", params["test"])
}

func TestAddNonZeroFloat(t *testing.T) {
	params := make(Params)
	params.AddNonZeroFloat("value", 1)
	assert.Len(t, params, 1)
	assert.Equal(t, "1.000000", params["value"])
	params.AddNonZeroFloat("test", 0)
	assert.Len(t, params, 1)
	assert.Equal(t, "", params["test"])
}

func TestAddInterface(t *testing.T) {
	params := make(Params)
	data := struct {
		Name string `json:"name"`
	}{
		Name: "test",
	}
	require.NoError(t, params.AddInterface("value", data))
	assert.Len(t, params, 1)
	assert.Equal(t, `{"name":"test"}`, params["value"])
	require.NoError(t, params.AddInterface("test", nil))
	assert.Len(t, params, 1)
	assert.Equal(t, "", params["test"])
}

func TestAddFirstValid(t *testing.T) {
	params := make(Params)
	require.NoError(t, params.AddFirstValid("value", 0, "", "test"))
	assert.Len(t, params, 1)
	assert.Equal(t, "test", params["value"])
	require.NoError(t, params.AddFirstValid("value2", 3, "test"))
	assert.Len(t, params, 2)
	assert.Equal(t, "3", params["value2"])
}
