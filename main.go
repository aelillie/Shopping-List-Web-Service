package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/gorilla/mux"
)

const itemKind string = "Item"

//Item represents a shopping item with name, supermarket and price
type Item struct {
	Name        string  `json:"name"`
	Supermarket string  `json:"supermarket"`
	Price       float32 `json:"price"`
}

func decodeItem(reader io.ReadCloser) (Item, error) {
	var i Item

	dec := json.NewDecoder(reader)
	err := dec.Decode(&i)
	return i, err
}

func createItem(w http.ResponseWriter, r *http.Request) {
	item, err := decodeItem(r.Body)
	if err != nil {
		http.Error(w, "could not decode body", http.StatusBadRequest)
		return
	}

	// items[item.Name] = item
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, itemKind, item.Name, 0, nil) //Stores it with item name as key
	key, err = datastore.Put(ctx, key, &item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Succesfully stored:{%v : %v}", key, item)
	w.WriteHeader(http.StatusOK)
}

func removeItems(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil { //no params given
		http.Error(w, "no parameters given", http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)
	q := datastore.NewQuery(itemKind).KeysOnly()
	if r.FormValue("delete-all") != "true" { //delete specific items
		names := r.Form["name"]
		for _, name := range names {
			q = q.Filter("Name=", name)
		}
	} //otherwise delete all
	keys, err := q.GetAll(ctx, nil)
	err = datastore.DeleteMulti(ctx, keys)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func getTotalPrice(w http.ResponseWriter, r *http.Request) {
	var total float32
	items, err := retrieveAllItems(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range items {
		total += item.Price
	}
	fmt.Fprintf(w, "Total price:%f", total)
	w.WriteHeader(http.StatusOK)
}

func getSupermarketItems(w http.ResponseWriter, r *http.Request) {
	supermarket := strings.TrimSpace(strings.ToLower(mux.Vars(r)["supermarket"]))
	if supermarket == "" {
		http.Error(w, "no supermarket value provided", http.StatusBadRequest)
		return
	}
	items, err := retrieveAllItems(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range items {
		if strings.TrimSpace(strings.ToLower(item.Supermarket)) == supermarket {
			fmt.Fprintf(w, "%#v\n", item)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func retrieveAllItems(r *http.Request) ([]Item, error) {
	ctx := appengine.NewContext(r)

	var items []Item
	q := datastore.NewQuery(itemKind).Order("Name")
	_, err := q.GetAll(ctx, &items)
	return items, err
}

func getAllItems(w http.ResponseWriter, r *http.Request) {
	items, err := retrieveAllItems(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, item := range items {
		fmt.Fprintln(w, item)
	}
}

func getItem(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("itemName")

	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, itemKind, name, 0, nil)

	var item Item
	err := datastore.Get(ctx, key, &item)

	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	fmt.Fprintln(w, item)
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the shopping list")

}

func init() {
	r := mux.NewRouter()

	r.HandleFunc("/", welcomeHandler)
	r.HandleFunc("/items", createItem).Methods("POST")                       //Add item
	r.HandleFunc("/items", removeItems).Methods("DELETE")                    //Remove single/all item(s)
	r.HandleFunc("/items/total-price", getTotalPrice).Methods("GET")         //Total price for all items
	r.HandleFunc("/items/{supermarket}", getSupermarketItems).Methods("GET") //All items for one supermarket
	r.HandleFunc("/items", getItem).Methods("GET")                           //Retrieve single item

	r.HandleFunc("/items/", getAllItems).Methods("GET")

	http.Handle("/", r)
	// err := http.ListenAndServe("localhost:8080", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
