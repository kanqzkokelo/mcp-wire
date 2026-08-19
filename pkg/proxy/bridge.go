package proxy

import (
	"bufio"
	"context"
	"io"
	"sync"

	"github.com/libp2p/go-libp2p/core/network"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 32*1024) // 32KB pool
		return &buf
	},
}

// BridgeStreams creates low-latency bidirectional proxy between P2P stream and process stdio
func BridgeStreams(ctx context.Context, stream network.Stream, processIn io.WriteCloser, processOut io.ReadCloser) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Direction 1: P2P Stream -> Process Stdin
	go func() {
		defer wg.Done()
		defer processIn.Close()
		copyBuffered(processIn, stream)
	}()

	// Direction 2: Process Stdout -> P2P Stream
	go func() {
		defer wg.Done()
		defer stream.CloseWrite()
		copyBuffered(stream, processOut)
	}()

	wg.Wait()
}

// BridgeStdioToStream bridges local os.Stdin / os.Stdout to a remote P2P Stream
func BridgeStdioToStream(ctx context.Context, stream network.Stream, stdIn io.Reader, stdOut io.Writer) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Direction 1: Local os.Stdin -> Remote P2P Stream
	go func() {
		defer wg.Done()
		defer stream.CloseWrite()
		copyBuffered(stream, stdIn)
	}()

	// Direction 2: Remote P2P Stream -> Local os.Stdout
	go func() {
		defer wg.Done()
		copyBuffered(stdOut, stream)
	}()

	wg.Wait()
}

func copyBuffered(dst io.Writer, src io.Reader) {
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buf := *bufPtr

	// Wrap dst in bufio.Writer if line flushing needed
	bw := bufio.NewWriterSize(dst, 32*1024)

	for {
		nr, rerr := src.Read(buf)
		if nr > 0 {
			_, werr := bw.Write(buf[:nr])
			_ = bw.Flush()
			if werr != nil {
				break
			}
		}
		if rerr != nil {
			break
		}
	}
}
