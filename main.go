package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

// TODO: split main.go for more readable file

// Field embeddes information about table's field
type Field struct {
	Name     string
	Nullable bool
	GoName   string
	GoType   string
	SQLName  string
	SQLType  string
}

// Table embeddes all datas for table creation
type Table struct {
	Name       string
	FrenchName string
	SQLName    string
	Fields     []Field
}

var fieldsPrompt = []prompt.Suggest{
	{Text: "ID", Description: "Clé primaire"},
	{Text: "bigint", Description: "Entier long"},
	{Text: "int", Description: "Entier classique"},
	{Text: "boolean", Description: "Booléen"},
	{Text: "varchar", Description: "Texte de longueur variable"},
	{Text: "text", Description: "Texte sans limite"},
	{Text: "double precision", Description: "Double précision de 64 bits"},
	{Text: "date", Description: "Date"},
}

var yesNoPrompt = []prompt.Suggest{
	{Text: "oui", Description: "Null possible"},
	{Text: "non", Description: "Champ jamais nul"},
}

func emptyCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func fieldCompleter(in prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(fieldsPrompt, in.GetWordBeforeCursor(), true)
}

func yesNoCompleter(in prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(yesNoPrompt, in.GetWordBeforeCursor(), true)
}

func capitalizeFirst(s string) string {
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-'a'+'A') + s[1:]
	}
	return s
}

func lowerFirst(s string) string {
	if s[0] >= 'A' && s[0] <= 'Z' {
		return string(s[0]-'A'+'a') + s[1:]
	}
	return s
}

func toSQL(s string) (formatted string) {
	if s[0] >= 'A' && s[0] <= 'Z' {
		formatted = string(s[0] - 'A' + 'a')
	} else {
		formatted = string(s[0])
	}
	for i := 1; i < len(s); i++ {
		c0 := s[i-1]
		c1 := s[i]
		if (c0 < 'A' || c0 > 'Z') && c1 >= 'A' && c1 <= 'Z' {
			formatted += "_" + string(c1-'A'+'a')
		} else {
			if c1 >= 'A' && c1 <= 'Z' {
				formatted += string(c1 - 'A' + 'a')

			} else {
				formatted += string(c1)
			}
		}
	}
	return formatted
}

func validateField(name string) error {
	for _, t := range fieldsPrompt {
		if name == t.Text && name != "varchar" {
			return nil
		}
	}
	if len(name) >= 7 && name[0:7] == "varchar" {
		var length int
		_, err := fmt.Sscanf(name, "varchar(%d)", &length)
		if err != nil {
			return errors.New(`varchar doit être suivi de la longueur : varchar(50)`)
		}
		return nil
	}
	return errors.New("Type de champ non reconnu")
}

func getFields() (*Table, error) {
	table := &Table{}
	table.Name = strings.TrimSpace(prompt.Input("Nom du modèle ? ", emptyCompleter,
		prompt.OptionPrefixTextColor(prompt.Green)))
	if table.Name == "" {
		return nil, errors.New("Nom du modèle vide, arrêt de be_factory")
	}
	table.Name = capitalizeFirst(table.Name)
	table.SQLName = toSQL(table.Name)
	for table.FrenchName == "" {
		table.FrenchName = prompt.Input("Nom français du modèle ? ", emptyCompleter,
			prompt.OptionPrefixTextColor(prompt.Green))
		if table.FrenchName == "" {
			fmt.Println("Le nom français ne peut pas être vide")
		}
	}
	table.FrenchName = capitalizeFirst(table.FrenchName)
	var fields []Field
	for i := 1; ; i++ {
		fieldName := prompt.Input("  Nom du champ n°"+strconv.Itoa(i)+" ? ", emptyCompleter,
			prompt.OptionPrefixTextColor(prompt.Yellow))
		if fieldName == "" {
			break
		}
		var fieldType string
		for {
			fieldType = prompt.Input("  Type du champ "+fieldName+" ? ", fieldCompleter,
				prompt.OptionPrefixTextColor(prompt.Yellow))
			err := validateField(fieldType)
			if err == nil {
				break
			}
			fmt.Println(err.Error())
		}
		var fieldNullable bool
		for {
			nullable := prompt.Input("  Le champ "+fieldName+" peut-il est null ? ", fieldCompleter,
				prompt.OptionPrefixTextColor(prompt.Yellow))
			if nullable != "o" && nullable != "O" && nullable != "n" && nullable != "N" && nullable != "oui" && nullable != "non" {
				fmt.Println("  Oui ou non attendu")
			} else {
				fieldNullable = nullable == "o" || nullable == "O" || nullable == "oui"
				break
			}
		}
		fields = append(fields, Field{Name: fieldName, SQLType: fieldType, Nullable: fieldNullable})
	}
	if len(fields) == 0 {
		return nil, errors.New("Aucun champ dans la table, arrêt de be_factory")
	}
	table.Fields = fields
	fillFields(table)
	return table, nil
}

func fillFields(t *Table) {
	for i, f := range t.Fields {
		goName := capitalizeFirst(strings.TrimSpace(f.Name))
		sQLName := toSQL(goName)
		var goType string
		switch f.SQLType {
		case "ID":
			goType = "int64"
		case "bigint":
			if f.Nullable {
				goType = "NullInt64"
			} else {
				goType = "int64"
			}
		case "int":
			if f.Nullable {
				goType = "NullInt64"
			} else {
				goType = "int64"
			}
		case "boolean":
			if f.Nullable {
				goType = "NullBool"
			} else {
				goType = "bool"
			}
		case "text":
			if f.Nullable {
				goType = "NullString"
			} else {
				goType = "string"
			}
		case "double precision":
			if f.Nullable {
				goType = "NullFloat64"
			} else {
				goType = "float64"
			}
		case "date":
			if f.Nullable {
				goType = "NullTime"
			} else {
				goType = "time.Time"
			}
		}
		if goType == "" {
			if f.Nullable {
				goType = "NullString"
			} else {
				goType = "string"
			}
		}
		t.Fields[i].GoName = goName
		t.Fields[i].GoType = goType
		t.Fields[i].SQLName = sQLName
	}
}

func createModel(t *Table) error {
	if _, err := os.Stat("./models"); os.IsNotExist(err) {
		if err = os.Mkdir("./models", os.ModeDir|0777); err != nil {
			return errors.New("Dossier models inexistant, impossible de le créer")
		}
	}
	file, err := os.Create("./models/" + t.SQLName + ".go")
	if err != nil {
		return errors.New("Création du fichier models " + err.Error())
	}
	defer file.Close()
	varName := string(t.Name[0] + 'a' - 'A')
	content := `package models
import (
	"database/sql"
	"errors"
)

// ` + t.Name + ` model
type ` + t.Name + ` struct {
`
	for _, f := range t.Fields {
		content += "  " + f.GoName + ` ` + f.GoType + " `json:\"" + f.GoName + "\"`\n"
	}
	content += `}

// ` + t.Name + `s embeddes an array of ` + t.Name + ` for json export
type ` + t.Name + `s struct {
	` + t.Name + `s []` + t.Name + " `json:\"" + t.Name + "\"`" +
		`}

// Validate checks if ` + t.Name + `'s fields are correctly filled
func (` + varName + ` *` + t.Name + `) Validate() error {
  // TODO: code validation rules
  //	if ` + varName + `. == "" {
  //		return errors.New("Champ incorrect")
  //	}
	return nil
}

// Create insert a new ` + t.Name + ` into database
func (` + varName + ` *` + t.Name + `) Create(db *sql.DB) (err error) {
	err = db.QueryRow(` + "`INSERT INTO " + t.SQLName + " ("
	var fieldNames []string
	var dollarVars []string
	var scanVars []string
	for i, f := range t.Fields {
		if f.GoName != "ID" {
			fieldNames = append(fieldNames, f.SQLName)
			dollarVars = append(dollarVars, "$"+strconv.Itoa(i+1))
			scanVars = append(scanVars, "&"+varName+"."+f.GoName)
		}
	}

	content += strings.Join(fieldNames, ",") + ")\n VALUES(" +
		strings.Join(dollarVars, ",") + ") RETURNING id`," +
		strings.Join(scanVars, ",") + `).Scan(&` + varName + `.ID)
	return err
}

// Update modifies a ` + t.SQLName + ` in database
func (` + varName + ` *` + t.Name + `) Update(db *sql.DB) (err error) {
	res, err := db.Exec(` + "`UPDATE " + t.SQLName + " SET "
	fieldNames = nil
	scanVars = nil
	i := 1
	for _, f := range t.Fields {
		if f.GoName != "ID" {
			fieldNames = append(fieldNames, f.SQLName+"=$"+strconv.Itoa(i))
			scanVars = append(scanVars, varName+"."+f.GoName)
			i++
		}
	}
	content += strings.Join(fieldNames, ",") + " WHERE id=$" + strconv.Itoa(i) +
		"`,\n" + strings.Join(scanVars, ",") + "," + varName + ".ID)\n" +
		`  if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("` + t.FrenchName + ` introuvable")
	}
	return err
}

// GetAll fetches all ` + t.Name + `s from database
func (` + varName + ` *` + t.Name + `s) GetAll(db *sql.DB) (err error) {
	rows, err := db.Query(` + "`SELECT "
	fieldNames = nil
	for _, f := range t.Fields {
		fieldNames = append(fieldNames, f.SQLName)
	}
	content += strings.Join(fieldNames, ",") + " FROM " + t.SQLName + "`)\n"
	content += `  if err != nil {
		return err
	}
	var row ` + t.Name + `
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(`
	fieldNames = nil
	for _, f := range t.Fields {
		fieldNames = append(fieldNames, "&row."+f.GoName)
	}
	content += strings.Join(fieldNames, ",") + `); err != nil {
			return err
		}
		` + varName + `.` + t.Name + `s = append(` + varName + `.` + t.Name + `s, row)
	}
	err = rows.Err()
	return err
}

// Delete removes ` + t.SQLName + ` whose ID is given from database
func (` + varName + ` *` + t.Name + `) Delete(db *sql.DB) (err error) {
	res, err := db.Exec("DELETE FROM ` + t.SQLName + ` WHERE id = $1", ` + varName + `.ID)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("` + t.FrenchName + ` introuvable")
	}
	return nil
}
`
	_, err = file.WriteString(content)
	if err != nil {
		return errors.New("Erreur d'écriture du fichier model : " + err.Error())
	}
	return nil
}

func createAction(t *Table) error {
	if _, err := os.Stat("./actions"); os.IsNotExist(err) {
		if err = os.Mkdir("./actions", os.ModeDir|0777); err != nil {
			return errors.New("Dossier actions inexistant, impossible de le créer")
		}
	}
	file, err := os.Create("./actions/" + t.SQLName + ".go")
	if err != nil {
		return errors.New("Création du fichier models " + err.Error())
	}
	defer file.Close()
	srcGoPAth := os.Getenv("GOPATH")
	path, err := os.Getwd()
	if err != nil {
		return errors.New("Impossible de récupérer le chemin courant " + err.Error())
	}
	importPath := strings.Replace(path[len(srcGoPAth)+5:], "\\", "/", -1) + "/models"
	content := `package actions

import (
	"database/sql"
	"net/http"

	"` + importPath + `"
	"github.com/kataras/iris"
)

type ` + t.SQLName + `Req struct {
	` + t.Name + ` models.` + t.Name + ` ` + "`json:\"" + t.Name + "\"`" + `
}

// Create` + t.Name + ` handles the post request to create a new ` + t.SQLName + `
func Create` + t.Name + `(ctx iris.Context) {
	var req ` + t.SQLName + `Req
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ` + lowerFirst(t.FrenchName) + `, décodage : " + err.Error()})
		return
	}
	if err := req.` + t.Name + `.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Création de ` + lowerFirst(t.FrenchName) + ` : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.` + t.Name + `.Create(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Création de ` + lowerFirst(t.FrenchName) + `, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(req)
}

// Update` + t.Name + ` handles the put request to modify a new ` + t.SQLName + `
func Update` + t.Name + `(ctx iris.Context) {
	var req ` + t.SQLName + `Req
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ` + lowerFirst(t.FrenchName) + `, décodage : " + err.Error()})
		return
	}
	if err := req.` + t.Name + `.Validate(); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonError{"Modification de ` + lowerFirst(t.FrenchName) + ` : " + err.Error()})
		return
	}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := req.` + t.Name + `.Update(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Modification de ` + lowerFirst(t.FrenchName) + `, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(req)
}

// Get` + t.Name + `s handles the get request to fetch all ` + t.SQLName + `s
func Get` + t.Name + `s(ctx iris.Context) {
	var resp models.` + t.Name + `s
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.GetAll(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Liste des ` + lowerFirst(t.FrenchName) + `s, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(resp)
}

// Delete` + t.Name + ` handles the get request to fetch all ` + t.SQLName + `s
func Delete` + t.Name + `(ctx iris.Context) {
	ID, err := ctx.Params().GetInt64("ID")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ` + lowerFirst(t.FrenchName) + `, paramètre : " + err.Error()})
		return
	}
	resp := models.` + t.Name + `{ID: ID}
	db := ctx.Values().Get("db").(*sql.DB)
	if err := resp.Delete(db); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonError{"Suppression de ` + lowerFirst(t.FrenchName) + `, requête : " + err.Error()})
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(jsonMessage{"Logement supprimé"})
}`

	_, err = file.WriteString(content)
	if err != nil {
		return errors.New("Erreur d'écriture du fichier action : " + err.Error())
	}
	return nil
}

func mockField(goType string, modified bool) string {
	switch goType {
	case "int":
		if modified {
			return "200"
		}
		return "100"
	case "int64":
		if modified {
			return "2000000"
		}
		return "1000000"
	case "string":
		if modified {
			return "\"Essai2\""
		}
		return "\"Essai\""
	case "bool":
		if modified {
			return "false"
		}
		return "true"
	case "float64":
		if modified {
			return "2.4"
		}
		return "1.5"
	case "time.Time":
		if modified {
			return "2018-04-02T00:00:00"
		}
		return "2019-03-01T00:00:00"
	case "NullInt64":
		if modified {
			return "null"
		}
		return "1000000"
	case "NullString":
		if modified {
			return "null"
		}
		return "\"Essai\""
	case "NullBool":
		if modified {
			return "null"
		}
		return "true"
	case "NullFloat64":
		if modified {
			return "null"
		}
		return "1.5"
	case "NullTime":
		if modified {
			return "null"
		}
		return "2019-03-01T00:00:00"
	}
	if modified {
		return "\"Essai2\""
	}
	return "\"Essai\""
}

func generateJSONFields(t *Table, modified bool) (json string) {
	var fields []string
	for _, f := range t.Fields {
		if f.GoName != "ID" {
			fields = append(fields, "\""+f.GoName+"\":"+mockField(f.GoType, modified))
		}
	}
	return strings.Join(fields, ",")
}

func createTest(t *Table) error {
	file, err := os.Create("./actions/" + t.SQLName + "_test.go")
	if err != nil {
		return errors.New("Création du fichier test " + err.Error())
	}
	defer file.Close()
	jsonFields := generateJSONFields(t, false)
	jsonFields2 := generateJSONFields(t, true)
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
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{` + jsonFields2 + `}}` + "`" + `),
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
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{"ID":0,` + jsonFields2 + `}}` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `Modification de ` + lowerFirst(t.FrenchName) + `, requête : ` + "`" + `},
				StatusCode:   http.StatusInternalServerError}, // 3 : bad ID
			{Sent: []byte(` + "`" + `{"` + t.Name + `":{"ID":` + "`" + ` + strconv.Itoa(ID) + ` + "`" + `,` + jsonFields2 + `}}` + "`" + `),
				Token:        c.Config.Users.Admin.Token,
				RespContains: []string{` + "`" + `"` + t.Name + `":{"ID":` + "`" + ` + strconv.Itoa(ID) + ` + "`" + `,` + jsonFields2 + `}` + "`" + `},
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
				RespContains: []string{` + "`" + `{"` + t.Name + `":[{"ID":1,` + jsonFields2 + `}]}` + "`" + `},
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

func addRoutes(t *Table) error {
	addRouteContent, err := ioutil.ReadFile("./actions/routes.go")
	if err != nil {
		return errors.New("Lecture du fichier routes " + err.Error())
	}
	idx1 := strings.Index(string(addRouteContent), "userParty :=")
	if idx1 == -1 {
		return errors.New("Impossible de trouver userParty")
	}
	idx2 := strings.Index(string(addRouteContent[idx1:]), "}")
	if idx2 == -1 {
		return errors.New("Impossible de trouver la fin de userParty")
	}

	return ioutil.WriteFile("./actions/routes.go",
		[]byte(
			string(addRouteContent[0:idx1-2])+`  adminParty.Post("/`+
				toSQL(t.Name)+`", Create`+t.Name+`)
	adminParty.Put("/`+toSQL(t.Name)+`", Update`+t.Name+`)
	adminParty.Delete("/`+toSQL(t.Name)+`/{ID}", Delete`+t.Name+")\n\n  "+
				string(addRouteContent[idx1:idx1+idx2])+
				"\n\tuserParty.Get(\"/"+toSQL(t.Name)+"s\", Get"+t.Name+"s)\n"+
				string(addRouteContent[idx1+idx2:])), 0666)
}

func sqlFieldsCreate(f *Field) (str string) {
	if f.SQLType == "ID" {
		return "    " + f.SQLName + " SERIAL PRIMARY KEY"
	}
	str = "    " + f.SQLName + " " + f.SQLType
	if !f.Nullable {
		str += " NOT NULL"
	}
	return str
}

func createTestTable(t *Table) error {
	commonsTestsContent, err := ioutil.ReadFile("./actions/commons_test.go")
	if err != nil {
		return errors.New("Impossible de lire commons_test.go" + err.Error())
	}
	idx0 := strings.Index(string(commonsTestsContent), "`DROP TABLE IF EXISTS")
	if idx0 == -1 {
		return errors.New("Impossible de trouver la suppression des tables de test")
	}
	idx1 := strings.Index(string(commonsTestsContent[idx0:]), " `);")
	if idx1 == -1 {
		return errors.New("Impossible de trouver la fun de suppression des tables de test")
	}
	idx2 := strings.Index(string(commonsTestsContent), "queries := []string{")
	if idx2 == -1 {
		return errors.New("Impossible de trouver les requêtes de création de table test")
	}
	idx3 := strings.Index(string(commonsTestsContent[idx2:]), "	}")
	if idx3 == -1 {
		return errors.New("Impossible de trouver la fin des requête dans commons_test")
	}
	count := strings.Count(string(commonsTestsContent[idx2:idx2+idx3]), "`CREATE TABLE")
	var fields []string
	for _, f := range t.Fields {
		fields = append(fields, sqlFieldsCreate(&f))
	}
	return ioutil.WriteFile("./actions/commons_test.go",
		[]byte(string(commonsTestsContent[0:idx0+idx1])+", "+t.SQLName+
			string(commonsTestsContent[idx0+idx1:idx2+idx3])+"\t\t`CREATE table "+t.SQLName+
			" (\n\t"+strings.Join(fields, ",\n\t")+"\n\t\t);`, // "+strconv.Itoa(count)+
			" : "+t.SQLName+"\n"+string(commonsTestsContent[idx2+idx3:])), 0666)
}

func main() {
	table, err := getFields()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	funcs := []func(*Table) error{createModel, createAction, createTest,
		addRoutes, createTestTable}
	for _, f := range funcs {
		err = f(table)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
	}
}
