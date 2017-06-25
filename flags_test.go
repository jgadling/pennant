package main

import (
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Flags(t *testing.T) {
	Convey("a flag json file can be loaded and parsed", t, func() {
		flagfile, err := ioutil.ReadFile("tests/data/flag1.json")
		flag, err := LoadAndParseFlag(flagfile)
		So(err, ShouldBeNil)
		So(flag.Name, ShouldEqual, "red_button")
		So(len(flag.Policies), ShouldEqual, 3)
		for _, policy := range flag.Policies {
			So(policy.ParsedExpr, ShouldNotBeNil)
		}
		Convey("Match zero policies", func() {
			doc1 := map[string]interface{}{
				"user_username": "nobody",
				"user_id":       1,
			}
			So(flag.GetValue(doc1), ShouldEqual, false)
		})
		Convey("Match first policy", func() {
			doc1 := map[string]interface{}{
				"user_username": "foo",
				"user_id":       1,
			}
			So(flag.GetValue(doc1), ShouldEqual, true)
		})
		Convey("Match second policy", func() {
			doc1 := map[string]interface{}{
				"user_username": "nobody",
				"user_id":       10,
			}
			So(flag.GetValue(doc1), ShouldEqual, true)
		})
		Convey("Match third policy", func() {
			doc1 := map[string]interface{}{
				"user_username": "jessfraz",
				"user_id":       1,
			}
			So(flag.GetValue(doc1), ShouldEqual, true)
		})
		Convey("useless document", func() {
			doc1 := map[string]interface{}{}
			So(flag.GetValue(doc1), ShouldEqual, false)
		})
	})
	Convey("a flag json file can be loaded and not parsed", t, func() {
		flagfile, err := ioutil.ReadFile("tests/data/flag1.json")
		flag, err := LoadFlagJson(flagfile)
		So(err, ShouldBeNil)
		So(flag.Name, ShouldEqual, "red_button")
		for _, policy := range flag.Policies {
			So(policy.ParsedExpr, ShouldBeNil)
		}
		So(len(flag.Policies), ShouldEqual, 3)
	})
	Convey("an invalid json raises errors", t, func() {
		flagfile := []byte("{'garbage:")
		_, err := LoadFlagJson(flagfile)
		So(err, ShouldNotBeNil)
	})
}
