package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
)

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
	Batch      bool
	Fields     []Field
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

func getFields() (*Table, error) {
	table := &Table{}
	var name, frenchName, fieldName, fieldType, length string
	var fieldNullable bool
	intValidator := func(val interface{}) error {
		str, ok := val.(string)
		if !ok {
			return errors.New("Chiffre attendu")
		}
		if _, err := strconv.Atoi(str); err != nil {
			return errors.New("Chiffre attendu")
		}
		return nil
	}
	prompt := &survey.Input{
		Message: "Nom du modèle",
	}
	survey.AskOne(prompt, &name, nil)
	if name == "" {
		return nil, errors.New("Impossible d'avoir le nom du modèle")
	}
	table.Name = capitalizeFirst(strings.TrimSpace(name))
	table.SQLName = toSQL(table.Name)
	prompt = &survey.Input{
		Message: "Nom du modèle français",
	}
	survey.AskOne(prompt, &frenchName, nil)
	if frenchName == "" {
		return nil, errors.New("Impossible d'avoir le nom français")
	}
	table.FrenchName = capitalizeFirst(strings.TrimSpace(frenchName))
	var fields []Field
	for i := 1; ; i++ {
		prompt = &survey.Input{Message: "Nom du champ n°" + strconv.Itoa(i)}
		survey.AskOne(prompt, &fieldName, nil)
		if fieldName == "" {
			break
		}
		if fieldName == "ID" && i == 1 {
			fields = append(fields, Field{Name: "ID", SQLType: "ID", Nullable: false})
			continue

		}
		fieldTypePrompt := &survey.Select{
			Message: "Type du champ" + strconv.Itoa(i),
			Options: []string{"ID", "bigint", "int", "boolean", "varchar", "text", "double precision", "date"},
		}
		survey.AskOne(fieldTypePrompt, &fieldType, nil)
		if fieldType == "varchar" {
			varcharLengthPrompt := &survey.Input{Message: "Longueur maximale"}
			survey.AskOne(varcharLengthPrompt, &length, intValidator)
			fieldType = fieldType + "(" + length + ")"
		}
		fieldNullablePrompt := &survey.Confirm{
			Message: "Champ annulable ?",
		}
		survey.AskOne(fieldNullablePrompt, &fieldNullable, nil)
		fields = append(fields, Field{Name: fieldName, SQLType: fieldType, Nullable: fieldNullable})
	}
	if len(fields) == 0 {
		return nil, errors.New("Aucun champ dans la table")
	}
	batchPrompt := &survey.Confirm{Message: "Inclure un import de batch"}
	survey.AskOne(batchPrompt, &table.Batch, nil)
	table.Fields = fields
	table.fillFields()
	return table, nil
}

func (t *Table) fillFields() {
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

func (t *Table) generateJSONFields(modified bool) (json string) {
	var fields []string
	for _, f := range t.Fields {
		if f.GoName != "ID" {
			fields = append(fields, "\""+f.GoName+"\":"+mockField(f.GoType, modified))
		}
	}
	return strings.Join(fields, ",")
}

func (f *Field) sqlFieldsCreate() (str string) {
	if f.SQLType == "ID" {
		return "    " + f.SQLName + " SERIAL PRIMARY KEY"
	}
	str = "    " + f.SQLName + " " + f.SQLType
	if !f.Nullable {
		str += " NOT NULL"
	}
	return str
}
