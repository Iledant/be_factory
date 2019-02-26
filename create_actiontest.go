package main

import (
	"errors"
	"os"
)

func createTest(t *Table) error {
	file, err := os.Create("./actions/" + t.SQLName + "_test.go")
	if err != nil {
		return errors.New("Création du fichier test " + err.Error())
	}
	defer file.Close()
	jsonFields := t.generateJSONFields(false)
	modJSONFields := t.generateJSONFields(true)
	content := `package actions
	
	import (
		"fmt"
		"net/http"
		"strconv"
		"strings"
		"testing"
	)
	
	// test` + t.Name + ` is the entry point for testing all renew projet requests
	func test` + t.Name + `(t *testing.T, c *TestContext) {
		t.Run("` + t.Name + `", func(t *testing.T) {
			ID := testCreate` + t.Name + `(t, c)
			if ID == 0 {
				t.Error("Impossible de créer le ` + lowerFirst(t.FrenchName) + `")
				t.FailNow()
				return
			}
			testUpdate` + t.Name + `(t, c, ID)
			testGet` + t.Name + `s(t, c)
			testDelete` + t.Name + `(t, c, ID)
		})
	}
	
	// testCreate` + t.Name + ` checks if route is admin protected and created budget action
	// is properly filled
	func testCreate` + t.Name + `(t *testing.T, c *TestContext) (ID int) {
		tcc := []TestCase{
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{` + jsonFields + `}}` + "`" + `),
				Token:        c.Config.Users.User.Token,
				RespContains: []string{` + "`" + `Droits administrateur requis` + "`" + `},
				StatusCode:   http.StatusUnauthorized}, // 0 : user unauthorized
			{Sent: []byte(` + "`" + `fake` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `Création de ` + lowerFirst(t.FrenchName) + `, décodage :` + "`" + `},
				StatusCode:   http.StatusInternalServerError}, // 1 : bad request
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{}}` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `Création de ` + lowerFirst(t.FrenchName) + ` : Champ reference incorrect` + "`" + `},
				StatusCode:   http.StatusBadRequest}, // 2 : TODO validation check
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{` + jsonFields + `}` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `"` + t.Name + `":{"ID":1,` + jsonFields + `` + "`" + `},
				StatusCode:   http.StatusCreated}, // 3 : ok
		}
		for i, tc := range tcc {
			response := c.E.POST("/api/` + toSQL(t.Name) + `").WithBytes(tc.Sent).
				WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			body := string(response.Content)
			for _, r := range tc.RespContains {
				if !strings.Contains(body, r) {
					t.Errorf("Create` + t.Name + `[%d]\n  ->attendu %s\n  ->reçu: %s", i, r, body)
				}
			}
			status := response.Raw().StatusCode
			if status != tc.StatusCode {
				t.Errorf("Create` + t.Name + `[%d]  ->status attendu %d  ->reçu: %d", i, tc.StatusCode, status)
			}
			if tc.StatusCode == http.StatusCreated {
				fmt.Sscanf(body, ` + "`" + `{"` + t.Name + `":{"ID":%d` + "`" + `, &ID)
			}
		}
		return ID
	}
	
	// testUpdate` + t.Name + ` checks if route is admin protected and Updated budget action
	// is properly filled
	func testUpdate` + t.Name + `(t *testing.T, c *TestContext, ID int) {
		tcc := []TestCase{
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{` + modJSONFields + `}}` + "`" + `),
				Token:        c.Config.Users.User.Token,
				RespContains: []string{` + "`" + `Droits administrateur requis` + "`" + `},
				StatusCode:   http.StatusUnauthorized}, // 0 : user unauthorized
			{Sent: []byte(` + "`" + `fake` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `Modification de ` + lowerFirst(t.FrenchName) + `, décodage :` + "`" + `},
				StatusCode:   http.StatusInternalServerError}, // 1 : bad request
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{}}` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `Modification de ` + lowerFirst(t.FrenchName) + ` : Champ reference incorrect` + "`" + `},
				StatusCode:   http.StatusBadRequest}, // 2 : TODO validation check
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{"ID":0,` + modJSONFields + `}}` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `Modification de ` + lowerFirst(t.FrenchName) + `, requête : ` + "`" + `},
				StatusCode:   http.StatusInternalServerError}, // 3 : bad ID
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{"ID":` + "`" + ` + strconv.Itoa(ID) + ` + "`" + `,` + modJSONFields + `}}` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `"` + t.Name + `":{"ID":` + "`" + ` + strconv.Itoa(ID) + ` + "`" + `,` + modJSONFields + `}` + "`" + `},
				StatusCode:   http.StatusOK}, // 4 : ok
		}
		for i, tc := range tcc {
			response := c.E.PUT("/api/` + toSQL(t.Name) + `").WithBytes(tc.Sent).
				WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			body := string(response.Content)
			for _, r := range tc.RespContains {
				if !strings.Contains(body, r) {
					t.Errorf("Update` + t.Name + `[%d]\n  ->attendu %s\n  ->reçu: %s", i, r, body)
				}
			}
			status := response.Raw().StatusCode
			if status != tc.StatusCode {
				t.Errorf("Update` + t.Name + `[%d]  ->status attendu %d  ->reçu: %d", i, tc.StatusCode, status)
			}
		}
	}
	
	// testGet` + t.Name + `s checks if route is user protected and ` + t.Name + `s correctly sent back
	func testGet` + t.Name + `s(t *testing.T, c *TestContext) {
		tcc := []TestCase{
			{Token: "",
				RespContains: []string{` + "`" + `Token absent` + "`" + `},
				Count:        1,
				StatusCode:   http.StatusInternalServerError}, // 0 : token empty
			{Token: c.Config.Users.User.Token,
				RespContains: []string{` + "`" + `{"` + t.Name + `":[{"ID":1,` + modJSONFields + `}]}` + "`" + `},
				Count:        1,
				StatusCode:   http.StatusOK}, // 1 : ok
		}
		for i, tc := range tcc {
			response := c.E.GET("/api/` + toSQL(t.Name) + `s").
				WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			body := string(response.Content)
			for _, r := range tc.RespContains {
				if !strings.Contains(body, r) {
					t.Errorf("Get` + t.Name + `s[%d]\n  ->attendu %s\n  ->reçu: %s", i, r, body)
				}
			}
			status := response.Raw().StatusCode
			if status != tc.StatusCode {
				t.Errorf("Get` + t.Name + `s[%d]  ->status attendu %d  ->reçu: %d", i, tc.StatusCode, status)
			}
			if status == http.StatusOK {
				count := strings.Count(body, ` + "`" + `"ID"` + "`" + `)
				if count != tc.Count {
					t.Errorf("Get` + t.Name + `s[%d]  ->nombre attendu %d  ->reçu: %d", i, tc.Count, count)
				}
			}
		}
	}
	
	// testDelete` + t.Name + ` checks if route is user protected and ` + toSQL(t.Name) + `s correctly sent back
	func testDelete` + t.Name + `(t *testing.T, c *TestContext, ID int) {
		tcc := []TestCase{
			{Token: c.Config.Users.User.Token,
				RespContains: []string{` + "`" + `Droits administrateur requis` + "`" + `},
				StatusCode:   http.StatusUnauthorized}, // 0 : user token
			{Token: c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `Suppression de ` + lowerFirst(t.FrenchName) + `, requête : ` + "`" + `},
				ID:           0,
				StatusCode:   http.StatusInternalServerError}, // 1 : bad ID
			{Token: c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `Logement supprimé` + "`" + `},
				ID:           ID,
				StatusCode:   http.StatusOK}, // 2 : ok
		}
		for i, tc := range tcc {
			response := c.E.DELETE("/api/` + toSQL(t.Name) + `/"+strconv.Itoa(tc.ID)).
				WithHeader("Authorization", "Bearer "+tc.Token).Expect()
			body := string(response.Content)
			for _, r := range tc.RespContains {
				if !strings.Contains(body, r) {
					t.Errorf("Delete` + t.Name + `[%d]\n  ->attendu %s\n  ->reçu: %s", i, r, body)
				}
			}
			status := response.Raw().StatusCode
			if status != tc.StatusCode {
				t.Errorf("Delete` + t.Name + `[%d]  ->status attendu %d  ->reçu: %d", i, tc.StatusCode, status)
			}
		}
	}
	`
	_, err = file.WriteString(content)
	if err != nil {
		return errors.New("Erreur d'écriture du fichier action : " + err.Error())
	}
	return nil
}
