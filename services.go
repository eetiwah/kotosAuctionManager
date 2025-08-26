package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection string
var orderCollection string
var auctionCollection string
var bidCollection string

// connectMongo opens a MongoDB connection.
func connectMongo(ctx context.Context) *mongo.Client {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatalln("MONGO_URI is empty")
	}

	collection = os.Getenv("MONGO_COLLECTION_NAME")
	if collection == "" {
		log.Fatalln("collectionName is empty")
	}

	auctionCollection = os.Getenv("AUCTION_COLLECTION_NAME")
	if auctionCollection == "" {
		log.Fatalln("auctionCollectionName is empty")
	}

	bidCollection = os.Getenv("BID_COLLECTION_NAME")
	if bidCollection == "" {
		log.Fatalln("bidCollectionName is empty")
	}

	orderCollection = os.Getenv("ORDER_COLLECTION_NAME")
	if orderCollection == "" {
		log.Fatalln("orderCollectionName is empty")
	}

	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo.Connect: %v", err)
	}

	if err := cl.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo.Ping: %v", err)
	}
	return cl
}

/*
func Create(w http.ResponseWriter, r *http.Request) {
	var obj MetaDataObject
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, err := client.Database(databaseName).
		Collection(collection).
		InsertOne(ctx, obj)
	if err != nil {
		http.Error(w, "insert error", http.StatusInternalServerError)
		return
	}

	log.Printf("Added doc with guid = %s\n", obj.GUID)
	jsonResponse(w, obj.GUID)
}

func Read(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid := strings.TrimSpace(vars["guid"])
	if guid == "" {
		http.Error(w, "missing guid", http.StatusBadRequest)
		log.Printf("Missing guid in request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var obj MetaDataObject
	err := client.Database(databaseName).
		Collection(collection).
		FindOne(ctx, bson.M{"guid": guid}).
		Decode(&obj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for guid: %s", guid)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for guid %s: %v", guid, err)
		}
		return
	}

	log.Printf("Found document for guid: %s", guid)
	jsonResponse(w, obj)
}

func Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid := strings.TrimSpace(vars["guid"])
	if guid == "" {
		http.Error(w, "missing guid", http.StatusBadRequest)
		log.Printf("Missing guid in request")
		return
	}

	var obj MetaDataObject
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, err := client.Database(databaseName).
		Collection(collection).
		ReplaceOne(ctx, bson.M{"guid": guid}, obj)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated doc with guid = %s", guid)
	jsonResponse(w, obj)
}

func UpdateURI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid := strings.TrimSpace(vars["guid"])
	if guid == "" {
		http.Error(w, "missing guid", http.StatusBadRequest)
		log.Printf("Missing guid in request")
		return
	}

	var obj CloneResponse
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var metaObj MetaDataObject
	err := client.Database(databaseName).
		Collection(collection).
		FindOne(ctx, bson.M{"guid": guid}).
		Decode(&metaObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for guid: %s", guid)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for guid %s: %v", guid, err)
		}
		return
	}

	ctx, cancel = context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Update uri && cloned
	metaObj.URI = obj.URI
	metaObj.Cloned = true

	_, err = client.Database(databaseName).
		Collection(collection).
		ReplaceOne(ctx, bson.M{"guid": obj.GUID}, metaObj)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated URI in doc with guid = %s", guid)
	jsonResponse(w, metaObj)
}

func UpdateOrderUri(w http.ResponseWriter, r *http.Request) {
	log.Printf("UpdateOrderUri")

	var obj CloneResponse
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		http.Error(w, fmt.Sprintf("bad request: %v", err), http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}

	// Validate GUID (optional: ensure it's a valid UUID)
	if _, err := uuid.Parse(obj.GUID); err != nil {
		http.Error(w, "invalid GUID format", http.StatusBadRequest)
		log.Printf("Invalid GUID: %s", obj.GUID)
		return
	}

	// Validate URI (optional: basic check for non-empty)
	if obj.URI == "" {
		http.Error(w, "URI cannot be empty", http.StatusBadRequest)
		log.Printf("Empty URI for GUID: %s", obj.GUID)
		return
	}

	// Define filter to find documents where orderitem.guid matches receivedObj.GUID
	filter := bson.M{"orderitem.guid": obj.GUID}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Count matching documents
	count, err := client.Database(databaseName).
		Collection(orderCollection).
		CountDocuments(ctx, filter)
	if err != nil {
		msg := fmt.Sprintf("failed to count documents: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Error: %s for GUID %s", msg, obj.GUID)
		return
	}

	// Check if more than one document matches
	if count > 1 {
		msg := fmt.Sprintf("multiple documents (%d) found with GUID %s", count, obj.GUID)
		http.Error(w, msg, http.StatusConflict)
		log.Printf("Error: %s", msg)
		return
	}

	// Prepare response
	response := Response{GUID: obj.GUID}

	// If no documents found, return early
	if count == 0 {
		log.Printf("No orders found for GUID %s", obj.GUID)
		jsonResponse2(w, http.StatusOK, response)
		return
	}

	// Find the single matching document
	var order OrderObject
	err = client.Database(databaseName).
		Collection(orderCollection).
		FindOne(ctx, filter).Decode(&order)
	if err == mongo.ErrNoDocuments {
		// This shouldn't happen since count == 1, but handle for robustness
		log.Printf("No orders found for GUID %s (unexpected)", obj.GUID)
		jsonResponse2(w, http.StatusOK, response)
		return
	}
	if err != nil {
		msg := fmt.Sprintf("failed to find order: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Error: %s for GUID %s", msg, obj.GUID)
		return
	}

	// Update the document in MongoDB
	update := bson.M{
		"$set": bson.M{
			"orderitem.uri": obj.URI,
		},
	}
	result, err := client.Database(databaseName).
		Collection(orderCollection).
		UpdateOne(ctx, bson.M{"_id": order.ID}, update)
	if err != nil {
		msg := fmt.Sprintf("failed to update order %s: %v", order.ID.Hex(), err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Error: %s for GUID %s", msg, obj.GUID)
		return
	}

	// Update response
	response.Updated = int(result.ModifiedCount)
	if result.ModifiedCount > 0 {
		log.Printf("Updated order %s: set orderitem.uri to %s", order.ID.Hex(), obj.URI)
	} else {
		log.Printf("No changes made to order %s for GUID %s", order.ID.Hex(), obj.GUID)
	}

	log.Printf("Updated URI in doc with guid = %s", obj.GUID)

	// Return JSON response
	jsonResponse2(w, http.StatusOK, response)

	//log.Printf("Updated URI in doc with guid = %s", obj.GUID)
	//jsonResponse(w, obj.GUID)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid := strings.TrimSpace(vars["guid"])
	if guid == "" {
		http.Error(w, "missing guid", http.StatusBadRequest)
		log.Printf("Missing guid in request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := client.Database(databaseName).
		Collection(collection).
		DeleteOne(ctx, bson.M{"guid": guid})

	if err != nil {
		log.Printf("delete error: %v", http.StatusInternalServerError)
		http.Error(w, "delete error", http.StatusInternalServerError)
		return
	}

	log.Printf("Deleted doc with guid = %s", guid)
	w.WriteHeader(http.StatusOK)
}
*/

/*
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderObj OrderObject
	if err := json.NewDecoder(r.Body).Decode(&orderObj); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Add order object to orderCollection
	_, err := client.Database(databaseName).
		Collection(orderCollection).
		InsertOne(ctx, orderObj)
	if err != nil {
		http.Error(w, "insert error", http.StatusInternalServerError)
		return
	}

	log.Printf("Order %s was added to orderTable\n", orderObj.OrderId)
	jsonResponse(w, orderObj.OrderId)
}

func GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId := strings.TrimSpace(vars["orderId"])
	if orderId == "" {
		http.Error(w, "missing orderId", http.StatusBadRequest)
		log.Printf("Missing orderId in request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var orderObj OrderObject
	err := client.Database(databaseName).
		Collection(orderCollection).
		FindOne(ctx, bson.M{"orderid": orderId}).
		Decode(&orderObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for orderId: %s", orderId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for orderId %s: %v", orderId, err)
		}
		return
	}

	// Send JSON response
	jsonResponse2(w, http.StatusOK, orderObj)

}
*/

func CreateAuction(w http.ResponseWriter, r *http.Request) {
	var auctionObj AuctionObject
	if err := json.NewDecoder(r.Body).Decode(&auctionObj); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ctx2, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Add auction object to auctionCollection
	_, err := client.Database(databaseName).
		Collection(auctionCollection).
		InsertOne(ctx2, auctionObj)
	if err != nil {
		http.Error(w, "insert error", http.StatusInternalServerError)
		return
	}

	log.Printf("Auction %s was added to auctionTable\n", auctionObj.AuctionId)
	jsonResponse(w, auctionObj.AuctionId)
}

// Returns a list of auctions that need to be started
func GetAuctionStartList(w http.ResponseWriter, r *http.Request) {
	// Create filter for StartDate = empty
	filter := bson.M{"startdate": time.Time{}}

	// Set the timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := client.Database(databaseName).
		Collection(auctionCollection).
		Find(ctx, filter)
	if err != nil {
		msg := fmt.Sprintf("failed to query database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Query error: %s", msg)
		return
	}
	defer cursor.Close(ctx)

	// Collect results
	var results []AuctionObject
	if err = cursor.All(ctx, &results); err != nil {
		msg := fmt.Sprintf("failed to read results: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Cursor error: %s", msg)
		return
	}

	// Log the number of results
	log.Printf("Found %d auctions with unset StartDate", len(results))

	// Send JSON response
	jsonResponse2(w, http.StatusOK, results)
}

// Returns a list of auctions that need to be stopped -> startDate + duration <= current time
func GetAuctionStopList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	durationStr := vars["duration"]
	if durationStr == "" {
		http.Error(w, "missing duration", http.StatusBadRequest)
		log.Printf("Missing duration in request")
		return
	}

	durationSeconds, err := strconv.Atoi(durationStr)
	if err != nil {
		http.Error(w, "duration must be a valid integer", http.StatusBadRequest)
		log.Printf("Invalid duration: %s", durationStr)
		return
	}

	if durationSeconds <= 0 {
		msg := "duration must be positive"
		http.Error(w, msg, http.StatusBadRequest)
		log.Printf("Non-positive duration: %d", durationSeconds)
		return
	}

	// Compute threshold: startdate + duration <= now => startdate <= now - duration
	duration := time.Duration(durationSeconds) * time.Second
	threshold := time.Now().Add(-duration)

	// Create filter
	filter := bson.M{
		"enddate":   time.Time{},
		"startdate": bson.M{"$lte": threshold},
	}

	// Set the timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := client.Database(databaseName).
		Collection(auctionCollection).
		Find(ctx, filter)
	if err != nil {
		msg := fmt.Sprintf("failed to query database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Query error: %s", msg)
		return
	}
	defer cursor.Close(ctx)

	// Collect results
	var results []AuctionObject
	if err = cursor.All(ctx, &results); err != nil {
		msg := fmt.Sprintf("failed to read results: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Cursor error: %s", msg)
		return
	}

	// Log the number of results
	log.Printf("Found %d auctions with unset EndDate", len(results))

	// Send JSON response
	jsonResponse2(w, http.StatusOK, results)
}

func StartAuction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	startTimeStr := strings.TrimSpace(vars["startTime"])
	if startTimeStr == "" {
		http.Error(w, "missing startTime", http.StatusBadRequest)
		log.Printf("Missing startTime in request")
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		http.Error(w, "startTime conversion from string to time.Time failed", http.StatusBadRequest)
		log.Printf("startTime conversion from string to time.Time failed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var auctionObj AuctionObject
	err = client.Database(databaseName).
		Collection(auctionCollection).
		FindOne(ctx, bson.M{"auctionid": auctionId}).
		Decode(&auctionObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for guid: %s", auctionId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for guid %s: %v", auctionId, err)
		}
		return
	}

	// Update startDate to reflect that the auction has started
	auctionObj.StartDate = startTime

	_, err = client.Database(databaseName).
		Collection(auctionCollection).
		ReplaceOne(ctx, bson.M{"auctionid": auctionId}, auctionObj)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	log.Printf("Auction: %s was updated", auctionId)
	w.WriteHeader(http.StatusOK)
}

func StopAuction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	endTimeStr := strings.TrimSpace(vars["endTime"])
	if endTimeStr == "" {
		http.Error(w, "missing endTime", http.StatusBadRequest)
		log.Printf("Missing endTime in request")
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		http.Error(w, "startTime conversion from string to time.Time failed", http.StatusBadRequest)
		log.Printf("startTime conversion from string to time.Time failed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var auctionObj AuctionObject
	err = client.Database(databaseName).
		Collection(auctionCollection).
		FindOne(ctx, bson.M{"auctionid": auctionId}).
		Decode(&auctionObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for auctionId: %s", auctionId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for auctionId %s: %v", auctionId, err)
		}
		return
	}

	// Update endDate to reflect that the auction has stopped
	auctionObj.EndDate = endTime

	_, err = client.Database(databaseName).
		Collection(auctionCollection).
		ReplaceOne(ctx, bson.M{"auctionid": auctionId}, auctionObj)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	log.Printf("Auction: %s was endDate was updated", auctionId)
	w.WriteHeader(http.StatusOK)
}

func GetAwaitingWinnerList(w http.ResponseWriter, r *http.Request) {
	// Create filter where enddate is not empty and winningbid is empty
	filter := bson.M{
		"enddate":    bson.M{"$ne": time.Time{}},
		"winningbid": bson.M{"$in": []interface{}{"", nil}},
	}

	// Set the timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := client.Database(databaseName).
		Collection(auctionCollection).
		Find(ctx, filter)
	if err != nil {
		msg := fmt.Sprintf("failed to query database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Query error: %s", msg)
		return
	}
	defer cursor.Close(ctx)

	// Collect results
	var results []AuctionObject
	if err = cursor.All(ctx, &results); err != nil {
		msg := fmt.Sprintf("failed to read results: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Cursor error: %s", msg)
		return
	}

	// Log the number of results
	log.Printf("Found %d pending winner auctions", len(results))

	// Send JSON response
	jsonResponse2(w, http.StatusOK, results)
}

func SetAuctionWinner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	bidId := strings.TrimSpace(vars["bidId"])
	if bidId == "" {
		http.Error(w, "missing bidId", http.StatusBadRequest)
		log.Printf("Missing bidId in request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var auctionObj AuctionObject
	err := client.Database(databaseName).
		Collection(auctionCollection).
		FindOne(ctx, bson.M{"auctionid": auctionId}).
		Decode(&auctionObj)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
			log.Printf("No document found for auctionId: %s", auctionId)
		} else {
			http.Error(w, "query error", http.StatusInternalServerError)
			log.Printf("Query error for auctionId %s: %v", auctionId, err)
		}
		return
	}

	// Update WinningBid to reflect that the auction has stopped
	auctionObj.WinningBid = bidId

	_, err = client.Database(databaseName).
		Collection(auctionCollection).
		ReplaceOne(ctx, bson.M{"auctionid": auctionId}, auctionObj)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	log.Printf("Auction: %s WinningBid was updated", auctionId)
	w.WriteHeader(http.StatusOK)
}

func AddBid(w http.ResponseWriter, r *http.Request) {
	var bidObj BidObject
	if err := json.NewDecoder(r.Body).Decode(&bidObj); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Add bid object to bidCollection
	_, err := client.Database(databaseName).
		Collection(bidCollection).
		InsertOne(ctx, bidObj)
	if err != nil {
		http.Error(w, "insert error", http.StatusInternalServerError)
		return
	}

	log.Printf("Bid %s was added to bid collection\n", bidObj.BidId)

	jsonResponse(w, bidObj.BidId)
}

func GetBidList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auctionId := strings.TrimSpace(vars["auctionId"])
	if auctionId == "" {
		http.Error(w, "missing auctionId", http.StatusBadRequest)
		log.Printf("Missing auctionId in request")
		return
	}

	filter := bson.M{"auctionid": auctionId}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cursor, err := client.Database(databaseName).
		Collection(bidCollection).
		Find(ctx, filter)
	if err != nil {
		msg := fmt.Sprintf("failed to query database: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Query error: %s", msg)
		return
	}
	defer cursor.Close(ctx)

	// Collect results
	var results []BidObject
	if err = cursor.All(ctx, &results); err != nil {
		msg := fmt.Sprintf("failed to read results: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		log.Printf("Cursor error: %s", msg)
		return
	}

	// Log the number of results
	log.Printf("Found %d bids", len(results))

	// Send JSON response
	jsonResponse2(w, http.StatusOK, results)
}
