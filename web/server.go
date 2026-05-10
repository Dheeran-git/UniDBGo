package web

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/unidb/unidb-go/parser"
	"github.com/unidb/unidb-go/planner"
	"github.com/unidb/unidb-go/security"
	"github.com/unidb/unidb-go/federation"
	"github.com/unidb/unidb-go/cache"
	"github.com/unidb/unidb-go/router"
)

var queryCache = cache.NewEngine()

// PipelineResponse is what the frontend expects
type PipelineResponse struct {
	Query     string                   `json:"query"`
	AST       *parser.QueryAST         `json:"ast"`
	Plan      *planner.ExecutionPlan   `json:"plan"`
	Decision  security.Decision        `json:"decision"`
	Results   []federation.Row         `json:"results,omitempty"`
	Metrics   PipelineMetrics          `json:"metrics"`
	Error     string                   `json:"error,omitempty"`
	CacheHit  bool                     `json:"cache_hit"`
}

type PipelineMetrics struct {
	ParseTimeMS    float64 `json:"parse_time_ms"`
	SecurityTimeMS float64 `json:"security_time_ms"`
	PlanTimeMS     float64 `json:"plan_time_ms"`
	ExecTimeMS     float64 `json:"exec_time_ms"`
	TotalTimeMS    float64 `json:"total_time_ms"`
}

type queryRequest struct {
	Query string `json:"query"`
}

// AI Router demo types
type aiRouterDemoRequest struct {
	Queries []string `json:"queries"`
}

type aiRouterPrediction struct {
	Query      string  `json:"query"`
	Database   string  `json:"database"`
	Dialect    string  `json:"dialect"`
	Confidence float64 `json:"confidence"`
}

type aiRouterDemoResponse struct {
	Predictions []aiRouterPrediction `json:"predictions"`
	TrainedOn   int                  `json:"trained_on"`
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	totalStart := time.Now()
	resp := PipelineResponse{Query: req.Query}

	// 1. Security Analysis
	secStart := time.Now()
	secEngine := &security.SecurityEngine{}
	resp.Decision = secEngine.Analyze(req.Query)
	resp.Metrics.SecurityTimeMS = float64(time.Since(secStart).Microseconds()) / 1000.0

	if resp.Decision.Block {
		resp.Error = "Query blocked by security engine: " + resp.Decision.Reason
		resp.Metrics.TotalTimeMS = float64(time.Since(totalStart).Microseconds()) / 1000.0
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 2. Parser
	parseStart := time.Now()
	ast, err := parser.Parse(req.Query)
	resp.Metrics.ParseTimeMS = float64(time.Since(parseStart).Microseconds()) / 1000.0
	if err != nil {
		resp.Error = fmt.Sprintf("Parse error: %v", err)
		resp.Metrics.TotalTimeMS = float64(time.Since(totalStart).Microseconds()) / 1000.0
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp.AST = ast

	// Check Cache
	if cachedRows, hit := queryCache.Get(req.Query, nil); hit {
		resp.CacheHit = true
		resp.Results = cachedRows
		resp.Metrics.TotalTimeMS = float64(time.Since(totalStart).Microseconds()) / 1000.0
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 3. Planner
	planStart := time.Now()
	plan, err := planner.GeneratePlan(ast)
	resp.Metrics.PlanTimeMS = float64(time.Since(planStart).Microseconds()) / 1000.0
	if err != nil {
		resp.Error = fmt.Sprintf("Plan error: %v", err)
		resp.Metrics.TotalTimeMS = float64(time.Since(totalStart).Microseconds()) / 1000.0
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp.Plan = plan

	// 4. Execution (Mocked for visualization purposes)
	execStart := time.Now()
	time.Sleep(50 * time.Millisecond) // artificially delay for mock execution
	// Create some generic mock results based on the tables
	var mockRows []federation.Row
	if len(ast.Tables) > 0 {
		row1 := make(federation.Row)
		row2 := make(federation.Row)
		for _, t := range ast.Tables {
			if t.Alias != "" {
				row1[t.Alias+".id"] = 1
				row2[t.Alias+".id"] = 2
				row1[t.Alias+".name"] = "Alice"
				row2[t.Alias+".name"] = "Bob"
			} else {
				row1["id"] = 1
				row2["id"] = 2
				row1["name"] = "Alice"
				row2["name"] = "Bob"
			}
		}
		mockRows = append(mockRows, row1, row2)
	}
	resp.Results = mockRows
	resp.Metrics.ExecTimeMS = float64(time.Since(execStart).Microseconds()) / 1000.0

	// Save to cache
	queryCache.Set(req.Query, nil, mockRows, 1 * time.Minute)

	resp.Metrics.TotalTimeMS = float64(time.Since(totalStart).Microseconds()) / 1000.0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	// Mock connection pool statistics
	stats := map[string]interface{}{
		"postgres": map[string]int{"active": 12, "idle": 8, "max": 20},
		"mysql":    map[string]int{"active": 4, "idle": 16, "max": 20},
		"mongodb":  map[string]int{"active": 1, "idle": 9, "max": 10},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleAIRouterDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req aiRouterDemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create a fresh model with seed data (no file dependency for demo)
	model := router.NewRoutingModel()
	seedData := []router.TrainingRecord{
		{Query: "SELECT * FROM users LIMIT 10", Database: "postgres", Dialect: "PostgreSQL"},
		{Query: "SELECT * FROM users WHERE ROWNUM <= 10", Database: "oracle", Dialect: "OracleSQL"},
		{Query: "SELECT * FROM users ORDER BY id DESC LIMIT 5", Database: "mysql", Dialect: "MySQL"},
		{Query: "SELECT JSON_EXTRACT(data, '$.name') FROM configs", Database: "mysql", Dialect: "MySQL"},
		{Query: "SELECT data->>'name' FROM configs", Database: "postgres", Dialect: "PostgreSQL"},
		{Query: `{ "find": "users", "filter": { "age": { "$gt": 25 } } }`, Database: "mongodb", Dialect: "MQL"},
	}
	model.Train(seedData)

	resp := aiRouterDemoResponse{
		TrainedOn: len(seedData),
	}

	for _, q := range req.Queries {
		prediction := model.Predict(q)
		
		// Normalize log-probability to a 0-1 confidence scale
		normalizedConf := 1.0 / (1.0 + math.Exp(-prediction.Confidence - 5.0))

		resp.Predictions = append(resp.Predictions, aiRouterPrediction{
			Query:      q,
			Database:   prediction.TargetDatabase,
			Dialect:    prediction.TargetDialect,
			Confidence: normalizedConf,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// StartServer starts the web server on the given port
func StartServer(port int) error {
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/query", handleQuery)
	http.HandleFunc("/api/stats", handleStats)
	http.HandleFunc("/api/schema", handleSchema)
	http.HandleFunc("/api/demo/ai-router", handleAIRouterDemo)

	fmt.Printf("Web Dashboard running at http://localhost:%d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handleSchema(w http.ResponseWriter, r *http.Request) {
	schema := map[string]interface{}{
		"postgres": []map[string]interface{}{
			{"name": "users", "columns": []string{"id", "name", "email", "created_at"}},
			{"name": "profiles", "columns": []string{"user_id", "bio", "avatar_url"}},
		},
		"mysql": []map[string]interface{}{
			{"name": "orders", "columns": []string{"id", "user_id", "total", "status"}},
			{"name": "items", "columns": []string{"id", "order_id", "product_id", "quantity"}},
		},
		"mongodb": []map[string]interface{}{
			{"name": "logs", "columns": []string{"_id", "level", "message", "timestamp"}},
			{"name": "sessions", "columns": []string{"_id", "user_id", "token", "expires"}},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}


