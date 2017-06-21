/*
A basic example of RESTful API in Go using MySQL.
Data will be accessed using HTTP CRUD operations (GET, POST, PUT & DELETE).
All data will be returned in JSON format.

Notable packages:
  Gin - HTTP web framework. Gin uses httprouter which supports variables in the routing pattern
  Gorp - Allows mapping structs to tables
  Go MySQL Driver & Go's database/sql/driver - Allows us to interface with with MySQL database

*/

package main

import (
    "database/sql"
    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
    "gopkg.in/gorp.v1"
    "log"
    "strconv"
    "os"
    "fmt"
)

// Define type and use tags to alias fields to our MySQL database column names.
type crudtest_struct struct {
    Id int64 `db:"id" json:"id"`
    Firstname string `db:"firstname" json:"firstname"`
    Lastname string `db:"lastname" json:"lastname"`
}

var dbmap = initDb()

func initDb() *gorp.DbMap {
    // Connect to cloudcake MySQL db. Return error upon failure
    db, err := sql.Open("mysql", "root:root@tcp(localhost:8889)/cloudcake")
    checkErr(err, "sql.Open failed")

    // Construct a gorp DbMap
    dbmap := &gorp.DbMap{Db: db, Dialect:gorp.MySQLDialect{"InnoDB", "UTF8"}}

    // Use Gorp to register example struct. Print error upon failure
    dbmap.AddTableWithName(crudtest_struct{}, "crudtest_struct").SetKeys(true, "Id")
    err = dbmap.CreateTablesIfNotExists()
    checkErr(err, "Create table failed")
    return dbmap
}

func checkErr(err error, msg string) {
    if err != nil {log.Fatalln(msg, err)}
}

func main() {
    router := gin.Default()
    // Grouping routes into a single group
    crudtest_group := router.Group("crudtest")
    {
        crudtest_group.GET("/api", GetUsers)
        crudtest_group.GET("/api/:whereclause", GetUser)
        crudtest_group.POST("/api", PostUser)
        crudtest_group.PUT("/api/:id", UpdateUser)
        crudtest_group.DELETE("/api/:id", DeleteUser)
    }
    router.Run(":"+os.Getenv("PORT"))
}

func GetUsers(c *gin.Context) {
    var crudtest_GetUsers []crudtest_struct
    _, err := dbmap.Select(&crudtest_GetUsers, "SELECT * FROM crudtest_struct")
    if err == nil {
      c.JSON(200, crudtest_GetUsers)
    } else {
      c.JSON(404, gin.H{"error": "not able to find in the table"})
    }
    // Example curl GET all users command
    // curl -i http://localhost:8080/crudtest/api
}

func GetUser(c *gin.Context) {
    whereclause := c.Params.ByName("whereclause")
    fmt.Printf("%s\n", whereclause)
    var crudtest_GetUser []crudtest_struct
    _, err := dbmap.Select(&crudtest_GetUser, "SELECT * FROM crudtest_struct WHERE " + whereclause)
    if err == nil {
        c.JSON(200, crudtest_GetUser)
    } else {
        c.JSON(404, gin.H{"error": "content not found"})
    }
    // Example curl GET specific user ID command
    // curl -i http://localhost:8080/crudtest/api/1
}

func PostUser(c *gin.Context) {
    var crudtest_PostUser crudtest_struct
    c.Bind(&crudtest_PostUser)
    if crudtest_PostUser.Firstname != "" && crudtest_PostUser.Lastname != "" {
        if insert, _ := dbmap.Exec(`INSERT INTO crudtest_struct (firstname, lastname) VALUES (?, ?)`, crudtest_PostUser.Firstname, crudtest_PostUser.Lastname); insert != nil {
            crudtest_PostUser_id, err := insert.LastInsertId()
            if err == nil {
                content := &crudtest_struct {
                Id: crudtest_PostUser_id,
                Firstname: crudtest_PostUser.Firstname,
                Lastname: crudtest_PostUser.Lastname,
                }
                c.JSON(201, content)
            } else {
                checkErr(err, "Insert failed")
            }
        }
    } else {
        c.JSON(422, gin.H{"error": "fields are empty"})
    }
    // Example curl POST user command
    // curl -i -X POST -H "Content-Type: application/json" -d "{ \"firstname\": \"Dennis\", \"lastname\": \"Ritchie\" }" http://localhost:8080/crudtest/api
    // curl -i -X POST -H "Content-Type: application/json" -d "{ \"firstname\": \"Rob\", \"lastname\": \"Pike\" }" http://localhost:8080/crudtest/api
    // curl -i -X POST -H "Content-Type: application/json" -d "{ \"firstname\": \"Ken\", \"Thompson\": \"Burge\" }" http://localhost:8080/crudtest/api
    // curl -i -X POST -H "Content-Type: application/json" -d "{ \"firstname\": \"Robert\", \"Griesemer\": \"Li\" }" http://localhost:8080/crudtest/api
}
func UpdateUser(c *gin.Context) {
    id := c.Params.ByName("id")
    var crudtest_UpdateUser crudtest_struct
    err := dbmap.SelectOne(&crudtest_UpdateUser, "SELECT * FROM crudtest_struct WHERE id=?", id)
    if err == nil {
        var json crudtest_struct
        c.Bind(&json)
        crudtest_UpdateUser_id, _ := strconv.ParseInt(id, 0, 64)
        crudtest_UpdateUser := crudtest_struct{
            Id: crudtest_UpdateUser_id,
            Firstname: json.Firstname,
            Lastname: json.Lastname,
    }
    if crudtest_UpdateUser.Firstname != "" && crudtest_UpdateUser.Lastname != "" {
        _, err = dbmap.Update(&crudtest_UpdateUser)
        if err == nil {
        c.JSON(200, crudtest_UpdateUser)
    } else {
        checkErr(err, "Update failed")
    }
    } else {
        c.JSON(422, gin.H{"error": "fields are empty"})
    }
    } else {
        c.JSON(404, gin.H{"error": "content not found"})
    }
    // Example curl PUT user command
    // curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"John\", \"lastname\": \"Doe\" }" http://localhost:8080/crudtest/api/1
    // curl -i -X PUT -H "Content-Type: application/json" -d "{ \"firstname\": \"Jane\", \"lastname\": \"Doe\" }" http://localhost:8080/crudtest/api/2
}

func DeleteUser(c *gin.Context) {
    id := c.Params.ByName("id")
    var crudtest_DeleteUser crudtest_struct
    err := dbmap.SelectOne(&crudtest_DeleteUser, "SELECT id FROM crudtest_struct WHERE id=?", id)
    if err == nil {
        _, err = dbmap.Delete(&crudtest_DeleteUser)
        if err == nil {
            c.JSON(200, gin.H{"id #" + id: " deleted"})
        } else {
            checkErr(err, "Delete failed")
        }
    } else {
        c.JSON(404, gin.H{"error": "content not found"})
    }
    // Example curl DELETE user by ID command
    // curl -i -X DELETE http://localhost:8080/crudtest/api/1
}
