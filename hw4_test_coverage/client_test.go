package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// код писать тут

var (
	file, _     = os.Open("dataset.xml")
	data, _     = ioutil.ReadAll(file)
	AccessToken = ""
)

type xmlUsers struct {
	XMLName xml.Name   `xml:"root"`
	Users   []UserTemp `xml:"row"`
}

type UserTemp struct {
	Id     int
	First  string `xml:"first_name"`
	Last   string `xml:"last_name"`
	Age    int
	About  string
	Gender string
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// Access Token check
	if r.Header.Get("AccessToken") != AccessToken {
		http.Error(w, "Bad AccessToken", http.StatusUnauthorized)
	}

	// Parse query string
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	orderBy, err := strconv.ParseInt(r.FormValue("order_by"), 10, 32)
	if err != nil || orderBy > 1 || orderBy < -1 {
		errResp, err := json.Marshal(SearchErrorResponse{"ErrorBadOrderBy"})
		if err != nil {
			http.Error(w, string(data)+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, string(errResp), http.StatusBadRequest)
		return
	}
	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	offset, err := strconv.ParseInt(r.FormValue("offset"), 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Unmarshall XML
	var xmlUsers xmlUsers
	err = xml.Unmarshal(data, &xmlUsers)
	if err != nil {
		http.Error(w, string(data)+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create result slice of users
	var res []User
	for _, user := range xmlUsers.Users {
		name := user.First + " " + user.Last
		if strings.Contains(name, query) || strings.Contains(user.About, query) {
			res = append(res, User{user.Id, name, user.Age, user.About, user.Gender})
		}
	}

	// Sort slice of users
	if orderBy == 1 {
		switch orderField {
		case "", "Name":
			sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
		case "Age":
			sort.Slice(res, func(i, j int) bool { return res[i].Age < res[j].Age })
		case "Id":
			sort.Slice(res, func(i, j int) bool { return res[i].Id < res[j].Id })
		}
	}
	if orderBy == -1 {
		switch orderField {
		case "", "Name":
			sort.Slice(res, func(i, j int) bool { return res[i].Name > res[j].Name })
		case "Age":
			sort.Slice(res, func(i, j int) bool { return res[i].Age > res[j].Age })
		case "Id":
			sort.Slice(res, func(i, j int) bool { return res[i].Id > res[j].Id })
		}
	}

	switch orderField {
	case "", "Name", "Age", "Id":
		break
	default:
		errResp, err := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		if err != nil {
			http.Error(w, string(data)+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, string(errResp), http.StatusBadRequest)
		return
	}

	// Offset
	if int(offset) > len(res) || offset < 0 {
		http.Error(w, "Bad offset param", http.StatusBadRequest)
		return
	}
	res = res[offset:]

	// Limit
	if int(limit) < len(res) {
		res = res[:limit]
	}

	body, err := json.Marshal(res)
	if err != nil {
		http.Error(w, string(data)+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

type TestCase struct {
	Req     SearchRequest
	Resp    *SearchResponse
	isError bool
}

func testHelper(cases []TestCase, sc *SearchClient, t *testing.T) {
	for caseNum, c := range cases {
		resp, err := sc.FindUsers(c.Req)
		if resp != nil && c.Resp == nil {
			t.Errorf("[%d] unexpected response %#v", caseNum, resp)
		}
		if resp == nil && c.Resp != nil {
			t.Errorf("[%d] expected response, got nil", caseNum)
		}
		if err != nil && !c.isError {
			t.Errorf("[%d] unexpected error %#v", caseNum, err)
		}
		if err == nil && c.isError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
	}
}

func TestFindUsersLimit(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
				Limit: -1,
			},
			nil,
			true,
		},
		TestCase{
			SearchRequest{
				Limit: 0,
			},
			&SearchResponse{},
			false,
		},
		TestCase{
			SearchRequest{
				Offset: -1,
			},
			nil,
			true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{"", ts.URL}
	testHelper(cases, sc, t)
}

func TestFindUsersRareQuery(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
				Limit: 1,
				Query: "2345lk3h53jg535666v4664gg4vv34g5",
			},
			&SearchResponse{},
			false,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{"", ts.URL}
	testHelper(cases, sc, t)
}

func TestFindUsersBadJSON(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
			},
			nil,
			true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("bad JSON"))
		}))
	sc := &SearchClient{"", ts.URL}
	testHelper(cases, sc, t)
}

func TestFindUsersInternalServerError(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
			},
			nil,
			true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}))
	sc := &SearchClient{"", ts.URL}
	testHelper(cases, sc, t)
}

func TestFindUsersTimeoutError(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
			},
			nil,
			true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusGatewayTimeout)
		}))
	sc := &SearchClient{"", ts.URL}
	testHelper(cases, sc, t)
}

func TestFindUsersUnknoenServerError(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
			},
			nil,
			true,
		},
	}
	sc := &SearchClient{"", ""}
	testHelper(cases, sc, t)
}

func TestFindUsersBadRequest(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
				Offset: math.MaxInt64,
			},
			nil,
			true,
		},
		TestCase{
			SearchRequest{
				OrderBy: 2,
			},
			nil,
			true,
		},
		TestCase{
			SearchRequest{
				OrderField: "Bad value",
			},
			nil,
			true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{"", ts.URL}
	testHelper(cases, sc, t)
}

func TestFindUsersLimit25(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
				Limit: 25,
			},
			&SearchResponse{},
			false,
		},
		TestCase{
			SearchRequest{
				Limit: 30,
			},
			&SearchResponse{},
			false,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{"", ts.URL}
	testHelper(cases, sc, t)
}

func TestFindUsersBadAccessToken(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest{
			},
			nil,
			true,
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	sc := &SearchClient{"bad access token", ts.URL}
	testHelper(cases, sc, t)
}
