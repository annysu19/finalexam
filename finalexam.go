package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func authMiddleware(c *gin.Context) {
	fmt.Println("This is a middleware")
	token := c.GetHeader("Authorization")
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized."})
		c.Abort()
		return
	}
	c.Next()

}

type MyApp struct {
	DB *sql.DB
}

type Customers struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var customer []Customers

func (app MyApp) InsertCustomerHandler(c *gin.Context) {
	var cs Customers

	err := c.ShouldBindJSON(&cs)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	customer = append(customer, cs)

	row := app.DB.QueryRow("INSERT INTO customer (name, email,status) values ($1, $2,$3) RETURNING id,name,email,status", cs.Name, cs.Email, cs.Status)

	csr := Customers{}
	err = row.Scan(&csr.ID, &csr.Name, &csr.Email, &csr.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusCreated, csr)
}

func (app MyApp) getCustomerHandler(c *gin.Context) {

	idx := c.Param("id")
	ii, _ := strconv.Atoi(idx)
	row, err := app.DB.Query("SELECT id, name,email, status FROM customer WHERE id=$1", ii)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}

	row.Next()

	csr := Customers{}
	err = row.Scan(&csr.ID, &csr.Name, &csr.Email, &csr.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, csr)

}

func (app MyApp) getAllCustomerHandler(c *gin.Context) {

	stmt, err := app.DB.Prepare("SELECT id, name,email, status FROM customer")
	if err != nil {
		log.Print("Can not prepare statement select customer", err)
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Print("Can not query all customer", err)
	}

	var customer []Customers
	for rows.Next() {
		var id int
		var name, email, status string
		err := rows.Scan(&id, &name, &email, &status)
		if err != nil {
			log.Print("Can not scan row into variable", err)
		}
		customer = append(customer, Customers{id, name, email, status})

	}
	c.JSON(http.StatusOK, customer)

}

func (app MyApp) updateCustomerHandler(c *gin.Context) {

	idd, _ := strconv.Atoi(c.Param("id"))
	var cs Customers

	err := c.ShouldBindJSON(&cs)

	row := app.DB.QueryRow("UPDATE customer SET name=$2, email=$3,status=$4 WHERE id=$1 RETURNING id,name,email,status", idd, cs.Name, cs.Email, cs.Status)
	csr := Customers{}
	err = row.Scan(&csr.ID, &csr.Name, &csr.Email, &csr.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		log.Print("Can not update data", err)
		return
	}

	c.JSON(http.StatusOK, csr)
}

func (app MyApp) deleteCustomerHandler(c *gin.Context) {

	idd, _ := strconv.Atoi(c.Param("id"))

	stmt, err := app.DB.Prepare("DELETE from customer WHERE id=$1")
	if err != nil {
		log.Print("can't prepare statment delete", err)
	}
	if _, err := stmt.Exec(idd); err != nil {
		log.Print("error execute delete ", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})

}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Can not connect database", err)
	}
	defer func() {
		db.Close()
	}()

	a := MyApp{db}
	r := gin.Default()

	r.Use(authMiddleware)

	createTb := `
	CREATE TABLE IF NOT EXISTS customer (
	id SERIAL PRIMARY KEY,
	name TEXT,
	email TEXT,
	status TEXT
	);
	`
	_, err = db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table", err)
	}
	r.POST("/customers", a.InsertCustomerHandler)
	r.GET("/customers/:id", a.getCustomerHandler)
	r.GET("/customers", a.getAllCustomerHandler)
	r.PUT("/customers/:id", a.updateCustomerHandler)
	r.DELETE("/customers/:id", a.deleteCustomerHandler)
	r.Run(":2019")

}
