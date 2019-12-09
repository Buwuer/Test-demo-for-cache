package datainit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"webcache/db"
)

func Query(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	stmtOut, err := db.MysqlClient.Prepare("SELECT * FROM Member WHERE id = ?")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var data struct {
		id   int
		name string
	}
	err = stmtOut.QueryRow(id).Scan(&data.id, &data.name)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("The member name is: %s", data.name)

	c.JSON(200, gin.H{
		"name": data.name,
		"ID":   data.id,
	})
}

func Set(c *gin.Context) {
	stmtIns, err := db.MysqlClient.Prepare("INSERT INTO Member (name) VALUES( ? )") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()

	// Insert square numbers in the database
	for i := 1; i < 501; i++ {
		_, err = stmtIns.Exec("name00" + strconv.Itoa(i))
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
	}
}
