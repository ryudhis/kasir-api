package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Stock int    `json:"stock"`
}

var product = []Product{
	{ID: 1, Name: "Sarimie Istri 3", Price: 6969, Stock: 69},
	{ID: 2, Name: "BahleelOil", Price: 9000, Stock: 20},
	{ID: 3, Name: "Minyak Jelantah", Price: 1500, Stock: 15},
}

func getProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/product/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	for _, p := range product {
		if p.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(p)
			return
		}
	}
	http.Error(w, "Product not found", http.StatusNotFound)
}

func updateProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/product/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var updateProduct Product
	err = json.NewDecoder(r.Body).Decode(&updateProduct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for i := range product {
		if product[i].ID == id {
			updateProduct.ID = id
			product[i] = updateProduct
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updateProduct)
			return
		}
	}
	http.Error(w, "Product not found", http.StatusNotFound)
}

func deleteProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/product/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	for i := range product {
		if product[i].ID == id {
			product = append(product[:i], product[i+1:]...)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
			"message": "Product deleted successfully.",
		})
			return
		}
	}
	http.Error(w, "Product not found", http.StatusNotFound)
}

func main() {

	//GET localhost:8080/product/{id}
	//PUT localhost:8080/product/{id}
	http.HandleFunc("/api/product/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getProductByID(w, r)
		case "PUT":
			updateProductByID(w, r)
		case "DELETE":
			deleteProductByID(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	//GET localhost:8080/product
	//POST localhost:8080/product
	http.HandleFunc("/api/product", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(product)
		case "POST":
			var newProduct Product
			err := json.NewDecoder(r.Body).Decode(&newProduct)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			newProduct.ID = len(product) + 1
			product = append(product, newProduct)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newProduct)
		}

	})

	//localhost:8080/health
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OK",
			"message": "API RUNNING",
		})
	})
	fmt.Println("Server is running di port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Gagal running server")
	}
}
