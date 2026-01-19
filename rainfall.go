package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// RainfallRecord represents a single rainfall measurement
type RainfallRecord struct {
	Date       time.Time `json:"date"`
	RainfallMM float64   `json:"rainfall_mm"`
}

// RainfallData represents the complete dataset
type RainfallData struct {
	Records []RainfallRecord `json:"records"`
}

type LabelledNumber struct {
	Period  string
	TotalMM float64
}

type PeriodTotals struct {
	Periods []LabelledNumber
}

// AverageComparison represents a comparison of average rainfall over a period with a specific average value
// This can be used to compare the average rainfall for a specific month or year against a known average
type AverageComparison struct {
	Period    string // e.g., "Jan", "2023"
	Average   float64
	LastTotal float64 // The value to compare against, e.g., last month
}

type ComparisonPeriodTotals struct {
	Periods []AverageComparison
}

var dataFile = "rainfall_data.json"
var rainfallData *RainfallData

// Custom unmarshaling to handle the nested JSON structure
func (rd *RainfallData) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to handle the year-based structure
	var yearData map[string][]struct {
		Date       string  `json:"date"`
		RainfallMM float64 `json:"rainfall_mm"`
	}

	if err := json.Unmarshal(data, &yearData); err != nil {
		return err
	}

	// Convert to our flat structure
	rd.Records = make([]RainfallRecord, 0)

	for _, yearRecords := range yearData {
		for _, record := range yearRecords {
			// Parse the date string
			date, err := time.Parse("2006-01-02", record.Date)
			if err != nil {
				log.Printf("Warning: could not parse date %s: %v", record.Date, err)
				continue
			}

			rd.Records = append(rd.Records, RainfallRecord{
				Date:       date,
				RainfallMM: record.RainfallMM,
			})
		}
	}
	return nil
}

func readJSONFile(filename string) (*RainfallData, error) {
	// Read the JSON file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %w", err)
	}
	// Parse the JSON into our struct
	var rainfallData RainfallData
	if err := json.Unmarshal(jsonData, &rainfallData); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	return &rainfallData, nil
}

func StartHandler(w http.ResponseWriter, r *http.Request) {
	// Read the JSON file to get all years (including those with no rainfall)
	jsonData, err := os.ReadFile(dataFile)
	if err != nil {
		http.Error(w, "Failed to read data file", http.StatusInternalServerError)
		return
	}

	// Parse JSON to get all years
	var yearData map[string][]struct {
		Date       string  `json:"date"`
		RainfallMM float64 `json:"rainfall_mm"`
	}
	if err := json.Unmarshal(jsonData, &yearData); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
		return
	}

	// Read and parse the rainfall data
	rainfallData, _ = readJSONFile(dataFile)
	// Print summary
	fmt.Printf("Loaded %d rainfall records\n", len(rainfallData.Records))

	// Get the set of yearly rainfall totals
	yearlyTotals := make(map[string]float64)

	// Initialize all years from the JSON (including those with 0 rainfall)
	for year := range yearData {
		yearlyTotals[year] = 0.0
	}

	// Now sum up the actual rainfall for each year
	for _, record := range rainfallData.Records {
		year := strconv.Itoa(record.Date.Year())
		yearlyTotals[year] += record.RainfallMM
	}

	// Convert the map to a slice for sorting
	yearlyTotalsSlice := make([]LabelledNumber, 0, len(yearlyTotals))
	for year, total := range yearlyTotals {
		yearlyTotalsSlice = append(yearlyTotalsSlice, LabelledNumber{
			Period:  year,
			TotalMM: total,
		})
	}
	// Sort the slice by year
	sort.Slice(yearlyTotalsSlice, func(i, j int) bool {
		return yearlyTotalsSlice[i].Period < yearlyTotalsSlice[j].Period
	})
	// Create the PeriodTotals struct to pass to the template
	yearlyTotalsStruct := PeriodTotals{
		Periods: yearlyTotalsSlice,
	}
	t, _ := template.ParseFiles("html/index.html")
	yearlyTotalsJSON, err := json.Marshal(yearlyTotalsStruct.Periods)
	if err != nil {
		http.Error(w, "Failed to marshal yearly totals", http.StatusInternalServerError)
		return
	}
	data := struct {
		Yearly template.JS
	}{
		Yearly: template.JS(yearlyTotalsJSON),
	}
	fmt.Printf("Rendering template with %d yearly totals\n", len(yearlyTotalsStruct.Periods))
	if t == nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

// MonthlyData handles the request for monthly rainfall data.
// It calculates the average rainfall per month and compares it with the last 12 months' totals.
// It then renders the results in a template.
// The monthly totals are calculated by aggregating the rainfall data by month and year.
// It also calculates the average rainfall for each month and compares it with the last 12 months
func MonthlyData(w http.ResponseWriter, r *http.Request) {
	// Read and parse the JSON data
	if rainfallData == nil {
		rainfallData, _ = readJSONFile(dataFile)
	}
	// Get the set of monthly rainfall totals
	monthlyTotals := make(map[string]float64)
	last12Months := make(map[string]float64)
	for _, record := range rainfallData.Records {
		monthKey := record.Date.Format("01") // Format as MM
		if _, exists := monthlyTotals[monthKey]; !exists {
			monthlyTotals[monthKey] = record.RainfallMM
		} else {
			monthlyTotals[monthKey] += record.RainfallMM
		}
		// Check if the month is in the last 12 months
		if record.Date.Year() == time.Now().Year() && record.Date.Month() == time.Now().Month() {
			if _, exists := last12Months[monthKey]; !exists {
				last12Months[monthKey] = record.RainfallMM
			} else {
				last12Months[monthKey] += record.RainfallMM
			}
		} else if record.Date.After(time.Now().AddDate(0, -12, 0)) && record.Date.Month() != time.Now().Month() {
			if _, exists := last12Months[monthKey]; !exists {
				last12Months[monthKey] = record.RainfallMM
			} else {
				last12Months[monthKey] += record.RainfallMM
			}
		}
	}
	monthlyAverages := make(map[string]AverageComparison)
	// Calculate the average rainfall per month
	// Count the number of unique years for each month to get the number of month totals
	monthlyCounts := make(map[string]int)
	yearsSeen := make(map[string]map[int]struct{})
	// Get a map of the previous 12 months inclusive in format "2025-06"
	for _, record := range rainfallData.Records {
		monthKey := record.Date.Format("01") // Format as MM
		year := record.Date.Year()
		if yearsSeen[monthKey] == nil {
			yearsSeen[monthKey] = make(map[int]struct{})
		}
		yearsSeen[monthKey][year] = struct{}{}
	}
	for month, years := range yearsSeen {
		monthlyCounts[month] = len(years)
	}
	for month, total := range monthlyTotals {
		if count, exists := monthlyCounts[month]; exists {
			ac := AverageComparison{Period: month, Average: total / float64(count), LastTotal: last12Months[month]} // Initialize LastTotal to 0.0
			// You may want to adjust the logic here for LastTotal as needed
			// For now, just set LastTotal to total for demonstration
			monthlyAverages[month] = ac
		} else {
			monthlyAverages[month] = AverageComparison{Period: month, Average: 0.0} // No data for this month
		}
	}
	// Print the monthly averages
	fmt.Println("\nAverage rainfall per month:")
	for month, average := range monthlyAverages {
		fmt.Printf("%s: %.1fmm vs %.1fmm\n", month, average.Average, average.LastTotal)
	}

	monthlyAveragesSlice := make([]AverageComparison, 0, len(monthlyAverages))
	for month, comparisonNos := range monthlyAverages {
		monthlyAveragesSlice = append(monthlyAveragesSlice, AverageComparison{
			Period:    month,
			Average:   comparisonNos.Average,
			LastTotal: comparisonNos.LastTotal,
		})
	}
	// Sort the slice by month
	sort.Slice(monthlyAveragesSlice, func(i, j int) bool {
		return monthlyAveragesSlice[i].Period < monthlyAveragesSlice[j].Period
	})
	// Create the PeriodTotals struct to pass to the template
	monthlyAveragesStruct := ComparisonPeriodTotals{
		Periods: monthlyAveragesSlice,
	}
	t, _ := template.ParseFiles("html/monthly.html")
	monthlyAveragesJSON, err := json.Marshal(monthlyAveragesStruct.Periods)
	if err != nil {
		http.Error(w, "Failed to marshal monthly averages", http.StatusInternalServerError)
		return
	}
	data := struct {
		Data template.JS
	}{
		Data: template.JS(monthlyAveragesJSON),
	}
	fmt.Printf("Rendering template with %d monthly totals\n", len(monthlyAveragesStruct.Periods))
	t.Execute(w, data)
}

func MonthVMonthHandler(w http.ResponseWriter, r *http.Request) {
	// This handler gets the totals for every month (YYYY-MM) and sorts them from highest to lowest so they can be compared
	fmt.Println("MonthVMonthHandler called")

	// Read the JSON file to get all years (including those with no rainfall)
	jsonData, err := os.ReadFile(dataFile)
	if err != nil {
		http.Error(w, "Failed to read data file", http.StatusInternalServerError)
		return
	}

	// Parse JSON to get all years
	var yearData map[string][]struct {
		Date       string  `json:"date"`
		RainfallMM float64 `json:"rainfall_mm"`
	}
	if err := json.Unmarshal(jsonData, &yearData); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
		return
	}

	// Read and parse the JSON data
	if rainfallData == nil {
		rainfallData, _ = readJSONFile(dataFile)
	}

	// Get the set of monthly rainfall totals
	monthlyTotals := make(map[string]float64)

	// Get current year and month
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Initialize all months from all years in the JSON (including those with 0 rainfall)
	for year := range yearData {
		yearInt, err := strconv.Atoi(year)
		if err != nil {
			continue
		}
		// Determine the last month to include for this year
		lastMonth := 12
		if yearInt == currentYear {
			lastMonth = currentMonth
		}
		// Create entries for months up to the cutoff
		for month := 1; month <= lastMonth; month++ {
			monthKey := fmt.Sprintf("%04d-%02d", yearInt, month)
			monthlyTotals[monthKey] = 0.0
		}
	}

	// Now sum up the actual rainfall for each month
	for _, record := range rainfallData.Records {
		monthKey := record.Date.Format("2006-01") // Format as YYYY-MM
		monthlyTotals[monthKey] += record.RainfallMM
	}

	// Convert the map to a slice for sorting
	monthlyTotalsSlice := make([]LabelledNumber, 0, len(monthlyTotals))
	for month, total := range monthlyTotals {
		monthlyTotalsSlice = append(monthlyTotalsSlice, LabelledNumber{
			Period:  month,
			TotalMM: total,
		})
	}
	// Sort the slice by total rainfall in descending order
	sort.Slice(monthlyTotalsSlice, func(i, j int) bool {
		return monthlyTotalsSlice[i].TotalMM < monthlyTotalsSlice[j].TotalMM
	})
	// Create the PeriodTotals struct to pass to the template
	monthlyTotalsStruct := PeriodTotals{
		Periods: monthlyTotalsSlice,
	}
	t, _ := template.ParseFiles("html/monthcomp.html")
	monthlyTotalsJSON, err := json.Marshal(monthlyTotalsStruct.Periods)
	if err != nil {
		http.Error(w, "Failed to marshal monthly totals", http.StatusInternalServerError)
		return
	}
	data := struct {
		Data template.JS
	}{
		Data: template.JS(monthlyTotalsJSON),
	}
	fmt.Printf("Rendering template with %d monthly totals\n", len(monthlyTotalsStruct.Periods))
	t.Execute(w, data)
}

// QuarterlyData handles the request for quarterly rainfall data.
// It calculates the average rainfall per quarter (Jan-Mar, Apr-Jun, Jul-Sep, Oct-Dec) and compares it with the last quarter's totals.
// It then renders the results in a template.
// The quarterly totals are calculated by aggregating the rainfall data by quarter and year.
// It also calculates the average rainfall for each quarter and compares it with the last quarter's totals.
func QuarterlyData(w http.ResponseWriter, r *http.Request) {
	// Read and parse the JSON data
	if rainfallData == nil {
		rainfallData, _ = readJSONFile(dataFile)
	}
	// Get the set of quarterly rainfall totals
	quarterlyTotals := make(map[string]float64)
	lastQuarterTotals := make(map[string]float64)
	q1Count := 0
	q2Count := 0
	q3Count := 0
	q4Count := 0
	for _, record := range rainfallData.Records {
		month := record.Date.Month()
		var quarter string
		// Count the number of years for each quarter
		q1Years := make(map[int]struct{})
		q2Years := make(map[int]struct{})
		q3Years := make(map[int]struct{})
		q4Years := make(map[int]struct{})

		for _, rec := range rainfallData.Records {
			year := rec.Date.Year()
			switch rec.Date.Month() {
			case time.January, time.February, time.March:
				q1Years[year] = struct{}{}
			case time.April, time.May, time.June:
				q2Years[year] = struct{}{}
			case time.July, time.August, time.September:
				q3Years[year] = struct{}{}
			case time.October, time.November, time.December:
				q4Years[year] = struct{}{}
			}
		}
		q1Count = len(q1Years)
		q2Count = len(q2Years)
		q3Count = len(q3Years)
		q4Count = len(q4Years)

		switch month {
		case time.January, time.February, time.March:
			quarter = "Q1"
		case time.April, time.May, time.June:
			quarter = "Q2"
		case time.July, time.August, time.September:
			quarter = "Q3"
		case time.October, time.November, time.December:
			quarter = "Q4"
		default:
			continue // Should not happen
		}
		if _, exists := quarterlyTotals[quarter]; !exists {
			quarterlyTotals[quarter] = record.RainfallMM
		} else {
			quarterlyTotals[quarter] += record.RainfallMM
		}
		// Only include the current year's total for the current quarter
		now := time.Now()
		currentQuarter := ""
		switch now.Month() {
		case time.January, time.February, time.March:
			currentQuarter = "Q1"
		case time.April, time.May, time.June:
			currentQuarter = "Q2"
		case time.July, time.August, time.September:
			currentQuarter = "Q3"
		case time.October, time.November, time.December:
			currentQuarter = "Q4"
		}
		if quarter == currentQuarter && record.Date.Year() == now.Year() {
			if _, exists := lastQuarterTotals[quarter]; !exists {
				lastQuarterTotals[quarter] = record.RainfallMM
			} else {
				lastQuarterTotals[quarter] += record.RainfallMM
			}
		} else if record.Date.After(time.Now().AddDate(-1, 0, 0)) && quarter != currentQuarter {
			if _, exists := lastQuarterTotals[quarter]; !exists {
				lastQuarterTotals[quarter] = record.RainfallMM
			} else {
				lastQuarterTotals[quarter] += record.RainfallMM
			}
		}
	}
	fmt.Println("\nQuarter totals:")
	for quarter, total := range quarterlyTotals {
		fmt.Printf("%s: %.1fmm\n", quarter, total)
	}
	fmt.Println("\nLast quarter:")
	for quarter, total := range lastQuarterTotals {
		fmt.Printf("%s: %.1fmm\n", quarter, total)
	}
	quarterlyAverages := make(map[string]AverageComparison)
	for quarter, total := range quarterlyTotals {
		if quarter == "Q1" {
			ac := AverageComparison{Period: quarter, Average: total / float64(q1Count), LastTotal: lastQuarterTotals[quarter]}
			quarterlyAverages[quarter] = ac
			continue
		}
		if quarter == "Q2" {
			ac := AverageComparison{Period: quarter, Average: total / float64(q2Count), LastTotal: lastQuarterTotals[quarter]}
			quarterlyAverages[quarter] = ac
			continue
		}
		if quarter == "Q3" {
			ac := AverageComparison{Period: quarter, Average: total / float64(q3Count), LastTotal: lastQuarterTotals[quarter]}
			quarterlyAverages[quarter] = ac
			continue
		}
		if quarter == "Q4" {
			ac := AverageComparison{Period: quarter, Average: total / float64(q4Count), LastTotal: lastQuarterTotals[quarter]}
			quarterlyAverages[quarter] = ac
			continue
		}
	}
	fmt.Println("\nAverage rainfall per quarter:")
	for quarter, average := range quarterlyAverages {
		fmt.Printf("%s: %.1fmm vs %.1fmm\n", quarter, average.Average, average.LastTotal)
	}

	quarterlyAveragesSlice := make([]AverageComparison, 0, len(quarterlyAverages))
	for quarter, comparisonNos := range quarterlyAverages {
		quarterlyAveragesSlice = append(quarterlyAveragesSlice, AverageComparison{
			Period:    quarter,
			Average:   comparisonNos.Average,
			LastTotal: comparisonNos.LastTotal,
		})
	}
	// Sort the slice by quarter
	sort.Slice(quarterlyAveragesSlice, func(i, j int) bool {
		return quarterlyAveragesSlice[i].Period < quarterlyAveragesSlice[j].Period
	})
	// Create the PeriodTotals struct to pass to the template
	quarterlyAveragesStruct := ComparisonPeriodTotals{
		Periods: quarterlyAveragesSlice,
	}
	t, _ := template.ParseFiles("html/quarterly.html")
	quarterlyAveragesJSON, err := json.Marshal(quarterlyAveragesStruct.Periods)
	if err != nil {
		http.Error(w, "Failed to marshal quarterly averages", http.StatusInternalServerError)
		return
	}
	data := struct {
		Data template.JS
	}{
		Data: template.JS(quarterlyAveragesJSON),
	}
	fmt.Printf("Rendering template with %d quarterly averages\n", len(quarterlyAveragesStruct.Periods))
	t.Execute(w, data)
}

// This handler gets the totals for every quarter (2004-Q1, 2004-Q2 etc.) and sorts them from highest to lowest so they can be compared
func QuarterVQuarterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("QuarterVQuarterHandler called")
	// Read and parse the JSON data
	if rainfallData == nil {
		rainfallData, _ = readJSONFile(dataFile)
	}
	// Get the set of quarterly rainfall totals
	quarterlyTotals := make(map[string]float64)
	for _, record := range rainfallData.Records {
		year := record.Date.Year()
		var quarter string
		switch record.Date.Month() {
		case time.January, time.February, time.March:
			quarter = fmt.Sprintf("%d-Q1", year)
		case time.April, time.May, time.June:
			quarter = fmt.Sprintf("%d-Q2", year)
		case time.July, time.August, time.September:
			quarter = fmt.Sprintf("%d-Q3", year)
		case time.October, time.November, time.December:
			quarter = fmt.Sprintf("%d-Q4", year)
		default:
			continue // Should not happen
		}
		if _, exists := quarterlyTotals[quarter]; !exists {
			quarterlyTotals[quarter] = record.RainfallMM
		} else {
			quarterlyTotals[quarter] += record.RainfallMM
		}
	}
	// Convert the map to a slice for sorting
	quarterlyTotalsSlice := make([]LabelledNumber, 0, len(quarterlyTotals))
	for quarter, total := range quarterlyTotals {
		quarterlyTotalsSlice = append(quarterlyTotalsSlice, LabelledNumber{
			Period:  quarter,
			TotalMM: total,
		})
	}
	// Sort the slice by total rainfall in ascending order
	sort.Slice(quarterlyTotalsSlice, func(i, j int) bool {
		return quarterlyTotalsSlice[i].TotalMM < quarterlyTotalsSlice[j].TotalMM
	})
	// Create the PeriodTotals struct to pass to the template
	quarterlyTotalsStruct := PeriodTotals{
		Periods: quarterlyTotalsSlice,
	}
	t, _ := template.ParseFiles("html/quartercomp.html")
	quarterlyTotalsJSON, err := json.Marshal(quarterlyTotalsStruct.Periods)
	if err != nil {
		http.Error(w, "Failed to marshal quarterly totals", http.StatusInternalServerError)
		return
	}
	data := struct {
		Data template.JS
	}{
		Data: template.JS(quarterlyTotalsJSON),
	}
	fmt.Printf("Rendering template with %d quarterly totals\n", len(quarterlyTotalsStruct.Periods))
	t.Execute(w, data)
}

// HalfYearHandler handles the request for half-year rainfall data.
// It calculates the average rainfall per half-year (H1: Jan-Jun, H2: Jul-Dec) and compares these averages with the last two half-yearly totals.
// It then renders the results in a template.
func HalfYearHandler(w http.ResponseWriter, r *http.Request) {
	// Read and parse the JSON data
	if rainfallData == nil {
		rainfallData, _ = readJSONFile(dataFile)
	}
	// Calculate the average rainfall per half-year (H1: Jan-Jun, H2: Jul-Dec) and compare with the last two half-yearly totals inclusive (i.e., the current half-year and the previous half-year).
	halfYearTotals := make(map[string]float64)
	lastHalfYearTotals := make(map[string]float64)
	h1Years := make(map[int]struct{})
	h2Years := make(map[int]struct{})

	for _, record := range rainfallData.Records {
		year := record.Date.Year()
		month := record.Date.Month()
		var half string
		var currentHalf string
		if month >= time.January && month <= time.June {
			half = "H1"
			h1Years[year] = struct{}{}
		} else {
			half = "H2"
			h2Years[year] = struct{}{}
		}
		// Determine the current half-year based on the current date
		now := time.Now()
		if now.Month() >= time.January && now.Month() <= time.June {
			currentHalf = "H1"
		} else {
			currentHalf = "H2"
		}
		if half == currentHalf && record.Date.Year() == now.Year() {
			if _, exists := lastHalfYearTotals[half]; !exists {
				lastHalfYearTotals[half] = record.RainfallMM
			} else {
				lastHalfYearTotals[half] += record.RainfallMM
			}
		} else if record.Date.After(time.Now().AddDate(0, -12, 0)) && half != currentHalf {
			if _, exists := lastHalfYearTotals[half]; !exists {
				lastHalfYearTotals[half] = record.RainfallMM
			} else {
				lastHalfYearTotals[half] += record.RainfallMM
			}
		}
		if _, exists := halfYearTotals[half]; !exists {
			halfYearTotals[half] = record.RainfallMM
		} else {
			halfYearTotals[half] += record.RainfallMM
		}
	}
	// Print the half-year totals
	fmt.Println("\nHalf-year totals:")
	for half, total := range halfYearTotals {
		fmt.Printf("%s: %.1fmm\n", half, total)
	}
	fmt.Println("\nLast half-year totals:")
	for half, total := range lastHalfYearTotals {
		fmt.Printf("%s: %.1fmm\n", half, total)
	}
	h1Count := len(h1Years)
	h2Count := len(h2Years)

	halfYearAverages := make(map[string]AverageComparison)
	for half, total := range halfYearTotals {
		var count int
		if half == "H1" {
			count = h1Count
		} else {
			count = h2Count
		}
		ac := AverageComparison{
			Period:    half,
			Average:   total / float64(count),
			LastTotal: lastHalfYearTotals[half],
		}
		halfYearAverages[half] = ac
	}

	fmt.Println("\nAverage rainfall per half-year:")
	for half, average := range halfYearAverages {
		fmt.Printf("%s: %.1fmm vs %.1fmm\n", half, average.Average, average.LastTotal)
	}

	halfYearAveragesSlice := make([]AverageComparison, 0, len(halfYearAverages))
	for half, comparisonNos := range halfYearAverages {
		halfYearAveragesSlice = append(halfYearAveragesSlice, AverageComparison{
			Period:    half,
			Average:   comparisonNos.Average,
			LastTotal: comparisonNos.LastTotal,
		})
	}
	// Sort the slice by H1, H2
	sort.Slice(halfYearAveragesSlice, func(i, j int) bool {
		return halfYearAveragesSlice[i].Period < halfYearAveragesSlice[j].Period
	})

	halfYearAveragesStruct := ComparisonPeriodTotals{
		Periods: halfYearAveragesSlice,
	}
	t, _ := template.ParseFiles("html/halfyear.html")
	halfYearAveragesJSON, err := json.Marshal(halfYearAveragesStruct.Periods)
	if err != nil {
		http.Error(w, "Failed to marshal half-year averages", http.StatusInternalServerError)
		return
	}
	data := struct {
		Data template.JS
	}{
		Data: template.JS(halfYearAveragesJSON),
	}
	fmt.Printf("Rendering template with %d half-year averages\n", len(halfYearAveragesStruct.Periods))
	t.Execute(w, data)
}

// HalfYearVHalfYearHandler handles the request for comparing half-year rainfall totals (e.g., 2022-H1, 2022-H2, etc.) sorted by total rainfall.
func HalfYearVHalfYearHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HalfYearVHalfYearHandler called")
	// Read and parse the JSON data
	if rainfallData == nil {
		rainfallData, _ = readJSONFile(dataFile)
	}
	// Aggregate rainfall by half-year (H1: Jan-Jun, H2: Jul-Dec) for each year
	halfYearTotals := make(map[string]float64)
	for _, record := range rainfallData.Records {
		year := record.Date.Year()
		var half string
		switch record.Date.Month() {
		case time.January, time.February, time.March, time.April, time.May, time.June:
			half = fmt.Sprintf("%d-H1", year)
		case time.July, time.August, time.September, time.October, time.November, time.December:
			half = fmt.Sprintf("%d-H2", year)
		default:
			continue // Should not happen
		}
		halfYearTotals[half] += record.RainfallMM
	}
	// Convert the map to a slice for sorting
	halfYearTotalsSlice := make([]LabelledNumber, 0, len(halfYearTotals))
	for half, total := range halfYearTotals {
		halfYearTotalsSlice = append(halfYearTotalsSlice, LabelledNumber{
			Period:  half,
			TotalMM: total,
		})
	}
	// Sort the slice by total rainfall in ascending order
	sort.Slice(halfYearTotalsSlice, func(i, j int) bool {
		return halfYearTotalsSlice[i].TotalMM < halfYearTotalsSlice[j].TotalMM
	})
	// Create the PeriodTotals struct to pass to the template
	halfYearTotalsStruct := PeriodTotals{
		Periods: halfYearTotalsSlice,
	}
	t, _ := template.ParseFiles("html/halfyearcomp.html")
	halfYearTotalsJSON, err := json.Marshal(halfYearTotalsStruct.Periods)
	if err != nil {
		http.Error(w, "Failed to marshal half-year totals", http.StatusInternalServerError)
		return
	}
	data := struct {
		Data template.JS
	}{
		Data: template.JS(halfYearTotalsJSON),
	}
	fmt.Printf("Rendering template with %d half-year totals\n", len(halfYearTotalsStruct.Periods))
	t.Execute(w, data)
}

func YearCompHandler(w http.ResponseWriter, r *http.Request) {
	// Read the JSON file to get all years (including those with no rainfall)
	jsonData, err := os.ReadFile(dataFile)
	if err != nil {
		http.Error(w, "Failed to read data file", http.StatusInternalServerError)
		return
	}

	// Parse JSON to get all years
	var yearData map[string][]struct {
		Date       string  `json:"date"`
		RainfallMM float64 `json:"rainfall_mm"`
	}
	if err := json.Unmarshal(jsonData, &yearData); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
		return
	}

	// Read and parse the JSON data
	if rainfallData == nil {
		rainfallData, _ = readJSONFile(dataFile)
	}
	// Get the set of yearly rainfall totals
	yearlyTotals := make(map[string]float64)

	// Initialize all years from the JSON (including those with 0 rainfall)
	for year := range yearData {
		yearlyTotals[year] = 0.0
	}

	// Now sum up the actual rainfall for each year
	for _, record := range rainfallData.Records {
		year := strconv.Itoa(record.Date.Year())
		yearlyTotals[year] += record.RainfallMM
	}

	// Convert the map to a slice for sorting
	yearlyTotalsSlice := make([]LabelledNumber, 0, len(yearlyTotals))
	for year, total := range yearlyTotals {
		yearlyTotalsSlice = append(yearlyTotalsSlice, LabelledNumber{
			Period:  year,
			TotalMM: total,
		})
	}
	// Sort the slice by total rainfall in ascending order
	sort.Slice(yearlyTotalsSlice, func(i, j int) bool {
		return yearlyTotalsSlice[i].TotalMM < yearlyTotalsSlice[j].TotalMM
	})
	yearlyTotalsStruct := PeriodTotals{
		Periods: yearlyTotalsSlice,
	}
	t, _ := template.ParseFiles("html/yearcomp.html")
	yearlyTotalsJSON, err := json.Marshal(yearlyTotalsStruct.Periods)
	if err != nil {
		http.Error(w, "Failed to marshal yearly totals", http.StatusInternalServerError)
		return
	}
	data := struct {
		Data template.JS
	}{
		Data: template.JS(yearlyTotalsJSON),
	}
	fmt.Printf("Rendering template with %d yearly totals\n", len(yearlyTotalsStruct.Periods))
	t.Execute(w, data)
}

// YearDailyRunningTotals holds the running daily totals for a year
type YearDailyRunningTotals struct {
	Year   int
	Totals []LabelledNumber
}

// Helper function to check if a year is a leap year
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func YearProgressHandler(w http.ResponseWriter, r *http.Request) {
	// Read and parse the JSON data
	if rainfallData == nil {
		rainfallData, _ = readJSONFile(dataFile)
	}
	// Map: year -> sorted slice of RainfallRecord
	yearRecords := make(map[int][]RainfallRecord)
	for _, record := range rainfallData.Records {
		year := record.Date.Year()
		yearRecords[year] = append(yearRecords[year], record)
	}
	// For each year, sort by date and compute running total
	var yearlyProgressSlice []YearDailyRunningTotals
	for year, records := range yearRecords {
		sort.Slice(records, func(i, j int) bool {
			return records[i].Date.Before(records[j].Date)
		})
		var runningTotal float64
		var totals []LabelledNumber

		// Build a map of date string to rainfall for quick lookup
		rainByDate := make(map[string]float64)
		for _, rec := range records {
			dateStr := rec.Date.Format("2006-01-02")
			rainByDate[dateStr] = rec.RainfallMM
		}

		// Find the first and last date for the year
		startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(year, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
		// Handle leap year edge case for Feb 29
		if time.Now().Month() == time.February && time.Now().Day() == 29 {
			if !isLeapYear(year) {
				endDate = time.Date(year, time.February, 28, 0, 0, 0, 0, time.UTC)
			}
		}

		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			rain := rainByDate[dateStr]
			runningTotal += rain
			totals = append(totals, LabelledNumber{
				Period:  d.Format("01-02"), // MM-DD
				TotalMM: runningTotal,
			})
		}

		yearlyProgressSlice = append(yearlyProgressSlice, YearDailyRunningTotals{
			Year:   year,
			Totals: totals,
		})
	}
	// Sort years ascending
	sort.Slice(yearlyProgressSlice, func(i, j int) bool {
		return yearlyProgressSlice[i].Year < yearlyProgressSlice[j].Year
	})
	yearlyProgressStruct := struct {
		Years []YearDailyRunningTotals
	}{
		Years: yearlyProgressSlice,
	}
	t, err := template.ParseFiles("html/yearprogress.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}
	yearlyProgressJSON, err := json.Marshal(yearlyProgressStruct.Years)
	if err != nil {
		http.Error(w, "Failed to marshal yearly progress", http.StatusInternalServerError)
		return
	}
	data := struct {
		Data template.JS
	}{
		Data: template.JS(yearlyProgressJSON),
	}
	fmt.Printf("Rendering template with %d years of daily running totals\n", len(yearlyProgressStruct.Years))
	// Print actual data for 2025
	// fmt.Println("2025 daily rainfall:")
	// for _, yearTotals := range yearlyProgressStruct.Years {
	// 	if yearTotals.Year == 2025 {
	// 		for _, daily := range yearTotals.Totals {
	// 			fmt.Printf("%s: %.1fmm\n", daily.Period, daily.TotalMM)
	// 		}
	// 	}
	// }
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}

func main() {
	fmt.Printf("Starting server at port 6655\n")
	var dir string
	flag.StringVar(&dir, "dir", ".", "")
	flag.Parse()
	r := mux.NewRouter()
	r.PathPrefix("/html").Handler(http.StripPrefix("/", http.FileServer(http.Dir(dir))))
	r.HandleFunc("/", StartHandler)
	r.HandleFunc("/yearcomp/", YearCompHandler)
	r.HandleFunc("/yearprogress/", YearProgressHandler)
	r.HandleFunc("/monthly/", MonthlyData)
	r.HandleFunc("/quarterly/", QuarterlyData)
	r.HandleFunc("/monthcomp/", MonthVMonthHandler)
	r.HandleFunc("/quartercomp/", QuarterVQuarterHandler)
	r.HandleFunc("/halfyear/", HalfYearHandler)
	r.HandleFunc("/halfyearcomp/", HalfYearVHalfYearHandler)
	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:6655",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
