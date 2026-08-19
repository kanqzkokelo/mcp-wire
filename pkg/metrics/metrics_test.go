package metrics

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicMetricsCounters(t *testing.T) {
	StreamOpened()
	assert.True(t, GlobalMetrics.ActiveStreams >= 1)

	var buf bytes.Buffer
	mw := NewMetricWriter(&buf)

	msg := []byte("hello metrics")
	n, err := mw.Write(msg)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)
	assert.True(t, GlobalMetrics.BytesSent >= int64(len(msg)))

	mr := NewMetricReader(bytes.NewReader(msg))
	readBuf := make([]byte, len(msg))
	n, err = mr.Read(readBuf)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)
	assert.True(t, GlobalMetrics.BytesReceived >= int64(len(msg)))

	StreamClosed()

	snapshot := GetStatusSnapshot("test-peer-id", []string{"/ip4/127.0.0.1/tcp/4001"}, []string{"gpu-whisper"}, map[string]string{})
	jsonStr, err := ExportStatusJSON(snapshot)
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "test-peer-id")
	assert.Contains(t, jsonStr, "gpu-whisper")
}
