package mongoimport

import (
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"os"
	"testing"
)

// TODO: currently doesn't work for lines like `a, b, "cccc,cccc", d`
func TestTSVImportDocument(t *testing.T) {
	Convey("With a TSV import input", t, func() {
		var err error
		var tsvFile, fileHandle *os.File
		Convey("integer valued strings should be converted", func() {
			contents := "1\t2\t3e\n"
			fields := []string{"a", "b", "c"}
			expectedRead := bson.M{
				"a": 1,
				"b": 2,
				"c": "3e",
			}

			tsvFile, err = ioutil.TempFile("", "mongoimport_")
			So(err, ShouldBeNil)
			_, err = io.WriteString(tsvFile, contents)
			So(err, ShouldBeNil)
			fileHandle, err := os.Open(tsvFile.Name())
			So(err, ShouldBeNil)
			tsvImporter := NewTSVImportInput(fields, fileHandle)
			bsonDoc, err := tsvImporter.ImportDocument()
			So(err, ShouldBeNil)
			So(bsonDoc, ShouldResemble, expectedRead)
		})

		Convey("extra fields should be prefixed with 'field'", func() {
			contents := "1\t2\t3e\t may\n"
			fields := []string{"a", "b", "c"}
			expectedRead := bson.M{
				"a":      1,
				"b":      2,
				"c":      "3e",
				"field3": " may",
			}

			tsvFile, err = ioutil.TempFile("", "mongoimport_")
			So(err, ShouldBeNil)
			_, err = io.WriteString(tsvFile, contents)
			So(err, ShouldBeNil)
			fileHandle, err = os.Open(tsvFile.Name())
			So(err, ShouldBeNil)
			tsvImporter := NewTSVImportInput(fields, fileHandle)
			bsonDoc, err := tsvImporter.ImportDocument()
			So(err, ShouldBeNil)
			So(bsonDoc, ShouldResemble, expectedRead)
		})

		Convey("mixed values should be parsed correctly", func() {
			contents := "12\t13.3\tInline\t14\n"
			fields := []string{"a", "b", "c", "d"}
			expectedRead := bson.M{
				"a": 12,
				"b": 13.3,
				"c": "Inline",
				"d": 14,
			}
			tsvFile, err = ioutil.TempFile("", "mongoimport_")
			So(err, ShouldBeNil)
			_, err = io.WriteString(tsvFile, contents)
			So(err, ShouldBeNil)
			fileHandle, err = os.Open(tsvFile.Name())
			So(err, ShouldBeNil)
			tsvImporter := NewTSVImportInput(fields, fileHandle)
			bsonDoc, err := tsvImporter.ImportDocument()
			So(err, ShouldBeNil)
			So(bsonDoc, ShouldResemble, expectedRead)
		})

		Convey("calling ImportDocument() in succession for TSVs should "+
			"return the correct next set of values", func() {
			contents := "1\t2\t3\n4\t5\t6\n"
			fields := []string{"a", "b", "c"}
			expectedReadOne := bson.M{
				"a": 1,
				"b": 2,
				"c": 3,
			}
			expectedReadTwo := bson.M{
				"a": 4,
				"b": 5,
				"c": 6,
			}

			tsvFile, err = ioutil.TempFile("", "mongoimport_")
			So(err, ShouldBeNil)
			_, err = io.WriteString(tsvFile, contents)
			So(err, ShouldBeNil)
			fileHandle, err := os.Open(tsvFile.Name())
			So(err, ShouldBeNil)
			tsvImporter := NewTSVImportInput(fields, fileHandle)
			bsonDoc, err := tsvImporter.ImportDocument()
			So(err, ShouldBeNil)
			So(bsonDoc, ShouldResemble, expectedReadOne)
			bsonDoc, err = tsvImporter.ImportDocument()
			So(err, ShouldBeNil)
			So(bsonDoc, ShouldResemble, expectedReadTwo)
		})

		Convey("calling ImportDocument() in succession for TSVs that contain "+
			"quotes should return the correct next set of values", func() {
			contents := "1\t2\t3\n4\t\"\t6\n"
			fields := []string{"a", "b", "c"}
			expectedReadOne := bson.M{
				"a": 1,
				"b": 2,
				"c": 3,
			}
			expectedReadTwo := bson.M{
				"a": 4,
				"b": `"`,
				"c": 6,
			}

			tsvFile, err = ioutil.TempFile("", "mongoimport_")
			So(err, ShouldBeNil)
			_, err = io.WriteString(tsvFile, contents)
			So(err, ShouldBeNil)
			fileHandle, err := os.Open(tsvFile.Name())
			So(err, ShouldBeNil)
			tsvImporter := NewTSVImportInput(fields, fileHandle)
			bsonDoc, err := tsvImporter.ImportDocument()
			So(err, ShouldBeNil)
			So(bsonDoc, ShouldResemble, expectedReadOne)
			bsonDoc, err = tsvImporter.ImportDocument()
			So(err, ShouldBeNil)
			So(bsonDoc, ShouldResemble, expectedReadTwo)
		})

		Convey("plain TSV input file sources should be parsed correctly and "+
			"subsequent imports should parse correctly",
			func() {
				fields := []string{"a", "b", "c"}
				expectedReadOne := bson.M{"a": 1, "b": 2, "c": 3}
				expectedReadTwo := bson.M{"a": 3, "b": 4.6, "c": 5}
				fileHandle, err := os.Open("testdata/test.tsv")
				So(err, ShouldBeNil)
				tsvImporter := NewTSVImportInput(fields, fileHandle)
				bsonDoc, err := tsvImporter.ImportDocument()
				So(err, ShouldBeNil)
				So(bsonDoc, ShouldResemble, expectedReadOne)
				bsonDoc, err = tsvImporter.ImportDocument()
				So(err, ShouldBeNil)
				So(bsonDoc, ShouldResemble, expectedReadTwo)
			})
		Reset(func() {
			tsvFile.Close()
			fileHandle.Close()
		})
	})
}

func TestTSVSetHeader(t *testing.T) {
	var err error
	var tsvFile, fileHandle *os.File
	Convey("With a TSV import input", t, func() {
		Convey("setting the header should read the first line of the TSV",
			func() {
				contents := "extraHeader1\textraHeader2\textraHeader3\n"
				fields := []string{}

				tsvFile, err = ioutil.TempFile("", "mongoimport_")
				So(err, ShouldBeNil)
				_, err = io.WriteString(tsvFile, contents)
				So(err, ShouldBeNil)
				fileHandle, err = os.Open(tsvFile.Name())
				So(err, ShouldBeNil)
				tsvImporter := NewTSVImportInput(fields, fileHandle)
				So(tsvImporter.SetHeader(), ShouldBeNil)
				So(len(tsvImporter.Fields), ShouldEqual, 3)
			})
		Convey("setting the header with fields already set, should "+
			"the header line with the existing fields",
			func() {
				contents := "extraHeader\textraHeader2\textraHeader3\n\n"
				fields := []string{"a", "b", "c"}

				tsvFile, err = ioutil.TempFile("", "mongoimport_")
				So(err, ShouldBeNil)
				_, err = io.WriteString(tsvFile, contents)
				So(err, ShouldBeNil)
				fileHandle, err = os.Open(tsvFile.Name())
				So(err, ShouldBeNil)
				tsvImporter := NewTSVImportInput(fields, fileHandle)
				So(tsvImporter.SetHeader(), ShouldBeNil)
				So(len(tsvImporter.Fields), ShouldEqual, 6)
			})
		Reset(func() {
			tsvFile.Close()
			fileHandle.Close()
		})
	})
}
