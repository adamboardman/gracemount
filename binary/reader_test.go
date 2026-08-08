package binary

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func populate_array_from_string(str string) []uint8 {
	data_len := len(str) / 2
	uint_array := make([]uint8, data_len)
	j := 0
	for i := 0; i < data_len; i++ {
		if (str[j] & 64) > 0 {
			uint_array[i] = str[j] + 9
		} else {
			uint_array[i] = str[j] << 4
		}
		j++
		if (str[j] & 64) > 0 {
			uint_array[i] = uint_array[i] | ((str[j] + 9) & 15)
		} else {
			uint_array[i] = uint_array[i] | ((str[j]) & 15)
		}
		j++
	}
	return uint_array
}

func TestReadMessage(t *testing.T) {
	hex_message := "0102070000019f9e2ff08203003a901d0812c5e4059fffffffffffffffff002443463239443934342d333639452d344146382d383532462d3345313638363737304636357f0623736d6f6b657e04746573747d04616e6f6e9e8734623e5637f956f4ae4728a343f63ca9a40f4baccbd2dfd62bdc929b2ad7506099e4508968139a4c523f64fe7a1f99cbc1e74441373f2e982aa8b47e9c0e"
	data := populate_array_from_string(hex_message)

	Convey("Reader can read", t, func() {
		reader := BinaryReader{
			Offset:     0,
			Buffer:     data,
			BufferSize: uint32(len(data)),
			Pos:        0,
		}
		var message Message
		message.ReadIn(&reader)
		So(message.MessageId, ShouldEqual, "CF29D944-369E-4AF8-852F-3E1686770F65")
		So(message.Content, ShouldEqual, "test")
		So(message.Channel, ShouldEqual, "#smoke")
		So(message.SenderNickname, ShouldEqual, "anon")
		So(message.PacketTimestampMs, ShouldEqual, 1785065336706)
		So(message.reply_id, ShouldEqual, 0)
		So(message.message_flags, ShouldEqual, 0)
		So(message.LatitudeI, ShouldEqual, 0)
		So(message.LongitudeI, ShouldEqual, 0)
		So(message.Altitude, ShouldEqual, 0)
		So(message.Malformed, ShouldEqual, false)
	})
}
