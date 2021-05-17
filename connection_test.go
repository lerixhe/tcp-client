package tcpclient

import (
	"errors"
	"net"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/nettest"
)

func TestTCPConnection(t *testing.T) {
	Convey("init a conn", t, func() {
		listener, _ := nettest.NewLocalListener("tcp")
		go func() {
			conn, _ := listener.Accept()
			data, _ := NewDataPack().Pack(NewMsgPackage(66, []byte("hello")))
			conn.Write(data)
		}()

		var connErr error
		var connMsg Messager
		var connected bool
		tcpConn := NewTCPConnection()

		tcpConn.OnError(func(err error) {
			connErr = err
			connected = false
		})
		tcpConn.OnMessage(func(msg Messager) {
			connMsg = msg
		})
		tcpConn.OnOpen(func() {
			connected = true
		})
		Convey("when dial to a not avalid server addr", func() {
			tcpConn.Dial("111abc")
			So(connErr.Error(), ShouldContainSubstring, "missing port in address")
			So(connMsg, ShouldBeNil)
			So(connected, ShouldBeFalse)
		})
		Convey("when dial to a closed server", func() {
			tcpConn.Dial("8.8.8.8:1234")
			So(connErr.Error(), ShouldContainSubstring, "i/o timeout")
			So(connMsg, ShouldBeNil)
			So(connected, ShouldBeFalse)
		})
		Convey("when dial to a normal server", func() {
			tcpConn.Dial("8.8.8.8:1234")
			So(connErr.Error(), ShouldContainSubstring, "i/o timeout")
			So(connMsg, ShouldBeNil)
			So(connected, ShouldBeFalse)
		})
		Convey("when write normal", func() {
			go tcpConn.Dial(listener.Addr().String())
			time.Sleep(5 * time.Second)
			tcpConn.SetWriteTimeout(10 * time.Second)
			tcpConn.SetReadTimeout(10 * time.Second)
			tcpConn.Write(NewMsgPackage(88, []byte("hello")))
			So(connMsg.GetMsgID(), ShouldNotBeEmpty)
		})
	})

}

type mockNetConn struct {
	makeReadErr  bool
	makeWriteErr bool
}

func (c *mockNetConn) Read(b []byte) (n int, err error) {
	if c.makeReadErr {
		return 0, errors.New("read err")
	}
	return
}

func (c *mockNetConn) Write(b []byte) (n int, err error) {
	if c.makeWriteErr {
		return 0, errors.New("write err")
	}
	return
}

func (c *mockNetConn) Close() error {
	return nil
}
func (c *mockNetConn) LocalAddr() net.Addr {
	return nil
}
func (c *mockNetConn) RemoteAddr() net.Addr {
	return nil
}
func (c *mockNetConn) SetDeadline(t time.Time) error {
	return nil
}
func (c *mockNetConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (c *mockNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}
