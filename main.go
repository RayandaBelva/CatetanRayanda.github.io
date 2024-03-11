package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Kebutuhan struct {
	Id        int    `json:"id"`
	Tanggal   string `json:"tanggal"`
	Kebutuhan string `json:"kebutuhan"`
	Jumlah    int    `json:"jumlah"`
	Uang      int    `json:"uang"`
}

var catatan []Kebutuhan

var fileName string = "data.csv"

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/catatan", getAllCatetan).Methods("GET")
	r.HandleFunc("/catatan/{id}", getCatetanByID).Methods("GET")
	r.HandleFunc("/catatan", addNewCatetan).Methods("POST")
	r.HandleFunc("/catatan/{id}", deleteCatetan).Methods("DELETE")
	r.HandleFunc("/addcatetan", serveAddCatetanHTML).Methods("GET")
	r.HandleFunc("/listcatetan", serveCatetanListHTML).Methods("GET")
	r.HandleFunc("/deletecatetan", serveDeleteCatetanHTML).Methods("GET")
	r.HandleFunc("/totaluang", getTotalUang).Methods("GET")
	r.HandleFunc("/", homeHandler).Methods("GET")
	loadDataFromCSV(fileName)

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe("0.0.0.0:8080", r)
}

func serveAddCatetanHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "addcatetan.html")
}

func serveCatetanListHTML(w http.ResponseWriter, r *http.Request) {
	// Baca file HTML
	file, err := os.Open("listcatetan.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Kirim header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Salin file HTML ke response writer
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Hitung total uang
	getTotalUang(w, r)
}

func serveDeleteCatetanHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "deletecatetan.html")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Mengirimkan file HTML home.html sebagai respons
	http.ServeFile(w, r, "home.html")
}

func addNewCatetan(w http.ResponseWriter, r *http.Request) {
	var newCatetan Kebutuhan

	err := json.NewDecoder(r.Body).Decode(&newCatetan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mencari nilai maksimum dari ID yang sudah ada
	maxID := 0
	for _, catetan := range catatan {
		if catetan.Id > maxID {
			maxID = catetan.Id
		}
	}

	// Mengatur ID baru
	newCatetan.Id = maxID + 1

	catatan = append(catatan, newCatetan)

	err = saveDataToCSV(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCatetan)
}

func getAllCatetan(w http.ResponseWriter, r *http.Request) {
	if len(catatan) == 0 {
		http.Error(w, "No catatan available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catatan)
}

func getCatetanByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	catetanID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid catetan ID", http.StatusBadRequest)
		return
	}

	catetan, err := findCatetanById(catetanID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Kebutuhan with ID %d not found", catetanID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catetan)
}

func saveDataToCSV(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, catetan := range catatan {
		// Format data sesuai dengan struktur Kebutuhan
		row := fmt.Sprintf("%d,%s,%s,%d,%d\n", catetan.Id, catetan.Tanggal, catetan.Kebutuhan, catetan.Jumlah, catetan.Uang)
		_, err := writer.WriteString(row)
		if err != nil {
			return fmt.Errorf("error writing to CSV file: %v", err)
		}
	}

	return nil
}

func loadDataFromCSV(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("error opening CSV file: %w", err)
	}
	defer file.Close()

	// Clear existing catatan before loading from CSV
	catatan = nil

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		record := strings.Split(scanner.Text(), ",")
		id, _ := strconv.Atoi(record[0])
		tanggal := record[1] // Tanggal sebagai string
		jumlah, _ := strconv.Atoi(record[3])
		uang, _ := strconv.Atoi(record[4])

		catetan := Kebutuhan{
			Id:        id,
			Tanggal:   tanggal,
			Kebutuhan: record[2],
			Jumlah:    jumlah,
			Uang:      uang,
		}
		catatan = append(catatan, catetan)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading CSV file: %w", err)
	}
	return nil
}

func findCatetanById(id int) (Kebutuhan, error) {
	for _, catetan := range catatan {
		if catetan.Id == id {
			return catetan, nil
		}
	}
	return Kebutuhan{}, fmt.Errorf("ID %d not found", id)
}

func deleteCatetan(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	catetanID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid catetan ID", http.StatusBadRequest)
		return
	}

	index := -1
	for i, catetan := range catatan {
		if catetan.Id == catetanID {
			index = i
			break
		}
	}

	if index == -1 {
		http.Error(w, fmt.Sprintf("Kebutuhan with ID %d not found", catetanID), http.StatusNotFound)
		return
	}

	catatan = append(catatan[:index], catatan[index+1:]...)

	err = saveDataToCSV(fileName)
	if err != nil {
		http.Error(w, "Failed to delete catetan", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Catetan with ID %d has been deleted", catetanID)
}

func getTotalUang(w http.ResponseWriter, r *http.Request) {
	total := 0
	for _, catetan := range catatan {
		total += catetan.Uang
	}
	fmt.Fprintf(w, "Total Uang: %d", total)
}
