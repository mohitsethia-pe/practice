package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// type users struct {
// 	UserId       int
// 	UserName     string
// 	UserEmail    string
// 	UserPassword string
// }

// type roles struct {
// 	UserId int
// 	Roles  string
// }

func signupPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "signup.html")
		return
	}

	UserName := string(req.FormValue("username"))
	UserPassword := string(req.FormValue("password"))
	UserEmail := string(req.FormValue("email"))
	var UserID string

	err := DB.QueryRow("Select UserId from users where UserName=? or UserEmail=?", UserName, UserEmail).Scan(&UserID)

	switch {
	case err == sql.ErrNoRows:
		_, err = DB.Exec("Insert into users(UserName, UserEmail, UserPassword) VALUES(?, ?, ?)", UserName, UserEmail, UserPassword)
		if err != nil {
			http.Error(res, "Server error, unable to create your account.", 500)
			return
		}
		var role string
		var UserId int
		// query := "SELECT UserId from users where UserName=" + UserName + ";"
		err := DB.QueryRow("SELECT UserId from users where UserName=?", UserName).Scan(&UserId)
		if err != nil {
			panic(err.Error())
		}

		role = "Member"
		_, err = DB.Exec("Insert into roles(UserId, Roles) Values(?, ?)", UserId, role)
		if err != nil {
			_, err = DB.Exec("Delete from users where UserId=?", UserId)
			if err != nil {
				panic(err.Error())
			}
			http.Error(res, "Server error, unable to create role for you.", 500)
			return
		}
		res.Write([]byte("User Created!"))
		return
	case err != nil:
		http.Error(res, "Server error, unable to create your account.", 500)
		return
	default:
		http.Redirect(res, req, "/", 301)
	}
}

func loginPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "login.html")
		return
	}

	UserName := req.FormValue("username")
	UserPassword := req.FormValue("password")

	var dbUserName string
	var dbUserPassword string
	err := DB.QueryRow("SELECT UserName, UserPassword from UserDB.users where UserName=? and UserPassword=?", UserName, UserPassword).Scan(&dbUserName, &dbUserPassword)

	if err != nil {
		http.Redirect(res, req, "/login", 301)
		return
	}
	// res.Write([]byte("Hello " + dbUserName))
	var UserId string
	err = DB.QueryRow("SELECT UserId from UserDB.users where UserName=?", UserName).Scan(&UserId)

	if err != nil {
		http.Error(res, "User not found!", 301)
		return
	}

	var role string
	err = DB.QueryRow("SELECT roles from UserDB.roles where UserId=?", UserId).Scan(&role)

	if err != nil {
		http.Error(res, "can't find roles!", 301)
		return
	}

	if role == "Admin" {
		http.ServeFile(res, req, "Admin.html")
		return
	} else {
		http.ServeFile(res, req, "Member.html")
	}
}

func homePage(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "index.html")
}

func AddUser(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "Admin.html")
		return
	}
	UserName := req.FormValue("username")

	var UserId int
	query := "Select UserId from UserDB.users where UserName=" + UserName + ";"
	err := DB.QueryRow(query).Scan(&UserId)
	if err != nil {
		res.Write([]byte("User not found"))
		return
	}
	role := "Admin"
	// query1 := "UPDATE UserDB.roles SET Roles=" + role + " WHERE UserId=" + strconv.Itoa(UserId)
	_, err = DB.Exec("UPDATE roles SET Roles=? WHERE UserId=?", role, UserId)
	if err != nil {
		panic(err.Error())
	}
	res.Write([]byte(UserName + "User is now admin"))
}

func main() {
	dbName := "UserDB"
	dbPass := "Admin@123"
	dbDriver := "mysql"
	dbUser := "root"
	dsn := dbUser + ":" + dbPass + "@tcp(127.0.0.1:3306)/" + dbName
	db, err := sql.Open(dbDriver, dsn)
	DB = db
	if err != nil {
		panic(err.Error())
	}
	td, err := DB.Exec("Create database if not exists " + dbName + ";")
	if err != nil {
		fmt.Println("Error creating database UserDB")
	}
	if td != nil {
		fmt.Println("Created Successfully")
	}

	/*
		users table
			type users struct {
				UserId       int		PRIMARY KEY
				UserName     string
				UserEmail    string
				UserPassword string
			}
	*/

	/*
		roles table
			type roles struct {
				UserId int		PRIMARY KEY
				Roles  string
			}
	*/

	_, err = DB.Exec("Create Table if not exists users(UserId int primary key AUTO_INCREMENT, UserName varchar(50), UserEmail varchar(50), UserPassword varchar(50))")
	if err != nil {
		panic(err.Error())
	}
	_, err = DB.Exec("Create Table if not exists roles(UserId int primary key, Roles varchar(20))")
	if err != nil {
		panic(err.Error())
	}
	UserName := "Admin"
	var UserId int

	err = DB.QueryRow("SELECT UserId from users where UserName=?", UserName).Scan(&UserId)

	if err != nil {
		panic(err.Error())
	}

	http.HandleFunc("/", homePage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/addUser", AddUser)
	http.ListenAndServe(":8080", nil)
}
