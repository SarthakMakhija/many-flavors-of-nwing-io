package conn

import (
	"bufio"
	"errors"
	proto2 "many-flavors-of-nwing-io/non_blocking_busy_waiting/proto"
	"net"
	"time"
)

const maxTimeoutErrorsTolerable = 10

// ConnectionReader represents an abstraction to read from the connection.
type ConnectionReader struct {
	connection     net.Conn
	closeChannel   chan struct{}
	bufferedReader *bufio.Reader
	netError       net.Error
}

// NewConnectionReader creates a new instance of ConnectionReader.
func NewConnectionReader(connection net.Conn) ConnectionReader {
	return ConnectionReader{
		connection:     connection,
		bufferedReader: bufio.NewReader(connection),
		closeChannel:   make(chan struct{}),
	}
}

// AttemptReadOrErrorOut attempts to read from the incoming TCP connection.
func (connectionReader ConnectionReader) AttemptReadOrErrorOut() (*proto2.KeyValueMessage, error) {
	totalTimeoutsErrors := 0
	for {
		select {
		case <-connectionReader.closeChannel:
			return nil, errors.New("ConnectionReader is closed")
		default:
			_ = connectionReader.connection.SetReadDeadline(time.Now().Add(120 * time.Millisecond))

			message, err := proto2.DeserializeFrom(connectionReader.bufferedReader)
			if err != nil {
				if errors.As(err, &connectionReader.netError) && connectionReader.netError.Timeout() {
					totalTimeoutsErrors += 1
				}
				if totalTimeoutsErrors <= maxTimeoutErrorsTolerable {
					continue
				}
				return nil, err
			}
			return message, nil
		}
	}
}

// Close closes the ConnectionReader.
func (connectionReader ConnectionReader) Close() {
	close(connectionReader.closeChannel)
}