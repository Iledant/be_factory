package main

import "testing"

type strTest struct {
	Send     string
	Attended string
}

func TestMain(t *testing.T) {
	testCapitalizeFirst(t)
	testLowerFirst(t)
	testToSQL(t)
}

func testCapitalizeFirst(t *testing.T) {
	tcc := []strTest{
		{Send: "Housing", Attended: "Housing"},
		{Send: "housing", Attended: "Housing"},
	}
	for i, tc := range tcc {
		cap := capitalizeFirst(tc.Send)
		if cap != tc.Attended {
			t.Errorf("CapitalizeFirst[%d] : attendu -> %s   reçu -> %s", i, tc.Attended, cap)
		}
	}
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
