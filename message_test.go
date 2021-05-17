package tcpclient

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMessage(t *testing.T) {
	content := "a mock message"
	len := len(content)
	id := uint32(10)
	Convey("new msg", t, func() {
		msg := NewMsgPackage(id, []byte(content))
		Convey("get msg", func() {
			So(msg.GetMsgID(), ShouldEqual, id)
			So(string(msg.GetData()), ShouldEqual, content)
			So(msg.GetDataLen(), ShouldEqual, len)
		})
		Convey("set msg", func() {
			msg.SetData([]byte("short msg"))
			msg.SetDataLen(1)
			msg.SetMsgID(20)
			So(msg.GetMsgID(), ShouldEqual, 20)
			So(string(msg.GetData()), ShouldEqual, "short msg")
			So(msg.GetDataLen(), ShouldEqual, 1)
		})
	})
}
