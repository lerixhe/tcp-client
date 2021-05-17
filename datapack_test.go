package tcpclient

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDataPack(t *testing.T) {
	Convey("new datapack", t, func() {
		p := NewDataPack()
		Convey("head len judge", func() {
			headLen := p.GetHeadLen()
			So(headLen, ShouldEqual, 8)
		})
		Convey("when pack nil", func() {
			msg := NewMsgPackage(99, nil)
			_, err := p.Pack(msg)
			So(err, ShouldBeEmpty)
		})
		Convey("create a message and pack", func() {
			msg := NewMsgPackage(99, []byte("test message"))
			data, err := p.Pack(msg)
			So(err, ShouldBeEmpty)
			So(data, ShouldNotBeNil)
			Convey("unpack the message", func() {
				msg, err := p.Unpack(data)
				So(err, ShouldBeEmpty)
				So(msg.GetMsgID(), ShouldEqual, 99)
				So(string(msg.GetData()), ShouldEqual, "")
			})
		})

	})
}
