package main

import "gitlab.svc.everyonce.com/everyonce/go-checklist/Godeps/_workspace/src/github.com/gin-gonic/gin"
import "fmt"
import "time"
import "os"
import "log"
import "strconv"
import "database/sql"
import _ "gitlab.svc.everyonce.com/everyonce/go-checklist/Godeps/_workspace/src/github.com/lib/pq"
import "gitlab.svc.everyonce.com/everyonce/go-checklist/Godeps/_workspace/src/gopkg.in/gorp.v1"

type DatabaseGeneric struct {
	Id       int64     `db:"id" json:"id"`
	Name     string    `db:"name" json:"name"`
	Created  time.Time `db:"created" json:"created"`
	Modified time.Time `db:"modified" json:"modified"`
}

type Checklist struct {
	DatabaseGeneric
	Items []ChecklistItem `db:"-"`
}

type ChecklistItem struct {
	DatabaseGeneric
	ChecklistId   int64     `db:"checklist_id" json:"checklist_id"`
	Completed     bool      `db:"completed" json:"completed"`
	CompletedDate time.Time `db:"completed_date" json:"completed_date"`
}

var dbName = os.Getenv("CHECKLIST_DB")
var dbUser = os.Getenv("CHECKLIST_USER")
var dbPass = os.Getenv("CHECKLIST_PASS")
var dbHost = os.Getenv("CHECKLIST_HOST")
var dbmap = initDb()

func main() {
	r := gin.Default()
	v1 := r.Group("api/v1")
	{
		v1.GET("/checklist", GetChecklists)
		v1.GET("/checklist/:id", GetChecklist)
		v1.POST("/checklist", PostChecklist)
		v1.POST("/checklist/:id", PostChecklistItem)
		v1.PUT("/checklist/:id", UpdateChecklist)
		v1.PUT("/checklist/:id/:itemid", UpdateChecklistItem)
		v1.DELETE("/checklist/:id", DeleteChecklist)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}

func GetChecklists(c *gin.Context) {
	var checklists []Checklist
	_, err := dbmap.Select(&checklists, "select * from checklist")
	checkErr(err, "select failed")
	c.JSON(200, checklists)
	// curl -i http://localhost:8080/api/v1/checklist
}
func GetChecklist(c *gin.Context) {
	checklist := findChecklist(c.Params.ByName("id"))
	c.JSON(200, checklist)
	// curl -i http://localhost:8080/api/v1/checklist/1
}
func PostChecklist(c *gin.Context) {
	var json Checklist
	c.Bind(&json)
	checklist := createChecklist(json)
	if checklist.Name == json.Name {
		content := gin.H{
			"result": "Success",
			"name":   checklist.Name,
		}
		c.JSON(201, content)
	} else {
		c.JSON(500, gin.H{"result": "An error occured"})
	}
}

func PostChecklistItem(c *gin.Context) {
	var json ChecklistItem
	c.Bind(&json)
	checklist := findChecklist(c.Params.ByName("id"))

	checklistItem := createChecklistItem(json, checklist)
	if checklistItem.Name == json.Name {
		content := gin.H{
			"result": "Success",
			"name":   checklistItem.Name,
		}
		c.JSON(201, content)
	} else {
		c.JSON(500, gin.H{"result": "An error occured"})
	}
}

func UpdateChecklist(c *gin.Context) {
	// The futur code.
}
func UpdateChecklistItem(c *gin.Context) {
	var json ChecklistItem
	c.Bind(&json)
	checklist := findChecklist(c.Params.ByName("id"))
	itemId, err := strconv.ParseInt(c.Params.ByName("itemid"), 0, 64)
	checkErr(err, "parsing checklist item id failed")
	checklistItem := findChecklistItem(checklist, itemId)
	if json.Completed {
		checklistItem.Completed = true
		checklistItem.CompletedDate = time.Now()
	}
	if !json.CompletedDate.IsZero() {
		checklistItem.CompletedDate = json.CompletedDate
	}
	items, err := dbmap.Update(&checklistItem)
	checkErr(err, "update problem")
	log.Println(items)
}

func DeleteChecklist(c *gin.Context) {
	// The futur code.
}
func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
func createChecklist(item Checklist) Checklist {
	checklist := Checklist{
		DatabaseGeneric: DatabaseGeneric{
			Name:     item.Name,
			Created:  time.Now(),
			Modified: time.Now(),
		},
		Items: []ChecklistItem{},
	}
	err := dbmap.Insert(&checklist)
	checkErr(err, "create failed")
	return checklist
}
func createChecklistItem(item ChecklistItem, checklist Checklist) ChecklistItem {
	checklistItem := ChecklistItem{
		ChecklistId: checklist.Id,
		DatabaseGeneric: DatabaseGeneric{
			Name:     item.Name,
			Created:  time.Now(),
			Modified: time.Now(),
		},
	}
	err := dbmap.Insert(&checklistItem)
	checkErr(err, "create failed")
	return checklistItem
}
func findChecklist(x string) Checklist {
	user_id, err := strconv.ParseInt(x, 0, 64)
	var checklist Checklist
	if err != nil {
		err = dbmap.SelectOne(&checklist, "select * from checklist where name=$1", x)
		checkErr(err, "select failed")
	} else {
		err = dbmap.SelectOne(&checklist, "select * from checklist where id=$1", user_id)
		checkErr(err, "select failed")
	}
	checklist.Items = findChecklistItems(checklist)
	return checklist
}
func findChecklistItem(parent Checklist, checklistItemId int64) ChecklistItem {
	var item ChecklistItem
	log.Println(parent.Id, checklistItemId)
	err := dbmap.SelectOne(&item, "select * from checklist_item where checklist_id=$1 and id=$2", parent.Id, checklistItemId)
	checkErr(err, "couldnt get child")
	return item
}
func findChecklistItems(parent Checklist) []ChecklistItem {
	var items []ChecklistItem
	_, err := dbmap.Select(&items, "select * from checklist_item where checklist_id=$1", parent.Id)
	checkErr(err, "couldnt get children")
	return items
}

func initDb() *gorp.DbMap {
	_ = "breakpoint"
	dbConnStr :=  fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName)
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTableWithName(Checklist{}, "checklist").SetKeys(true, "Id")
	dbmap.AddTableWithName(ChecklistItem{}, "checklist_item").SetKeys(true, "Id")
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed for str: "+dbConnStr)
	return dbmap
}
