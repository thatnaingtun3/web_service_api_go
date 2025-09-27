package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	err := a.Initialise(DBUser, DBPassword, "tests")
	if err != nil {
		log.Fatal("error")
	}
	createTable()
	m.Run()
}

func createTable() {
	createTableQuery := `CREATE TABLE IF NOT EXISTS products (
           id int NOT NULL AUTO_INCREMENT,
           name varchar(255) NOT NULL,
           quantity int,
           price float(10,7),
           PRIMARY KEY (id));`
	_, err := a.DB.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE from products")
	a.DB.Exec("ALTER TABLE products AUTO_INCREMENT = 1")

	log.Println("createTable")
}
func addProduct(name string, quantity int, price float64) {

	query := fmt.Sprintf("INSERT into products (name,quantity,price) VALUES('%v',%v,%v)", name, quantity, price)
	_, err := a.DB.Exec(query)
	if err != nil {
		log.Println(err)
	}
}
func TestGetProduct(t *testing.T) {
	clearTable()
	addProduct("keyboard", 1, 500)
	request, _ := http.NewRequest("GET", "/product/1", nil)
	response := sendRequest(request)
	checkStatusCode(t, http.StatusOK, response.Code)
}
func checkStatusCode(t *testing.T, expectedStatusCode int, actualStatusCode int) {
	if expectedStatusCode != actualStatusCode {
		t.Errorf("Expected status: %v, Received: %v", expectedStatusCode, actualStatusCode)
	}
}
func sendRequest(request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, request)
	return recorder
}

func TestCreateProduct(t *testing.T) {
	clearTable()
	body := []byte(`{"name":"chair","quantity":2,"price":500}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := sendRequest(req)
	checkStatusCode(t, http.StatusCreated, resp.Code)

	var m map[string]interface{}
	_ = json.Unmarshal(resp.Body.Bytes(), &m)

	if m["name"] != "chair" {
		t.Errorf("Expected name: %v, Got: %v", "chair", m["name"])
	}
	if m["quantity"] != 2.0 { // numbers unmarshal as float64
		t.Errorf("Expected quantity: %v, Got: %v", 2.0, m["quantity"])
	}
}

// write more tests for update and delete
func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProduct("table", 1, 700)
	body := []byte(`{"name":"table","quantity":3,"price":700}`)
	req, _ := http.NewRequest("PUT", "/product/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := sendRequest(req)
	checkStatusCode(t, http.StatusOK, resp.Code)
	var m map[string]interface{}
	_ = json.Unmarshal(resp.Body.Bytes(), &m)

	if m["quantity"] != 3.0 {
		t.Errorf("Expected quantity: %v, Got: %v", 3.0, m["quantity"])
	}
}
func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProduct("mouse", 1, 300)
	req, _ := http.NewRequest("DELETE", "/product/1", nil)

	resp := sendRequest(req)
	checkStatusCode(t, http.StatusOK, resp.Code)

	// Try to get the deleted product
	req, _ = http.NewRequest("GET", "/product/1", nil)
	resp = sendRequest(req)
	checkStatusCode(t, http.StatusNotFound, resp.Code)
}
