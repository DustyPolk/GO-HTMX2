package main

import (
    "database/sql"
    "log"
    "net/http"
    "html/template"

    _ "github.com/mattn/go-sqlite3"
)

type ToDoItem struct {
    ID   int
    Task string
}

func initDB() *sql.DB {
    db, err := sql.Open("sqlite3", "./todos.db")
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec("CREATE TABLE IF NOT EXISTS todos (id INTEGER PRIMARY KEY AUTOINCREMENT, task TEXT NOT NULL)")
    if err != nil {
        log.Fatal(err)
    }

    return db
}

func main() {
    db := initDB()
    defer db.Close()

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        tmpl := template.Must(template.ParseFiles("index.html"))
        tmpl.Execute(w, nil)
    })

    http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
        log.Println("Received a request to /todos")

        if r.Method == "POST" {
            err := r.ParseForm()
            if err != nil {
                log.Printf("Error parsing form: %v\n", err)
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            task := r.FormValue("task")
            log.Printf("Parsed ToDo item: %s\n", task)

            _, err = db.Exec("INSERT INTO todos (task) VALUES (?)", task)
            if err != nil {
                log.Printf("Error adding item: %v\n", err)
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            listTodosHandler(db)(w, r)
        } else {
            listTodosHandler(db)(w, r)
        }
    })

    http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Query().Get("id")
        log.Printf("Deleting ToDo item with ID: %s\n", id)
        _, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
        if err != nil {
            log.Printf("Error deleting item: %v\n", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    })

    log.Println("Server started at http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

func listTodosHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rows, err := db.Query("SELECT * FROM todos")
        if err != nil {
            log.Printf("Error querying items: %v\n", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var todos []ToDoItem
        for rows.Next() {
            var todo ToDoItem
            if err := rows.Scan(&todo.ID, &todo.Task); err != nil {
                log.Printf("Error scanning item: %v\n", err)
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            todos = append(todos, todo)
        }

        tmpl := template.Must(template.ParseFiles("todos.html"))
        tmpl.Execute(w, todos)
    }
}
