package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/amaumene/snowfinder-common/models"
	"github.com/amaumene/snowfinder-common/repository"
)

type Handler struct {
	repo      repository.Reader
	templates *template.Template
}

type ResortResult struct {
	Rank            int      `json:"rank"`
	Name            string   `json:"name"`
	Prefecture      string   `json:"prefecture"`
	AvgSnowfall     *int     `json:"avg_snowfall"`
	YearsWithData   *int     `json:"years_with_data"`
	TopElevation    *int     `json:"top_elevation"`
	BaseElevation   *int     `json:"base_elevation"`
	VerticalDrop    *int     `json:"vertical_drop"`
	NumCourses      *int     `json:"num_courses"`
	LongestCourseKM *float64 `json:"longest_course_km"`
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"code":  code,
	})
}

func NewHandler(repo repository.Reader) (*Handler, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		repo:      repo,
		templates: tmpl,
	}, nil
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	err := h.templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("ERROR: render index template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) AboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	err := h.templates.ExecuteTemplate(w, "about.html", nil)
	if err != nil {
		log.Printf("ERROR: render about template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	startDate := query.Get("start_date")
	endDate := query.Get("end_date")
	prefecture := query.Get("prefecture")
	limitStr := query.Get("limit")

	if startDate == "" {
		jsonError(w, "start_date is required in MM-DD format", http.StatusBadRequest)
		return
	}

	if !isValidMonthDay(startDate) {
		jsonError(w, "start_date must be in MM-DD format (e.g., 02-08)", http.StatusBadRequest)
		return
	}

	if endDate == "" {
		endDate = startDate
	} else if !isValidMonthDay(endDate) {
		jsonError(w, "end_date must be in MM-DD format (e.g., 02-14)", http.StatusBadRequest)
		return
	}

	limit := 10
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.searchSnowiestResorts(ctx, startDate, endDate, prefecture, limit)
	if err != nil {
		log.Printf("ERROR: search resorts: %v", err)
		jsonError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	results := h.convertToResortResults(stats)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("ERROR: encode search results: %v", err)
	}
}

func (h *Handler) searchSnowiestResorts(ctx context.Context, startDate, endDate, prefecture string, limit int) ([]models.WeeklyResortStats, error) {
	if prefecture == "" || prefecture == "all" {
		return h.repo.GetSnowiestResortsForDateRange(ctx, startDate, endDate, limit)
	}
	return h.repo.GetSnowiestResortsForDateRangeByPrefecture(ctx, startDate, endDate, prefecture, limit)
}

func (h *Handler) convertToResortResults(stats []models.WeeklyResortStats) []ResortResult {
	results := make([]ResortResult, len(stats))
	for i, stat := range stats {
		results[i] = ResortResult{
			Rank:            i + 1,
			Name:            stat.Name,
			Prefecture:      stat.Prefecture,
			AvgSnowfall:     stat.TotalSnowfall,
			YearsWithData:   stat.YearsWithData,
			TopElevation:    stat.TopElevationM,
			BaseElevation:   stat.BaseElevationM,
			VerticalDrop:    stat.VerticalM,
			NumCourses:      stat.NumCourses,
			LongestCourseKM: stat.LongestCourseKM,
		}
	}
	return results
}

func (h *Handler) ResortsWithPeaksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resortsWithPeaks, err := h.repo.GetAllResortsWithPeaks(ctx)
	if err != nil {
		log.Printf("ERROR: get resorts with peaks: %v", err)
		jsonError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Simplify response for dropdown - just ID and name
	type ResortOption struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	options := make([]ResortOption, len(resortsWithPeaks))
	for i, rp := range resortsWithPeaks {
		options[i] = ResortOption{
			ID:   rp.Resort.ID,
			Name: rp.Resort.Name,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(options); err != nil {
		log.Printf("ERROR: encode resort options: %v", err)
	}
}

func (h *Handler) PeakInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	resortID := r.URL.Query().Get("resort_id")

	// If no resort_id or "all", return all resorts with peaks
	if resortID == "" || resortID == "all" {
		resortsWithPeaks, err := h.repo.GetAllResortsWithPeaks(ctx)
		if err != nil {
			log.Printf("ERROR: get all resorts with peaks: %v", err)
			jsonError(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resortsWithPeaks); err != nil {
			log.Printf("ERROR: encode resorts with peaks: %v", err)
		}
		return
	}

	// Get specific resort with peaks
	resort, err := h.repo.GetResortByID(ctx, resortID)
	if err != nil {
		jsonError(w, "Resort not found", http.StatusNotFound)
		return
	}

	peaks, err := h.repo.GetPeakPeriodsForResort(ctx, resort.ID)
	if err != nil {
		log.Printf("ERROR: get peak periods for resort %s: %v", resort.ID, err)
		jsonError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := []models.ResortWithPeaks{{
		Resort: *resort,
		Peaks:  peaks,
	}}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: encode peak info response: %v", err)
	}
}

func isValidMonthDay(monthDay string) bool {
	if len(monthDay) != 5 || monthDay[2] != '-' {
		return false
	}

	month, err1 := strconv.Atoi(monthDay[0:2])
	day, err2 := strconv.Atoi(monthDay[3:5])

	if err1 != nil || err2 != nil {
		return false
	}

	if month < 1 || month > 12 {
		return false
	}

	if day < 1 || day > 31 {
		return false
	}

	daysInMonth := map[int]int{
		1: 31, 2: 29, 3: 31, 4: 30, 5: 31, 6: 30,
		7: 31, 8: 31, 9: 30, 10: 31, 11: 30, 12: 31,
	}

	if day > daysInMonth[month] {
		return false
	}

	return true
}
