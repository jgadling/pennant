package main

import (
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/franela/goreq"
	. "github.com/smartystreets/goconvey/convey"
)

func getTestUrl(ts *httptest.Server, url string) string {
	return ts.URL + url
}

func Test_RestGet(t *testing.T) {
	Convey("Given a rest server is launched", t, func() {
		fc, _ := NewFlagCache()
		driver, _ := NewMemoryDriver()
		driver.fc = fc
		ts := httptest.NewServer(pennantRouter(fc, driver))
		defer ts.Close()

		Convey("flag list returns an empty list", func() {
			res, err := goreq.Request{Uri: getTestUrl(ts, "/flags")}.Do()
			So(err, ShouldBeNil)

			var flagList FlagListResponse
			res.Body.FromJsonTo(&flagList)

			So(flagList.Status, ShouldEqual, 200)
			So(len(flagList.Flags), ShouldEqual, 0)
		})
		Convey("A flag can be saved", func() {
			flagfile, err := ioutil.ReadFile("tests/data/flag1.json")
			res, err := goreq.Request{
				Uri:    getTestUrl(ts, "/flags"),
				Method: "POST",
				Body:   string(flagfile)}.Do()
			So(err, ShouldBeNil)

			var flagItem FlagItemResponse
			res.Body.FromJsonTo(&flagItem)

			So(flagItem.Status, ShouldEqual, 200)
			So(flagItem.Flag.Name, ShouldEqual, "red_button")

			Convey("and then fetched", func() {
				res, err := goreq.Request{
					Uri: getTestUrl(ts, "/flags/red_button")}.Do()
				So(err, ShouldBeNil)

				var flagItem FlagItemResponse
				res.Body.FromJsonTo(&flagItem)

				So(flagItem.Status, ShouldEqual, 200)
				So(flagItem.Flag.Name, ShouldEqual, "red_button")
			})
			Convey("and then listed", func() {
				res, err := goreq.Request{
					Uri: getTestUrl(ts, "/flags")}.Do()
				So(err, ShouldBeNil)

				var flagList FlagListResponse
				res.Body.FromJsonTo(&flagList)

				So(flagList.Status, ShouldEqual, 200)
				So(flagList.Flags[0], ShouldEqual, "red_button")
			})
			Convey("and then deleted", func() {
				res, err := goreq.Request{
					Method: "DELETE",
					Uri:    getTestUrl(ts, "/flags/red_button")}.Do()
				So(err, ShouldBeNil)

				var flagVal FlagValueResponse
				res.Body.FromJsonTo(&flagVal)

				So(flagVal.Status, ShouldEqual, 200)

				Convey("and then not listed", func() {
					res, err := goreq.Request{
						Uri: getTestUrl(ts, "/flags")}.Do()
					So(err, ShouldBeNil)

					var flagList FlagListResponse
					res.Body.FromJsonTo(&flagList)

					So(flagList.Status, ShouldEqual, 200)
					So(len(flagList.Flags), ShouldEqual, 0)
				})
			})
			Convey("FlagValue returns 404 for missing flag", func() {
				res, err := goreq.Request{
					Method: "POST",
					Body:   "{\"user_username\":\"foobar\"}",
					Uri:    getTestUrl(ts, "/flagValue/not_a_real_flag")}.Do()
				So(err, ShouldBeNil)

				var flagVal FlagValueResponse
				res.Body.FromJsonTo(&flagVal)

				So(flagVal.Status, ShouldEqual, 404)
				So(flagVal.Enabled, ShouldEqual, false)
			})
			Convey("And return a false value for post", func() {
				res, err := goreq.Request{
					Method: "POST",
					Body:   "{\"user_id\":1}",
					Uri:    getTestUrl(ts, "/flagValue/red_button")}.Do()
				So(err, ShouldBeNil)

				var flagVal FlagValueResponse
				res.Body.FromJsonTo(&flagVal)

				So(flagVal.Status, ShouldEqual, 200)
				So(flagVal.Enabled, ShouldEqual, false)
			})
			Convey("And return a true value for post", func() {
				res, err := goreq.Request{
					Method: "POST",
					Body:   "{\"user_id\":10}",
					Uri:    getTestUrl(ts, "/flagValue/red_button")}.Do()
				So(err, ShouldBeNil)

				var flagVal FlagValueResponse
				res.Body.FromJsonTo(&flagVal)

				So(flagVal.Status, ShouldEqual, 200)
				So(flagVal.Enabled, ShouldEqual, true)
			})
			Convey("And return a false value for get", func() {
				qstring := url.Values{}
				qstring.Set("user_username", "nobody")
				res, err := goreq.Request{
					QueryString: qstring,
					Uri:         getTestUrl(ts, "/flagValue/red_button")}.Do()
				So(err, ShouldBeNil)

				var flagVal FlagValueResponse
				res.Body.FromJsonTo(&flagVal)

				So(flagVal.Status, ShouldEqual, 200)
				So(flagVal.Enabled, ShouldEqual, false)
			})
			Convey("And return a true value for get", func() {
				qstring := url.Values{}
				qstring.Set("user_username", "foobar")
				res, err := goreq.Request{
					QueryString: qstring,
					Uri:         getTestUrl(ts, "/flagValue/red_button")}.Do()
				So(err, ShouldBeNil)

				var flagVal FlagValueResponse
				res.Body.FromJsonTo(&flagVal)

				So(flagVal.Status, ShouldEqual, 200)
				So(flagVal.Enabled, ShouldEqual, true)
			})
		})
	})
}
