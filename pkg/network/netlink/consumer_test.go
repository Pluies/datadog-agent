// +build linux_bpf

package netlink

import (
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConsumerKeepsRunningAfterCircuitBreakerTrip(t *testing.T) {
	c := NewConsumer("/proc", 1, false)
	require.NotNil(t, c)
	exited := make(chan struct{})
	defer func() {
		c.Stop()
		<-exited
	}()

	ev, err := c.Events()
	require.NoError(t, err)
	require.NotNil(t, ev)

	go func() {
		for range ev {
		}

		close(exited)
	}()

	isRecvLoopRunning := func() bool {
		return atomic.LoadInt32(&c.recvLoopRunning) == 1
	}

	require.Eventually(t, func() bool {
		return isRecvLoopRunning()
	}, 3*time.Second, 100*time.Millisecond)

	srv := startServerTCP(t, net.ParseIP("127.0.0.1"), 0)
	defer srv.Close()

	l := srv.(net.Listener)

	// this should trip the circuit breaker
	// we have to run this loop for more than
	// `tickInterval` seconds (3s currently) for
	// the circuit breaker to detect the over-limit
	// rate of updates
	for i := 0; i < 16; i++ {
		conn, err := net.Dial("tcp", l.Addr().String())
		require.NoError(t, err)
		defer conn.Close()

		time.Sleep(250 * time.Millisecond)
	}

	require.Eventually(t, func() bool {
		return c.samplingRate < 1.0
	}, 3*time.Second, 100*time.Millisecond)

	require.True(t, isRecvLoopRunning())
}
