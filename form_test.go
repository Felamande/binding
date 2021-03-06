// Copyright 2014 martini-contrib/binding Authors
// Copyright 2014 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package binding

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"gitea.com/lunny/tango"
	. "github.com/smartystreets/goconvey/convey"
)

var formTestCases = []formTestCase{
	{
		description:   "Happy path",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Happy path with interface",
		shouldSucceed: true,
		withInterface: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Empty payload",
		shouldSucceed: false,
		payload:       ``,
		contentType:   formContentType,
		expected:      Post{},
	},
	{
		description:   "Empty content type",
		shouldSucceed: false,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   ``,
		expected:      Post{},
	},
	{
		description:   "Malformed form body",
		shouldSucceed: false,
		payload:       `title=%2`,
		contentType:   formContentType,
		expected:      Post{},
	},
	{
		description:   "With nested and embedded structs",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Required embedded struct field not specified",
		shouldSucceed: false,
		payload:       `id=1&name=Matt+Holt`,
		contentType:   formContentType,
		expected:      BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Required nested struct field not specified",
		shouldSucceed: false,
		payload:       `title=Glorious+Post+Title&id=1`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:   "Multiple values into slice",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt&rating=4&rating=3&rating=5`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}, Ratings: []int{4, 3, 5}},
	},
	{
		description:   "Unexported field",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt&unexported=foo`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Query string POST",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Query string with Content-Type (POST request)",
		shouldSucceed: true,
		queryString:   "?title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Query string without Content-Type (GET request)",
		shouldSucceed: true,
		method:        "GET",
		queryString:   "?title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Embed struct pointer",
		shouldSucceed: true,
		deepEqual:     true,
		method:        "GET",
		queryString:   "?name=Glorious+Post+Title&email=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		expected:      EmbedPerson{&Person{Name: "Glorious Post Title", Email: "Lorem ipsum dolor sit amet"}},
	},
	{
		description:   "Embed struct pointer remain nil if not binded",
		shouldSucceed: true,
		deepEqual:     true,
		method:        "GET",
		queryString:   "?",
		payload:       ``,
		expected:      EmbedPerson{nil},
	},
	{
		description:   "Custom error handler",
		shouldSucceed: true,
		deepEqual:     true,
		method:        "GET",
		queryString:   "?",
		payload:       ``,
		expected:      CustomErrorHandle{},
	},
}

func init() {
	AddRule(&Rule{
		func(rule string) bool {
			return rule == "CustomRule"
		},
		func(_ Errors, _ string, _ interface{}) bool {
			return false
		},
	})
	SetNameMapper(nameMapper)
}

func Test_Form(t *testing.T) {
	Convey("Test form", t, func() {
		for _, testCase := range formTestCases {
			performFormTest(t, testCase)
		}
	})
}

/*
var obj interface{}
var errors Errors
*/
type FormAction struct {
	Binder
}

func (v *FormAction) Get() error {
	return v.Post()
}

func (v *FormAction) Post() error {
	errors = v.MapForm(obj)
	if errors.Len() > 0 {
		return fmt.Errorf("%+v", errors)
	}
	return nil
}

func performFormTest(t *testing.T, testCase formTestCase) {
	resp := httptest.NewRecorder()
	m := tango.Classic()
	m.Use(Bind())

	formTestHandler := func(actual interface{}, errs Errors) {
		if testCase.shouldSucceed && len(errs) > 0 {
			So(len(errs), ShouldEqual, 0)
		} else if !testCase.shouldSucceed && len(errs) == 0 {
			So(len(errs), ShouldNotEqual, 0)
		}
		expString := fmt.Sprintf("%+v", testCase.expected)
		actString := fmt.Sprintf("%+v", actual)
		if actString != expString && !(testCase.deepEqual && reflect.DeepEqual(testCase.expected, actual)) {
			So(actString, ShouldEqual, expString)
		}
	}

	obj = reflect.New(reflect.TypeOf(testCase.expected)).Interface()

	switch testCase.expected.(type) {
	case Post:
		if testCase.withInterface {
			m.Post(testRoute, new(FormAction))
		} else {
			m.Post(testRoute, new(FormAction))
			m.Get(testRoute, new(FormAction))
		}

	case BlogPost:
		m.Post(testRoute, new(FormAction))

	case EmbedPerson:
		m.Post(testRoute, new(FormAction))
		m.Get(testRoute, new(FormAction))
	case CustomErrorHandle:
		m.Get(testRoute, new(FormAction))
	}

	if len(testCase.method) == 0 {
		testCase.method = "POST"
	}

	req, err := http.NewRequest(testCase.method, testRoute+testCase.queryString, strings.NewReader(testCase.payload))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", testCase.contentType)

	m.ServeHTTP(resp, req)

	switch resp.Code {
	case http.StatusNotFound:
		if testCase.shouldSucceed {
			panic("Routing is messed up in test fixture (got 404): check methods and paths on '" + testCase.description + "'")
		}
	case http.StatusInternalServerError:
		if testCase.shouldSucceed {
			panic("Something bad happened on '" + testCase.description + "'")
		}
	default:
		if !testCase.shouldSucceed {
			panic("code not equal happened on '" + testCase.description + "'")
		}
	}

	formTestHandler(reflect.ValueOf(obj).Elem().Interface(), errors)
}

type (
	formTestCase struct {
		description   string
		shouldSucceed bool
		deepEqual     bool
		withInterface bool
		queryString   string
		payload       string
		contentType   string
		expected      interface{}
		method        string
	}
)

type defaultForm struct {
	Default string `binding:"Default(hello world)"`
}

var f defaultForm

type DefaultAction struct {
	Binder
}

func (d DefaultAction) Get() {
	d.Bind(&f)
}

func Test_Default(t *testing.T) {
	Convey("Test default value", t, func() {
		m := tango.Classic()
		m.Use(Bind())
		m.Get("/", new(DefaultAction))

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		m.ServeHTTP(resp, req)
		So(f.Default, ShouldEqual, "hello world")
	})
}
