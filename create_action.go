package main

import (
	"errors"
	"os"
	"strings"
)

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
`
	if t.Create {
		content += `// Create` + t.Name + ` handles the post request to create a new ` + t.SQLName + `
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
`
	}
	if t.Update {
		content += `// Update` + t.Name + ` handles the put request to modify a new ` + t.SQLName + `
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
`
	}
	if t.Get {
		content += "// Get" + t.Name + " handles the get request to fetch a " + t.SQLName + "\n" +
			"func Get" + t.Name + "ctx iris.Content) {\n\tvar resp " + t.Name + "Req\n" +
			"\tdb := ctx.Values().Get(\"db\").(*sql.DB)\n\tif err := resp.Get(db); err != nil {\n" +
			"\t\tctx.StatusCode(http.StatusInternalServerError)\n" +
			"\t\tctx.JSON(jsonError{\"Récupération de " + lowerFirst(t.FrenchName) +
			", requête : \" + err.Error())\n\treturn\n}\n+\tctx.StatusCode(http.StatusOK)\n" +
			"\tctx.JSON(resp)\n}\n"
	}
	if t.GetAll {
		content += `// Get` + t.Name + `s handles the get request to fetch all ` + t.SQLName + `s
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
`
	}
	if t.Delete {
		content += `// Delete` + t.Name + ` handles the get request to fetch all ` + t.SQLName + `s
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
	}
	if t.Batch {
		content += `

// Batch` + t.Name + `s handle the post request to update and insert a batch of ` + t.SQLName + `s into the database
func Batch` + t.Name + `s(ctx iris.Context) {
var b models.` + t.Name + `Batch
if err := ctx.ReadJSON(&b); err != nil {
	ctx.StatusCode(http.StatusInternalServerError)
	ctx.JSON(jsonError{"Batch de ` + t.FrenchName + `s, décodage : " + err.Error()})
	return
}
db := ctx.Values().Get("db").(*sql.DB)
if err := b.Save(db); err != nil {
	ctx.StatusCode(http.StatusInternalServerError)
	ctx.JSON(jsonError{"Batch de ` + t.FrenchName + `s, requête : " + err.Error()})
	return
}
ctx.StatusCode(http.StatusOK)
ctx.JSON(jsonMessage{"Batch de ` + t.FrenchName + `s importé"})
}

	`
	}

	_, err = file.WriteString(content)
	if err != nil {
		return errors.New("Erreur d'écriture du fichier action : " + err.Error())
	}
	return nil
}
