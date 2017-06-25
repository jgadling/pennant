package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_FlagCache(t *testing.T) {
	Convey("given an empty flag cache", t, func() {
		fc, _ := NewFlagCache()
		flagName := "testing_flag"
		flag1 := Flag{Name: flagName, Description: "testing 1"}
		flag2 := Flag{Name: flagName, Description: "testing 2"}
		flag3 := Flag{Name: "second_flag", Description: "testing"}
		Convey("a row can be inserted", func() {
			err := fc.Upsert(&flag1)
			flag, _ := fc.Get(flagName)
			So(flag, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(flag.Name, ShouldEqual, flagName)
			So(flag.Description, ShouldEqual, "testing 1")
			Convey("and then deleted", func() {
				err = fc.Delete(flagName)
				So(err, ShouldBeNil)
				flag, err := fc.Get(flagName)
				So(flag, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
			Convey("flag list returns two items", func() {
				fc.Upsert(&flag3)
				list := fc.List()
				So(len(list), ShouldEqual, 2)
				So(list[flagName].Name, ShouldEqual, flagName)
				So(list["second_flag"].Name, ShouldEqual, "second_flag")
			})
			Convey("and then updated", func() {
				fc.Upsert(&flag2)
				flag, _ := fc.Get(flagName)
				So(err, ShouldBeNil)
				So(flag.Name, ShouldEqual, flagName)
				So(flag.Description, ShouldEqual, "testing 2")
				Convey("and then deleted", func() {
					err = fc.Delete(flagName)
					So(err, ShouldBeNil)
					flag, err := fc.Get(flagName)
					So(flag, ShouldBeNil)
					So(err, ShouldNotBeNil)
				})
			})
			Convey("Invalid flag lookup returns error", func() {
				flag, err = fc.Get("nothing")
				So(err, ShouldNotBeNil)
				So(flag, ShouldBeNil)
			})
			Convey("Invalid flag delete is ignored", func() {
				err = fc.Delete("nothing")
				So(err, ShouldBeNil)
			})
		})
	})
}
