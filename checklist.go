package main

import "github.com/gin-gonic/gin"
import "time"
import "log"
import "strconv"
import "database/sql"
import _ "github.com/lib/pq"
import "gopkg.in/gorp.v1"

type DatabaseGeneric struct {
	Id       int64     `db:"id" json:"id"`
	Name     string    `db:"name" json:"name"`
	Created  time.Time `db:"created" json:"created"`
	Modified time.Time `db:"modified" json:"modified"`
}

type Checklist struct {
	DatabaseGeneric
}

type ChecklistItem struct {
	Completed bool `db:"completed" json:"completed"`
}

var dbmap = initDb()

func initDb() *gorp.DbMap {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:32776?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTableWithName(Checklist{}, "checklist").SetKeys(true, "Id")
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
	return dbmap
}

func main() {
	r := gin.Default()
	v1 := r.Group("api/v1")
	{
		v1.GET("/checklist", GetChecklists)
		v1.GET("/checklist/:id", GetChecklist)
		v1.POST("/checklist", PostChecklist)
		v1.PUT("/checklist/:id", UpdateChecklist)
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
	id := c.Params.ByName("id")
	user_id, err := strconv.ParseInt(id, 0, 64)
	var checklists []Checklist
	if err != nil {
		_, err = dbmap.Select(&checklists, "select * from checklist where name=$1", id)
		checkErr(err, "select failed")
	} else {
		_, err = dbmap.Select(&checklists, "select * from checklist where id=$1", user_id)
		checkErr(err, "select failed")
	}
	c.JSON(200, checklists)
	// curl -i http://localhost:8080/api/v1/checklist/1
}
func PostChecklist(c *gin.Context) {
	var json Checklist
	c.Bind(&json)
	checklist := createChecklist(json)
	if checklist.Name == json.Name {
		content := gin.H{
			"result": "Success",
			"name": checklist.Name,
			}
		c.JSON(201, content)
	} else {
		c.JSON(500, gin.H{"result": "An error occured"})
	}
}

func UpdateChecklist(c *gin.Context) {
	// The futur code.
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
			Name: item.Name,
			Created: time.Now(),
			Modified: time.Now(),
		},
	}
	err := dbmap.Insert(&checklist)
	checkErr(err, "create failed")
	return checklist
}

