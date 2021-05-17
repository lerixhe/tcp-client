package tcpclient

import (
	"errors"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClient(t *testing.T) {
	Convey("init a client", t, func() {
		conn := &mockConn{}
		opt := &mockClientOpt{}
		opt.ConnTimeoutOpt(1 * time.Second)
		opt.MaxReconnectCycleOpt(1 * time.Second)
		opt.RecvTimeoutOpt(1 * time.Second)

		client := NewTCPClient(conn, opt)
		go client.OnServerChange("mock.com")
		for {
			conn.mu.Lock()
			addr := conn.addr
			conn.mu.Unlock()
			if addr == "" {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			break
		}
		Convey("when open conn succeed", func() {
			conn.onOpen()
			So(client.IsConnected(), ShouldBeTrue)

		})

		Convey("when meet a err", func() {
			err := errors.New("conn err")
			conn.onError(err)
			So(client.IsConnected(), ShouldBeFalse)
		})

		Convey("when received a message", func() {
			conn.onMessage(NewMsgPackage(88, []byte("mock message")))
			client.chanMapMu.RLock()
			chanCount := len(client.MessageChanMap)
			client.chanMapMu.RUnlock()
			So(chanCount, ShouldBeGreaterThan, 0)
		})
		Convey("when Request a message with no response", func() {
			msg := NewMsgPackage(8, []byte("request message"))
			rsp, err := client.Request(msg, 0)
			So(rsp, ShouldBeEmpty)
			So(err.Error(), ShouldContainSubstring, "receive message timeout ")
		})
		Convey("when add handler", func() {
			handler := &mockHandler{}
			client.AddHandler(10, handler)

			client.chanMapMu.RLock()
			chanCount := len(client.MessageChanMap)
			client.chanMapMu.RUnlock()

			So(chanCount, ShouldBeGreaterThanOrEqualTo, 1)
			Convey("when delete handler", func() {
				client.DeleteHandler(10)
				client.chanMapMu.RLock()
				chanCountNow := len(client.MessageChanMap)
				client.chanMapMu.RUnlock()

				So(chanCount, ShouldEqual, chanCountNow+1)
			})
		})

	})

}

type mockConn struct {
	onOpen    func()
	onMessage func(Messager)
	onError   func(error)
	mu        sync.RWMutex
	addr      string
}

func (c *mockConn) OnOpen(f func()) {
	c.onOpen = f
}
func (c *mockConn) OnMessage(f func(Messager)) {
	c.onMessage = f
}
func (c *mockConn) OnError(f func(error)) {
	c.onError = f
}
func (c *mockConn) Dial(addr string) {
	c.mu.Lock()
	c.addr = addr
	c.mu.Unlock()
}
func (c *mockConn) RemoteAddr() string {
	return ""
}
func (c *mockConn) LocalAddr() string {
	return ""
}
func (c *mockConn) SetReadTimeout(readTimeout time.Duration) {

}
func (c *mockConn) Write(p Messager) (n int, err error) {
	return 0, nil
}
func (c *mockConn) Close() error {
	return nil
}
func (c *mockConn) SetWriteTimeout(writeTimeout time.Duration) {

}

type mockDataPack struct{}

func (dp *mockDataPack) GetHeadLen() uint32 {
	return 8
}
func (dp *mockDataPack) Pack(msg Messager) ([]byte, error) {
	return nil, nil
}
func (dp *mockDataPack) Unpack(binaryData []byte) (Messager, error) {
	return NewMsgPackage(88, binaryData), nil
}

type mockClientOpt struct{}

func (opt *mockClientOpt) MaxReconnectCycleOpt(defValue time.Duration) time.Duration {
	return defValue
}
func (opt *mockClientOpt) RecvTimeoutOpt(defValue time.Duration) time.Duration {
	return defValue
}
func (opt *mockClientOpt) ConnTimeoutOpt(defValue time.Duration) time.Duration {
	return defValue
}

type mockHandler struct {
}

func (h *mockHandler) Serve(msg ReqMessager, sender RspMessagerWriter) {}
