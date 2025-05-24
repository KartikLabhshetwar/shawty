package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type URL struct {
	ID           string    `json:"id" bson:"_id"` // Added bson tag for MongoDB
	OriginalUrl  string    `json:"original_url" bson:"original_url"`
	ShortUrl     string    `json:"short_url" bson:"short_url"`
	CreationDate time.Time `json:"creation_date" bson:"creation_date"`
}

// var urlDB = make(map[string]URL) // We will replace this with MongoDB

var urlCollection *mongo.Collection

/*
   0a137b37 {
       ID: "0a137b37",
       OriginalUrl: "www.google.com",
       ShortUrl:"0a137b37",
       CreationDate: time.Now()

   }
*/

func connectDB() *mongo.Collection {
	// Find .env
	err := godotenv.Load(".env") // Assuming .env is in the same directory as the executable
	if err != nil {
		log.Printf("Error loading .env file: %s. Make sure it exists.", err)
		// Continue without .env if MONGO_URI is set in environment
	}

	// Get value from .env or environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not found in .env file or environment variables")
	}

	// Connect to the database.
	clientOption := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOption)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection.
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create collection
	// You can change "shawtydb" and "urls" to your preferred database and collection names
	collection := client.Database("shawtydb").Collection("urls")
	fmt.Println("Connected to MongoDB!")
	return collection
}

func generateShortURL(OriginalUrl string) string {
	hasher := md5.New()
	hasher.Write([]byte(OriginalUrl)) // it converts the originalURL to byte slice.
	data := hasher.Sum(nil)
	hash := hex.EncodeToString(data)
	fmt.Println("final string:", hash[:8])
	return hash[:8]
}

func createURL(originalURL string) string {
	shortURL := generateShortURL(originalURL)
	id := shortURL // use the short URL as ID for simplicity

	urlEntry := URL{
		ID:           id,
		OriginalUrl:  originalURL,
		ShortUrl:     shortURL,
		CreationDate: time.Now(),
	}

	// Insert into MongoDB
	_, err := urlCollection.InsertOne(context.Background(), urlEntry)
	if err != nil {
		// Handle error, perhaps return an error or log it
		// For now, we'll log and proceed, but in a real app, you'd want better error handling
		log.Printf("Failed to insert URL into MongoDB: %v", err)
		// You might want to check if the error is due to a duplicate key if you have a unique index on ID
		// and handle it accordingly (e.g., return the existing short URL).
	}

	return shortURL
}

func getURL(id string) (URL, error) {
	var url URL
	// Find in MongoDB
	filter := bson.M{"_id": id} // In MongoDB, the default ID field is _id
	err := urlCollection.FindOne(context.Background(), filter).Decode(&url)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return URL{}, errors.New("URL not found")
		}
		log.Printf("Error retrieving URL from MongoDB: %v", err)
		return URL{}, err
	}

	return url, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world from Shawty URL Shortener!")
}

func shortURLHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	if data.URL == "" {
		http.Error(w, "URL field is missing or empty", http.StatusBadRequest)
		return
	}

	shortenedURL := createURL(data.URL)

	response := struct {
		ShortURL string `json:"short_url"`
	}{ShortURL: shortenedURL}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		// It's unlikely json.NewEncoder(w).Encode(response) would fail here with http.StatusBadRequest
		// More likely an internal server error if writing to ResponseWriter fails.
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to create response", http.StatusInternalServerError)
		return
	}
}

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/redirect/"):]
	if id == "" {
		http.Error(w, "Short URL ID is missing", http.StatusBadRequest)
		return
	}

	url, err := getURL(id)
	if err != nil {
		if err.Error() == "URL not found" {
			http.Error(w, "Short URL not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving URL", http.StatusInternalServerError)
		}
		return
	}

	// It's good practice to ensure OriginalUrl is a valid URL before redirecting
	// For simplicity, we're skipping that here.
	// Also, ensure it has a scheme (http:// or https://)
	// If not, you might want to prepend "http://"
	redirectURL := url.OriginalUrl
	if !strings.HasPrefix(redirectURL, "http://") && !strings.HasPrefix(redirectURL, "https://") {
		redirectURL = "http://" + redirectURL
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func main() {
	// Initialize MongoDB connection
	urlCollection = connectDB()

	// Ensure the client is disconnected when the application exits
	// defer func() {
	// 	if err := urlCollection.Database().Client().Disconnect(context.Background()); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println("Disconnected from MongoDB.")
	// }()
	// Deferring client disconnect can be tricky with log.Fatal in connectDB.
	// For a simple app, OS will clean up connections. For robust apps, manage client lifecycle carefully.

	http.HandleFunc("/", handler)
	http.HandleFunc("/shorten", shortURLHandler)
	http.HandleFunc("/redirect/", redirectURLHandler)

	fmt.Println("Server starting on port 8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error on starting the server", err)
		return
	}
}
