package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

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
		http.Redirect(res, req, "index.html", 301)
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
		http.ServeFile(res, req, "addUser.html")
		return
	} else {
		http.ServeFile(res, req, "Member.html")
	}
}

func AddUser(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "addUser.html")
		return
	}
	UserName := req.FormValue("username")
	fmt.Println("Updating role")

	var UserId int
	err := DB.QueryRow("SELECT UserId from users where UserName=?", UserName).Scan(&UserId)
	if err != nil {
		if err != sql.ErrNoRows {
			panic(err.Error())
		}
		res.Write([]byte(UserName + " not found in our database"))
		return
	}
	role := "Admin"
	// query1 := "UPDATE UserDB.roles SET Roles=" + role + " WHERE UserId=" + strconv.Itoa(UserId)
	_, err = DB.Exec("UPDATE roles SET Roles=? WHERE UserId=?", role, UserId)
	if err != nil {
		panic(err.Error())
	}
	res.Write([]byte(UserName + " is now admin"))
}

func DeleteCity(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "deleteCity.html")
		return
	}

	deleteCity := req.FormValue("city")
	fmt.Println("deleting city " + deleteCity)

	//check if city is present or not
	var cityP string

	err := DB.QueryRow("SELECT city FROM cities where city=?", deleteCity).Scan(&cityP)
	if err != nil {
		if err != sql.ErrNoRows {
			panic(err.Error())
		}
		res.Write([]byte(deleteCity + " City not present"))
		return
	}

	_, err = DB.Exec("DELETE FROM cities where city=?", deleteCity)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Deleted city " + deleteCity)
	res.Write([]byte(deleteCity + " City deleted successfully"))
}

func AddCity(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.ServeFile(res, req, "addCity.html")
		return
	}

	cityAdd := req.FormValue("city")
	cityAdd = strings.ToUpper(cityAdd)
	stateAdd := req.FormValue("state")
	stateAdd = strings.ToUpper(stateAdd)

	//storing values if city and state already present
	var cityP string
	var stateP string

	//checking if city is already present
	err := DB.QueryRow("SELECT city, state from cities where city=?", cityAdd).Scan(&cityP, &stateP)
	if err != nil {
		if err == sql.ErrNoRows {
			//if not present then insert it
			fmt.Println("Adding City")
			_, err = DB.Exec("Insert into cities(city, state) Values(?, ?)", cityAdd, stateAdd)
			if err != nil {
				fmt.Println("City not added ")
				res.Write([]byte("Unable to add City " + cityAdd))
			}
			res.Write([]byte(cityAdd + " city successfully added"))
			return
		} else {
			panic(err.Error())
		}
	}

	//city already present
	fmt.Println("City already present in database")
	res.Write([]byte(cityAdd + " city already present"))
}

func homePage(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "index.html")
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

	_, err = DB.Exec("Create Table if not exists users(UserId int primary key AUTO_INCREMENT, UserName varchar(50), UserEmail varchar(50), UserPassword varchar(50))")
	if err != nil {
		panic(err.Error())
	}

	/*
		roles table
			type roles struct {
				UserId int		PRIMARY KEY
				Roles  string
			}
	*/

	_, err = DB.Exec("Create Table if not exists roles(UserId int primary key, Roles varchar(20))")
	if err != nil {
		panic(err.Error())
	}

	/*

		city table
		type cities struct {
			city string
			state string
		}
	*/

	_, err = DB.Exec("Create Table if not exists cities(city varchar(30) primary key, state varchar(30))")
	if err != nil {
		panic(err.Error())
	}

	UserName := "Admin"
	var UserId int

	err = DB.QueryRow("SELECT UserId from users where UserName=?", UserName).Scan(&UserId)

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Running Port on: 8080")
	http.HandleFunc("/", homePage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/signup", signupPage)
	http.HandleFunc("/addUser", AddUser)
	http.HandleFunc("/addCity", AddCity)
	http.HandleFunc("/deleteCity", DeleteCity)
	http.ListenAndServe(":8080", nil)
}
