package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/gorilla/mux"
)

// ================================================================
// CONFIGURATION
// ================================================================

const (
	CouchbaseHost       = "db"
	CouchbaseUsername   = "admin"
	CouchbasePassword   = "T1ku$H1t4m"
	CouchbaseBucket     = "ms_demo"
	CouchbaseScope      = "merchant"
	CouchbaseCollection = "summary"
	ServerPort          = ":8080"
)

// ================================================================
// DATA STRUCTURES
// ================================================================

// Request body structure
type MerchantSummaryRequest struct {
	MID []string `json:"mid"`
}

// Couchbase summary document structure
type MerchantSummary struct {
	MerchantID  string `json:"merchant_id"`
	SummaryDate string `json:"summary_date"`
	Amount      int64  `json:"amount"`
	Count       int    `json:"count"`
	LastTrxDate string `json:"last_trx_date"`
}

// Error schema
type ErrorSchema struct {
	ErrorCode    string            `json:"error_code"`
	ErrorMessage map[string]string `json:"error_message"`
}

// Output schema
type OutputSchema struct {
	MerchantIDs        []string `json:"merchant_ids"`
	CurrentDate        string   `json:"current_date"`
	TodayTotalAmount   string   `json:"today_total_amount"`
	WeeklyTotalAmount  string   `json:"weekly_total_amount"`
	MonthlyTotalAmount string   `json:"monthly_total_amount"`
}

// Response structure
type MerchantSummaryResponse struct {
	ErrorSchema  ErrorSchema  `json:"error_schema"`
	OutputSchema OutputSchema `json:"output_schema"`
}

// ================================================================
// GLOBAL VARIABLES
// ================================================================

var collection *gocb.Collection

// ================================================================
// MAIN FUNCTION
// ================================================================

func main() {
	// Initialize Couchbase connection
	var err error
	collection, err = initCouchbase()
	if err != nil {
		log.Fatalf("Failed to initialize Couchbase: %v", err)
	}

	// Setup HTTP router
	router := mux.NewRouter()
	router.HandleFunc("/api/merchant/summary", getMerchantSummary).Methods("POST")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Start server
	log.Printf("🚀 Server starting on port %s", ServerPort)
	log.Printf("📊 Endpoint: POST http://localhost%s/api/merchant/summary", ServerPort)
	log.Fatal(http.ListenAndServe(ServerPort, router))
}

// ================================================================
// COUCHBASE INITIALIZATION
// ================================================================

func initCouchbase() (*gocb.Collection, error) {
	log.Println("🔌 Connecting to Couchbase...")

	// Connect to cluster
	cluster, err := gocb.Connect(
		fmt.Sprintf("couchbase://%s", CouchbaseHost),
		gocb.ClusterOptions{
			Authenticator: gocb.PasswordAuthenticator{
				Username: CouchbaseUsername,
				Password: CouchbasePassword,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// Get bucket
	bucket := cluster.Bucket(CouchbaseBucket)

	// Wait until bucket is ready
	err = bucket.WaitUntilReady(10*time.Second, nil)
	if err != nil {
		return nil, fmt.Errorf("bucket not ready: %w", err)
	}

	// Get collection
	collection := bucket.Scope(CouchbaseScope).Collection(CouchbaseCollection)

	log.Printf("✅ Connected to Couchbase: %s.%s.%s", CouchbaseBucket, CouchbaseScope, CouchbaseCollection)

	return collection, nil
}

// ================================================================
// API HANDLERS
// ================================================================

// Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Get merchant summary endpoint
func getMerchantSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req MerchantSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, "E001", "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.MID) == 0 {
		sendErrorResponse(w, "E002", "Merchant IDs are required", http.StatusBadRequest)
		return
	}

	log.Printf("📊 Processing summary for %d merchants", len(req.MID))

	// Get current date
	currentDate := time.Now()
	currentDateStr := currentDate.Format("2006-01-02")

	// Calculate totals
	todayTotal, err := calculateDailyTotal(req.MID, currentDate)
	if err != nil {
		log.Printf("❌ Error calculating daily total: %v", err)
		sendErrorResponse(w, "E003", "Failed to calculate daily total", http.StatusInternalServerError)
		return
	}

	weeklyTotal, err := calculateWeeklyTotal(req.MID, currentDate)
	if err != nil {
		log.Printf("❌ Error calculating weekly total: %v", err)
		sendErrorResponse(w, "E004", "Failed to calculate weekly total", http.StatusInternalServerError)
		return
	}

	monthlyTotal, err := calculateMonthlyTotal(req.MID, currentDate)
	if err != nil {
		log.Printf("❌ Error calculating monthly total: %v", err)
		sendErrorResponse(w, "E005", "Failed to calculate monthly total", http.StatusInternalServerError)
		return
	}

	// Build response
	response := MerchantSummaryResponse{
		ErrorSchema: ErrorSchema{
			ErrorCode: "D000",
			ErrorMessage: map[string]string{
				"indonesian": "Berhasil",
				"english":    "Success",
			},
		},
		OutputSchema: OutputSchema{
			MerchantIDs:        req.MID,
			CurrentDate:        currentDateStr,
			TodayTotalAmount:   fmt.Sprintf("%d", todayTotal),
			WeeklyTotalAmount:  fmt.Sprintf("%d", weeklyTotal),
			MonthlyTotalAmount: fmt.Sprintf("%d", monthlyTotal),
		},
	}

	log.Printf("✅ Summary calculated - Today: %d, Weekly: %d, Monthly: %d",
		todayTotal, weeklyTotal, monthlyTotal)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ================================================================
// CALCULATION FUNCTIONS
// ================================================================

// Calculate daily total for all merchants
func calculateDailyTotal(merchantIDs []string, date time.Time) (int64, error) {
	var total int64

	dateStr := date.Format("2006-01-02")

	for _, mid := range merchantIDs {
		docID := fmt.Sprintf("%s_daily_%s", mid, dateStr)

		var summary MerchantSummary
		getResult, err := collection.Get(docID, nil)

		if err != nil {
			if err == gocb.ErrDocumentNotFound {
				log.Printf("   ⚠️  Daily summary not found for %s", docID)
				continue
			}
			return 0, fmt.Errorf("failed to get document %s: %w", docID, err)
		}

		if err := getResult.Content(&summary); err != nil {
			return 0, fmt.Errorf("failed to decode document %s: %w", docID, err)
		}

		total += summary.Amount
		log.Printf("   📈 Daily %s: %d", mid, summary.Amount)
	}

	return total, nil
}

// Calculate weekly total (Monday-Friday) for all merchants
func calculateWeeklyTotal(merchantIDs []string, date time.Time) (int64, error) {
	var total int64

	// Find Monday and Friday of current week
	dayOfWeek := date.Weekday()
	var daysToMonday int
	if dayOfWeek == time.Sunday {
		daysToMonday = 6
	} else {
		daysToMonday = int(dayOfWeek - time.Monday)
	}

	monday := date.AddDate(0, 0, -daysToMonday)
	friday := monday.AddDate(0, 0, 4)

	mondayStr := monday.Format("2006-01-02")
	fridayStr := friday.Format("2006-01-02")

	for _, mid := range merchantIDs {
		docID := fmt.Sprintf("%s_weekly_%s_%s", mid, mondayStr, fridayStr)

		var summary MerchantSummary
		getResult, err := collection.Get(docID, nil)

		if err != nil {
			if err == gocb.ErrDocumentNotFound {
				log.Printf("   ⚠️  Weekly summary not found for %s", docID)
				continue
			}
			return 0, fmt.Errorf("failed to get document %s: %w", docID, err)
		}

		if err := getResult.Content(&summary); err != nil {
			return 0, fmt.Errorf("failed to decode document %s: %w", docID, err)
		}

		total += summary.Amount
		log.Printf("   📈 Weekly %s: %d", mid, summary.Amount)
	}

	return total, nil
}

// Calculate monthly total for all merchants
func calculateMonthlyTotal(merchantIDs []string, date time.Time) (int64, error) {
	var total int64

	// Get last day of current month
	year, month, _ := date.Date()
	lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, date.Location())
	lastDayStr := lastDayOfMonth.Format("2006-01-02")

	for _, mid := range merchantIDs {
		docID := fmt.Sprintf("%s_monthly_%s", mid, lastDayStr)

		var summary MerchantSummary
		getResult, err := collection.Get(docID, nil)

		if err != nil {
			if err == gocb.ErrDocumentNotFound {
				log.Printf("   ⚠️  Monthly summary not found for %s", docID)
				continue
			}
			return 0, fmt.Errorf("failed to get document %s: %w", docID, err)
		}

		if err := getResult.Content(&summary); err != nil {
			return 0, fmt.Errorf("failed to decode document %s: %w", docID, err)
		}

		total += summary.Amount
		log.Printf("   📈 Monthly %s: %d", mid, summary.Amount)
	}

	return total, nil
}

// ================================================================
// HELPER FUNCTIONS
// ================================================================

// Send error response
func sendErrorResponse(w http.ResponseWriter, errorCode, message string, statusCode int) {
	response := MerchantSummaryResponse{
		ErrorSchema: ErrorSchema{
			ErrorCode: errorCode,
			ErrorMessage: map[string]string{
				"indonesian": message,
				"english":    message,
			},
		},
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
