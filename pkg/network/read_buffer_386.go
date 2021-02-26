
// +build windows

package network

import (
	"fmt"
)

func allocateReadBuffer() (*_readbuffer, error) {
	return nil, fmt.Errorf("unsupported on 386")
}

func freeReadBuffer(buf * _readbuffer) {
	// this code is unreachable because AllocateReadBuffer will not succeed on 386
}