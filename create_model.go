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
	var fieldNames []string
	var setFields []string
	var dollarVars []string
	var scanVars []string
	var stmtVars []string
	i := 1
	for _, f := range t.Fields {
		if f.GoName != "ID" {
			fieldNames = append(fieldNames, f.SQLName)
			setFields = append(setFields, f.SQLName+"=t."+f.SQLName)
			dollarVars = append(dollarVars, "$"+strconv.Itoa(i))
			scanVars = append(scanVars, "&"+varName+"."+f.GoName)
			stmtVars = append(stmtVars, "r."+f.GoName)
			i++
		}
	}
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
	` + t.Name + `s []` + t.Name + " `json:\"" + t.Name + "\"`\n}"
	if t.Batch {
		content += "\n// " + t.Name + `Line is used to decode a line of ` + t.Name + ` batch
type ` + t.Name + `Line struct {
`
		for _, f := range t.Fields {
			if f.Name != "ID" {
				content += "  " + f.GoName + ` ` + f.GoType + " `json:\"" + f.GoName + "\"`\n"
			}
		}
		content += `}

// ` + t.Name + `Batch embeddes an array of ` + t.Name + `Line for json export
type ` + t.Name + `Batch struct {
	Lines []` + t.Name + `Line ` + "`json:\"" + t.Name + "\"`\n}\n"
	}
	content += `
// Validate checks if ` + t.Name + `'s fields are correctly filled
func (` + varName + ` *` + t.Name + `) Validate() error {
  // TODO: code validation rules
  //	if ` + varName + `. == "" {
  //		return errors.New("Champ incorrect")
  //	}
	return nil
}
`
	if t.Create {
		content += "// Create insert a new " + t.Name + " into database\n" +
			"func (" + varName + " *" + t.Name + ") Create(db *sql.DB) (err error) {\n" +
			"\terr = db.QueryRow(`INSERT INTO " + t.SQLName + " (" +
			strings.Join(fieldNames, ",") + ")\n VALUES(" +
			strings.Join(dollarVars, ",") + ") RETURNING id`," +
			strings.Join(scanVars, ",") + `).Scan(&` + varName + ".ID)\n\treturn err\n}\n"
	}
	if t.Get {
		fieldNames = nil
		for _, f := range t.Fields {
			fieldNames = append(fieldNames, f.SQLName)
		}
		content += `// Get fetches a ` + t.Name + ` from database` +
			"func (" + varName + " *" + t.Name + ") Get  (db*sql.DB) (err error) {\n" +
			"\terr = db.QueryRow(`GET " + strings.Join(fieldNames, ", ") + " FROM " +
			t.SQLName + " WHERE ID=$1`, " + varName + ".ID).Scan(" + strings.Join(scanVars, ", ") +
			")\n\tif err!=nil{\n\t\t return err\n\t}\n\treturn nil\n}\n"
	}
	if t.Update {
		fieldNames = nil
		scanVars = nil
		i = 1
		for _, f := range t.Fields {
			if f.GoName != "ID" {
				fieldNames = append(fieldNames, f.SQLName+"=$"+strconv.Itoa(i))
				scanVars = append(scanVars, varName+"."+f.GoName)
				i++
			}
		}
		content += "// Update modifies a " + t.SQLName + " in database\n" +
			"func (" + varName + " *" + t.Name + ") Update(db *sql.DB) (err error) {\n" +
			"\tres, err := db.Exec(`UPDATE " + t.SQLName + " SET " +
			strings.Join(fieldNames, ",") + " WHERE id=$" + strconv.Itoa(i) + "`,\n" +
			strings.Join(scanVars, ",") + "," + varName + ".ID)\n\tif err != nil {\n" +
			"\t\treturn err\n\t}\tcount, err := res.RowsAffected()\n\tif err != nil {\n" +
			"\t\treturn err\n\t}\n\tif count != 1 {\n\t\treturn errors.New(\"" +
			t.FrenchName + " introuvable\")\n\t}\n\treturn err\n}\n"
	}
	if t.GetAll {
		content += `// GetAll fetches all ` + t.Name + `s from database
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
`
	}
	if t.Delete {
		content += `// Delete removes ` + t.SQLName + ` whose ID is given from database
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
	}

	if t.Batch {
		fieldNames = nil
		dollarVars = nil
		i = 1
		for _, f := range t.Fields {
			if f.GoName != "ID" {
				fieldNames = append(fieldNames, f.SQLName)
				dollarVars = append(dollarVars, "$"+strconv.Itoa(i))
				i++
			}
		}
		content += `
	// Save insert a batch of ` + t.Name + `Line into database
func (` + varName + ` *` + t.Name + `Batch) Save(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(` + "`INSERT INTO temp_" + t.SQLName + " (" + strings.Join(fieldNames, ",") +
			") VALUES (" + strings.Join(dollarVars, ",") + ")`)" + `
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, r := range ` + varName + `.Lines {
		// TODO : fields validation
		// if r. {
		//	tx.Rollback()
		//	return errors.New("Champs incorrects")
		//}
		if _, err = stmt.Exec(` + strings.Join(stmtVars, ",") + `); err != nil {
			tx.Rollback()
			return err
		}
	}
	_, err = tx.Exec(` + "`UPDATE " + t.SQLName + " SET " + strings.Join(setFields, ",") +
			" FROM temp_" + t.SQLName + " t WHERE t. = " + t.SQLName + ".`) // TODO add reference fields for updating\n" +
			`if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(` + "`INSERT INTO " + t.SQLName + " (" + strings.Join(fieldNames, ",") + `)
	SELECT ` + strings.Join(fieldNames, ",") + ` from temp_` + t.SQLName + ` 
	  WHERE  NOT IN (SELECT  from ` + t.SQLName + ")`)// TODO : add reference fields for inserting\n" + `
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}`
	}
	_, err = file.WriteString(content)
	if err != nil {
		return errors.New("Erreur d'écriture du fichier model : " + err.Error())
	}
	return nil
}
