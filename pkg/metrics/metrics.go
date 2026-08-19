package metrics

import (
	"encoding/json"
	"io"
	"runtime"
	"sync/atomic"
	"time"
)

type MetricsCollector struct {
	BytesSent        int64 `json:"bytes_sent"`
	BytesReceived    int64 `json:"bytes_received"`
	ActiveStreams    int64 `json:"active_streams"`
	TotalConnections int64 `json:"total_connections"`
}

var GlobalMetrics MetricsCollector

func RecordBytesSent(n int64) {
	atomic.AddInt64(&GlobalMetrics.BytesSent, n)
}

func RecordBytesReceived(n int64) {
	atomic.AddInt64(&GlobalMetrics.BytesReceived, n)
}

func StreamOpened() {
	atomic.AddInt64(&GlobalMetrics.ActiveStreams, 1)
	atomic.AddInt64(&GlobalMetrics.TotalConnections, 1)
}

func StreamClosed() {
	atomic.AddInt64(&GlobalMetrics.ActiveStreams, -1)
}

type MetricWriter struct {
	w io.Writer
}

func NewMetricWriter(w io.Writer) *MetricWriter {
	return &MetricWriter{w: w}
}

func (mw *MetricWriter) Write(p []byte) (int, error) {
	n, err := mw.w.Write(p)
	if n > 0 {
		RecordBytesSent(int64(n))
	}
	return n, err
}

type MetricReader struct {
	r io.Reader
}

func NewMetricReader(r io.Reader) *MetricReader {
	return &MetricReader{r: r}
}

func (mr *MetricReader) Read(p []byte) (int, error) {
	n, err := mr.r.Read(p)
	if n > 0 {
		RecordBytesReceived(int64(n))
	}
	return n, err
}

type StatusSnapshot struct {
	PeerID           string            `json:"peer_id"`
	ListenAddresses  []string          `json:"listen_addresses"`
	ActiveStreams    int64             `json:"active_streams"`
	TotalConnections int64             `json:"total_connections"`
	BytesSent        int64             `json:"bytes_sent"`
	BytesReceived    int64             `json:"bytes_received"`
	HeapAllocBytes   uint64            `json:"heap_alloc_bytes"`
	NumGoroutines    int               `json:"num_goroutines"`
	Uptime           string            `json:"uptime"`
	Services         []string          `json:"services"`
	ConnectedPeers   map[string]string `json:"connected_peers"`
}

var startTime = time.Now()

func GetStatusSnapshot(peerID string, listenAddrs []string, services []string, connectedPeers map[string]string) StatusSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return StatusSnapshot{
		PeerID:           peerID,
		ListenAddresses:  listenAddrs,
		ActiveStreams:    atomic.LoadInt64(&GlobalMetrics.ActiveStreams),
		TotalConnections: atomic.LoadInt64(&GlobalMetrics.TotalConnections),
		BytesSent:        atomic.LoadInt64(&GlobalMetrics.BytesSent),
		BytesReceived:    atomic.LoadInt64(&GlobalMetrics.BytesReceived),
		HeapAllocBytes:   m.HeapAlloc,
		NumGoroutines:    runtime.NumGoroutine(),
		Uptime:           time.Since(startTime).Truncate(time.Second).String(),
		Services:         services,
		ConnectedPeers:   connectedPeers,
	}
}

func ExportStatusJSON(snapshot StatusSnapshot) (string, error) {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
