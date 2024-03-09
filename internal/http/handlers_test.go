package http

/*
We only test that the handlers work as intended. We are not testing business logic here
*/

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"testing"

	"github.com/mehix/gopher-burrows/internal/burrows"
)

var testData = []burrows.Burrow{
	{Name: "Burrow 1"},
	{Name: "Burrow 2"},
}

type manager struct {
	data    []burrows.Burrow
	canRent bool
}

func (m *manager) CurrentStatus() []burrows.Burrow {
	return m.data
}

func (m *manager) Load(_ <-chan burrows.Burrow) {}
func (m *manager) Rentout(_ context.Context) (burrows.Burrow, error) {
	if m.canRent {
		return m.data[0], nil
	}
	return burrows.Burrow{}, errors.New("no burrows available")
}
func (m *manager) Report() burrows.Report { return burrows.Report{} }

var _ burrows.Manager = &manager{}

func TestShowStatus(t *testing.T) {

	m := &manager{data: testData}

	srvr := httptest.NewServer(Handler(m))
	defer srvr.Close()

	resp, err := http.Get(srvr.URL + "/")
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	var burrows []burrows.Burrow
	if err := json.NewDecoder(resp.Body).Decode(&burrows); err != nil {
		t.Error(err)
	}

	if !slices.Equal(testData, burrows) {
		t.Errorf("received different data. expected: %v, got: %v", testData, burrows)
	}
}

func TestRentoutSuccess(t *testing.T) {

	m := &manager{data: testData, canRent: true}

	srvr := httptest.NewServer(Handler(m))
	defer srvr.Close()

	resp, err := http.Post(srvr.URL+"/rent", "", nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	var response = struct {
		Burrow burrows.Burrow
		Error  string
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(testData[0], response.Burrow) {
		t.Errorf("received different data. expected: %v, got: %v", testData[0], response)
	}
}

func TestRentoutFail(t *testing.T) {

	m := &manager{data: testData, canRent: false}

	srvr := httptest.NewServer(Handler(m))
	defer srvr.Close()

	resp, err := http.Post(srvr.URL+"/rent", "", nil)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	var response = struct {
		Burrow burrows.Burrow
		Error  string
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Error(err)
	}

	if response.Error == "" {
		t.Errorf("expecting error but got none. received: %v", response)
	}

	if !reflect.DeepEqual(burrows.Burrow{}, response.Burrow) {
		t.Errorf("in case of error the Burrow should be empty. received: %v", response)
	}
}
