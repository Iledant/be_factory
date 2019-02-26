package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

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
