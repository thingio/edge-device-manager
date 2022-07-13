package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/thingio/edge-device-std/errors"
	"os"
)

// SendWSMessage will upgrade an HTTP connection to a WebSocket connection,
// and then sends message from the bus into this connection until it is closed
func SendWSMessage(request *restful.Request, response *restful.Response, bus <-chan interface{}, stop func()) error {
	conn, _, _, err := ws.UpgradeHTTP(request.Request, response.ResponseWriter)
	if err != nil {
		return errors.Internal.Cause(err, "fail to upgrade HTTP as WebSocket")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer func() {
			cancel()
			_ = conn.Close()
		}()
		for {
			data, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				if err.Error() == "EOF" { // the connection is already disconnected
					return
				}
				continue
			}
			if string(data) == "stop" {
				return
			}
		}
	}()
	for {
		select {
		case event := <-bus:
			if data, err := json.Marshal(event); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, err.Error())
				break
			} else {
				_ = wsutil.WriteServerMessage(conn, ws.OpText, data)
			}
		case <-ctx.Done():
			stop()
			return nil
		}
	}
}
