package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

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
	var adminContent, userContent string
	if t.Create {
		adminContent += "\n\tadminParty.Post(\"/" + toSQL(t.Name) + "\", Create" + t.Name + ")"
	}
	if t.Update {
		adminContent += "\n\tadminParty.Put(\"/" + toSQL(t.Name) + "\", Update" + t.Name + ")"
	}
	if t.Delete {
		adminContent += "\n\tadminParty.Delete(\"/" + toSQL(t.Name) + "/{ID}\", Delete" + t.Name + ")"
	}
	if t.GetAll {
		userContent += "\n\tadminParty.Get(\"/" + toSQL(t.Name) + "s\", Get" + t.Name + "s)"
	}
	if t.Batch {
		adminContent += "\n\tadminParty.Post(\"/" + t.SQLName + "s\", Batch" + t.Name + "s)"
	}
	return ioutil.WriteFile("./actions/routes.go",
		[]byte(string(addRouteContent[0:idx1-2])+adminContent+
			string(addRouteContent[idx1:idx1+idx2])+userContent+
			string(addRouteContent[idx1+idx2:])), 0666)
}

func updateCommonsTest(t *Table) error {
	commonsTestsContent, err := ioutil.ReadFile("./actions/commons_test.go")
	if err != nil {
		return errors.New("Impossible de lire commons_test.go" + err.Error())
	}
	idxTestAll := strings.Index(string(commonsTestsContent), "func TestAll(t *testing.T) {")
	if idxTestAll == -1 {
		return errors.New("Impossible de trouver la fonction TestAll")
	}
	idxEndTestAll := strings.Index(string(commonsTestsContent[idxTestAll:]), "}")
	if idxEndTestAll == -1 {
		return errors.New("Impossible de trouver la fin de la fonction TestAll")
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
	var tempFields []string
	for _, f := range t.Fields {
		fields = append(fields, f.sqlFieldsCreate())
		if f.Name != "ID" {
			tempFields = append(tempFields, f.sqlFieldsCreate())
		}
	}
	var tempSQLName, tempCreateQry string
	if t.Batch {
		tempSQLName = " , temp_" + t.SQLName
		tempCreateQry = "\t\t`CREATE TABLE temp_" + t.SQLName + " (\n\t" +
			strings.Join(tempFields, ",\n\t") + "\n\t\t);`, // " + strconv.Itoa(count+1) +
			" : temp_" + t.SQLName + "\n"
	}
	return ioutil.WriteFile("./actions/commons_test.go",
		[]byte(string(commonsTestsContent[0:idxTestAll+idxEndTestAll])+"\n\ttest"+t.Name+"(t, cfg)\n"+
			string(commonsTestsContent[idxTestAll+idxEndTestAll:idx0+idx1])+", "+t.SQLName+tempSQLName+
			string(commonsTestsContent[idx0+idx1:idx2+idx3])+"\t\t`CREATE TABLE "+t.SQLName+
			" (\n\t"+strings.Join(fields, ",\n\t")+"\n\t\t);`, // "+strconv.Itoa(count)+
			" : "+t.SQLName+"\n"+tempCreateQry+string(commonsTestsContent[idx2+idx3:])), 0666)
}

func main() {
	table, err := getFields()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	funcs := []func(*Table) error{createModel, createAction, createTest,
		addRoutes, updateCommonsTest}
	for _, f := range funcs {
		err = f(table)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
	}
}
