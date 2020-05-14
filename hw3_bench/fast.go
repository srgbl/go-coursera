package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	"io"
	"os"
	"strconv"
	"strings"
)

type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	seenBrowsers := make(map[string]bool)

	u := User{}
	_, err = fmt.Fprintln(out, "found users:")
	if err != nil {
		panic(err)
	}

	for i := 0; scanner.Scan(); i++ {
		line := scanner.Bytes()

		//fmt.Printf("%v %v\n", err, line)
		err = u.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range u.Browsers {
			if !isMSIE && strings.Contains(browser, "MSIE") {
				isMSIE = true
			}
			if !isAndroid && strings.Contains(browser, "Android") {
				isAndroid = true
			}
			if strings.Contains(browser, "MSIE") || strings.Contains(browser, "Android") {
				seenBrowsers[browser] = true
			}
		}

		if isAndroid && isMSIE {
			// log.Println("Android and MSIE user:", user["name"], user["email"])
			email := strings.ReplaceAll(u.Email, "@", " [at] ")
			fmt.Fprintln(out, "["+strconv.Itoa(i)+"] "+u.Name+" <"+email+">")
		}
	}
	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson84c0690eDecodeLearningCourseraMod1Hw3BenchData(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "name":
			out.Name = string(in.String())
		case "email":
			out.Email = string(in.String())
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson84c0690eDecodeLearningCourseraMod1Hw3BenchData(&r, v)
	return r.Error()
}
