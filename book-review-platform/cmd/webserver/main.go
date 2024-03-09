package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
	ctx    context.Context
)

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
	Email    string `bson:"email"`
}

type Book struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title"`
	Author      string             `bson:"author"`
	Description string             `bson:"description"`
}

func registerPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/register.html")
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/login.html")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	collection := client.Database("book").Collection("users")
	_, err = collection.InsertOne(ctx, bson.M{"username": username, "password": password, "email": email})
	if err != nil {
		http.Error(w, "Error inserting user into database", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	collection := client.Database("book").Collection("users")
	var user User
	err = collection.FindOne(ctx, bson.M{"username": username, "password": password}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("book").Collection("books")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Error retrieving book reviews: %v\n", err)
		http.Error(w, "Error retrieving book reviews", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var books []Book
	for cursor.Next(ctx) {
		var book Book
		if err := cursor.Decode(&book); err != nil {
			log.Printf("Error decoding book review: %v\n", err)
			http.Error(w, "Error decoding book review", http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	tmpl, err := template.ParseFiles("./web/templates/index.html")
	if err != nil {
		log.Printf("Error parsing HTML template: %v\n", err)
		http.Error(w, "Error parsing HTML template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Books []Book
	}{
		Books: books,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing HTML template: %v\n", err)
		http.Error(w, "Error executing HTML template", http.StatusInternalServerError)
		return
	}
}

func createBookHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")

	id := primitive.NewObjectID()

	collection := client.Database("book").Collection("books")
	_, err = collection.InsertOne(ctx, bson.M{"_id": id, "title": title, "author": author, "description": description})
	if err != nil {
		http.Error(w, "Error creating book", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func editBookHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}
	title := r.FormValue("title")
	author := r.FormValue("author")
	description := r.FormValue("description")

	collection := client.Database("book").Collection("books")
	var book Book
	err = collection.FindOne(ctx, bson.M{"title": title}).Decode(&book)
	if err != nil {
		log.Printf("Error retrieving book details: %v\n", err)
		http.Error(w, "Error retrieving book details", http.StatusInternalServerError)
		return
	}

	book.Author = author
	book.Description = description

	_, err = collection.ReplaceOne(ctx, bson.M{"_id": book.ID}, book)
	if err != nil {
		log.Printf("Error updating book review: %v\n", err)
		http.Error(w, "Error updating book review", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func deleteBookReviewHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	collection := client.Database("book").Collection("books")
	_, err = collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		log.Printf("Error deleting book review: %v\n", err)
		http.Error(w, "Error deleting book review", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/register", registerPageHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/login", loginPageHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/index", indexPageHandler).Methods("GET")
	r.HandleFunc("/create-book", createBookHandler).Methods("POST")
	r.HandleFunc("/delete-book-review", deleteBookReviewHandler).Methods("POST")
	r.HandleFunc("/edit-book", editBookHandler).Methods("GET")

	port := ":8080"
	fmt.Printf("Server listening on port %s\n", port)

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connectionString := "mongodb+srv://Ansar:Ansar02012004@cluster0.eotba7o.mongodb.net/"
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Connected to MongoDB Atlas!")

	go func() {
		if err := http.ListenAndServe(port, r); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("Shutting down server...")
}
