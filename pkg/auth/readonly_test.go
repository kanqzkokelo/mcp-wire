package auth

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyFilter(t *testing.T) {
	// Allowed call: tools/list
	allowedInput := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	buf := bytes.NewBufferString(allowedInput)
	filter := NewReadOnlyFilterReader(buf, true)

	var out bytes.Buffer
	err := filter.ReadLineAndFilter(&out, &out)
	require.NoError(t, err)
	assert.Equal(t, allowedInput, out.String())

	// Blocked call: tools/call
	blockedInput := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"whisper"}}` + "\n"
	buf = bytes.NewBufferString(blockedInput)
	filter = NewReadOnlyFilterReader(buf, true)

	out.Reset()
	err = filter.ReadLineAndFilter(&out, &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), `-32601`)
	assert.Contains(t, out.String(), `Method 'tools/call' not allowed in read-only mode`)
}
