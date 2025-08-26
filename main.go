package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	client       *mongo.Client
	databaseName = "digitalthread"
)

func init() {
	_ = godotenv.Load()
}

// jsonResponse sets header and writes JSON.
func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// jsonResponse writes a JSON response with the given status code
func jsonResponse2(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func main() {
	ctx := context.Background()
	client = connectMongo(ctx)
	defer client.Disconnect(ctx)

	r := mux.NewRouter()

	// Auction functions
	r.HandleFunc("/createAuction", CreateAuction).Methods("POST") // Adds a prder object to MongoDB/data_objects
	r.HandleFunc("/getStartList", GetAuctionStartList).Methods("GET")
	r.HandleFunc("/getStopList/{duration}", GetAuctionStopList).Methods("GET")
	r.HandleFunc("/startAuction/{auctionId}/{startTime}", StartAuction).Methods("PUT")
	r.HandleFunc("/stopAuction/{auctionId}/{endTime}", StopAuction).Methods("PUT")
	r.HandleFunc("/getAwaitingWinnerList", GetAwaitingWinnerList).Methods("GET")
	r.HandleFunc("/setAuctionWinner/{auctionId}/{bidId}", SetAuctionWinner).Methods("PUT")

	// Bid functions
	r.HandleFunc("/addBid", AddBid).Methods("POST")
	r.HandleFunc("/getBidList/{auctionId}", GetBidList).Methods("GET")

	port := os.Getenv("AUCTION_PORT")
	if port == "" {
		log.Println("Error: AUCTION_PORT is empty")
		return
	}

	log.Printf("Auction Manager is listening on :%s", port) // 9090
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
