package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelInfoScanAcceptsSQLiteTextAndBlob(t *testing.T) {
	payload := []byte(`{"is_multi_key":false,"multi_key_size":0,"multi_key_polling_index":0,"multi_key_mode":""}`)

	var fromBlob ChannelInfo
	require.NoError(t, fromBlob.Scan(payload))
	assert.False(t, fromBlob.IsMultiKey)

	var fromText ChannelInfo
	require.NoError(t, fromText.Scan(string(payload)))
	assert.Equal(t, fromBlob, fromText)

	var fromEmpty ChannelInfo
	require.NoError(t, fromEmpty.Scan([]byte{}))
	require.NoError(t, fromEmpty.Scan(""))
	require.NoError(t, fromEmpty.Scan(nil))
}
