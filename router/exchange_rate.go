package router

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chynkm/ratesdb/currencystore"
	"github.com/chynkm/ratesdb/datastore"
	"github.com/chynkm/ratesdb/redisdb"
)

var (
	exchangeRateErr = map[string]string{
		"from_missing":     "The 'from' currency is missing in the query parameters",
		"to_missing":       "The 'to' currency is missing in the query parameters",
		"date_missing":     "The 'date' value is missing in the query parameters",
		"only_one_from":    "Only one 'from' currency is supported",
		"only_one_to":      "Only one 'to' currency is supported",
		"only_one_date":    "Only one 'date' value is supported",
		"unsupported_from": "The 'from' currency is unsupported",
		"unsupported_to":   "The 'to' currency is unsupported",
		"invalid_date":     "The 'date' value is invalid",
		"oldest_date":      "Only last " + strconv.Itoa(redisdb.Days) + " days exchange rates are supported",
		"future_date":      "Future date exchange rates are unavailable",
	}
	currencies map[string]int
)

type validationError struct {
	err     bool
	message string
}

func getExchangeRate(w http.ResponseWriter, req *http.Request) {
	v := validateGetExchangeRate(currencies, req.URL.Query())

	if !v.err {
		e := map[string][]map[string]interface{}{
			"errors": {
				{
					"status":  http.StatusUnprocessableEntity,
					"message": v.message,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(e)
		return
	}

	date, from, to := extractGetExchangeRateQueryParams(req.URL.Query())
	rate := redisdb.GetExchangeRate(date, from, to)
	json.NewEncoder(w).Encode(map[string]interface{}{"rate": rate, "status": 200})
}

// extractGetExchangeRateQueryParams retrieves the query params.
// returns the current date if no date is specified
func extractGetExchangeRateQueryParams(
	q map[string][]string,
) (string, string, string) {
	date := redisdb.LatestDate

	if _, ok := q["date"]; ok {
		date = q["date"][0]
	}

	return date, q["from"][0], q["to"][0]
}

// validateGetExchangeRate validate the API request
// q is URL query parameters
func validateGetExchangeRate(
	currencies map[string]int,
	q map[string][]string,
) *validationError {
	if _, ok := q["from"]; !ok {
		return &validationError{false, exchangeRateErr["from_missing"]}
	}
	if _, ok := q["to"]; !ok {
		return &validationError{false, exchangeRateErr["to_missing"]}
	}

	if len(q["from"]) > 1 {
		return &validationError{false, exchangeRateErr["only_one_from"]}
	}
	if len(q["to"]) > 1 {
		return &validationError{false, exchangeRateErr["only_one_to"]}
	}

	if date, ok := q["date"]; ok {
		if len(q["date"]) == 0 {
			return &validationError{false, exchangeRateErr["date_missing"]}
		}
		if len(q["date"]) > 1 {
			return &validationError{false, exchangeRateErr["only_one_date"]}
		}

		d, err := time.Parse(currencystore.DateLayout, date[0])
		if err != nil {
			return &validationError{false, exchangeRateErr["invalid_date"]}
		}

		lastDate := time.Now().AddDate(0, 0, -redisdb.Days).Format(currencystore.DateLayout)
		if d.Format(currencystore.DateLayout) < lastDate {
			return &validationError{false, exchangeRateErr["oldest_date"]}
		}
		futureDate := time.Now().AddDate(0, 0, 1).Format(currencystore.DateLayout)
		if d.Format(currencystore.DateLayout) >= futureDate {
			return &validationError{false, exchangeRateErr["future_date"]}
		}
	}

	if _, ok := currencies[q["from"][0]]; !ok {
		return &validationError{false, exchangeRateErr["unsupported_from"]}
	}

	if _, ok := currencies[q["to"][0]]; !ok {
		return &validationError{false, exchangeRateErr["unsupported_to"]}
	}

	return &validationError{true, ""}
}

// Routes holds all the routes supported by the application
func Routes() {
	currencies = datastore.GetCurrencies()
	http.HandleFunc("/v1/rates", getExchangeRate)

	log.Fatal(http.ListenAndServe(":8080", nil))
}