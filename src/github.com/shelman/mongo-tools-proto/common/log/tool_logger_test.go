package log

import (
	"bytes"
	"github.com/shelman/mongo-tools-proto/common/options"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBasicToolLoggerFunctionality(t *testing.T) {
	var tl *ToolLogger
	Convey("With a new ToolLogger", t, func() {
		v1 := &options.Verbosity{
			Quiet:   false,
			Verbose: []bool{true, true, true},
		}
		tl = NewToolLogger(v1)
		So(tl, ShouldNotBeNil)
		So(tl.w, ShouldNotBeNil)
		So(tl.v, ShouldEqual, 3)

		Convey("writing the output to a buffer", func() {
			buf := bytes.NewBuffer(make([]byte, 1024))
			tl.SetWriter(buf)

			Convey("should return reasonable results", func() {
				tl.Logf(0, "test this string")
				tl.Logf(5, "this log level is too high and will not log")
				tl.Logf(2, "====!%v!====", 12.5)
				l1, _ := buf.ReadString('\n')
				So(l1, ShouldContainSubstring, ":")
				So(l1, ShouldContainSubstring, "test this string")
				l2, _ := buf.ReadString('\n')
				So(l2, ShouldContainSubstring, ":")
				So(l2, ShouldContainSubstring, "====!12.5!====")
			})
		})
	})
}

//TODO test panics, global loggger, etc
