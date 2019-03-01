package main

import "testing"

type strTest struct {
	Send     string
	Attended string
}

func TestMain(t *testing.T) {
	testLowerFirst(t)
	testToSQL(t)
}

func testLowerFirst(t *testing.T) {
	tcc := []strTest{
		{Send: "Housing", Attended: "housing"},
		{Send: "housing", Attended: "housing"},
	}
	for i, tc := range tcc {
		cap := lowerFirst(tc.Send)
		if cap != tc.Attended {
			t.Errorf("LowerFirst[%d] : attendu -> %s   reçu -> %s", i, tc.Attended, cap)
		}
	}
}

func testToSQL(t *testing.T) {
	tcc := []strTest{
		{Send: "Housing", Attended: "housing"},
		{Send: "housing", Attended: "housing"},
		{Send: "CompleteName", Attended: "complete_name"},
		{Send: "Name2", Attended: "name2"},
		{Send: "ID", Attended: "id"},
		{Send: "PLUS", Attended: "plus"},
		{Send: "EventID", Attended: "event_id"},
		{Send: "Event2ID", Attended: "event2_id"},
	}
	for i, tc := range tcc {
		cap := toSQL(tc.Send)
		if cap != tc.Attended {
			t.Errorf("ToSQL[%d] : attendu -> %s   reçu -> %s", i, tc.Attended, cap)
		}
	}
}
