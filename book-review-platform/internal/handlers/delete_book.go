package handlers

import (
	"context"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func deleteBookReviewHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client, ctx context.Context) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}
	title := r.FormValue("title")

	collection := client.Database("book").Collection("books")
	_, err = collection.DeleteOne(ctx, bson.M{"title": title})
	if err != nil {
		log.Printf("Error deleting book review: %v\n", err)
		http.Error(w, "Error deleting book review", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
