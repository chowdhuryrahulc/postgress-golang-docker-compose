package exchangeapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	// "io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type PostgressExchangeDataStore struct {
	ID   int    `json:"id"`
	Date string `json:"date"`
	USD  int    `json:"usd"`
	EUR  int    `json:"eur"`
	GBP  int    `json:"gbp"`
}

// get last 10 exchange rates
func GetLast10ExchangeRates(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// SELECT * FROM ( SELECT * FROM postgress-exchange-data-store ORDER BY date DESC LIMIT 10
		rows, err := db.Query("SELECT * FROM postgress-exchange-data-store LIMIT 10")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		postgressExchangeDataStore := []PostgressExchangeDataStore{}
		for rows.Next() {
			var p PostgressExchangeDataStore
			if err := rows.Scan(&p.ID, &p.Date, &p.USD, &p.EUR, &p.GBP); err != nil {
				log.Fatal(err)
			}
			postgressExchangeDataStore = append(postgressExchangeDataStore, p)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(postgressExchangeDataStore)
	}
}

func GetTodayDate() string {
	t := time.Now()
	return t.Format("2006-01-02")
}

func GetDateTenDaysBack() string {
	t := time.Now().AddDate(0, 0, -10)
	return t.Format("2006-01-02")
}

func CreateHTTPRequest() (*http.Request, error) {
	baseURL := "https://api.apilayer.com/exchangerates_data/timeseries"
	// base=INR&symbols=USD,EUR,GBP&start_date=2022-05-01&end_date=2022-05-05"

	queryParams := url.Values{}
	queryParams.Add("base", "INR")
	queryParams.Add("symbols", "USD,EUR,GBP")
	queryParams.Add("start_date", GetDateTenDaysBack())
	queryParams.Add("end_date", GetTodayDate())

	// encode the query params and append them to the base URL
	url := baseURL + "?" + queryParams.Encode()

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", "0RhIgK5LbHXiIWnGJiUVAnnvbH2aihBS")
	return req, err
}

func GetApiResponse() *http.Response {
	client := &http.Client{}
	req, err := CreateHTTPRequest()
	if err != nil {
		fmt.Println("error creating exchange api url request: " + err.Error())
	}

	// Timeout added of 3 sec
	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Second)
	defer cancel()
	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Println("error fetching exchange api url response: " + err.Error())
	}
	// if res.Body != nil {
	// 	defer res.Body.Close()
	// }

	//
	// ex := &ExchangeAPIResponse{}
	// json.NewDecoder(res.Body).Decode(&ex)
	// fmt.Println("Decoded:", ex.Rates)

	//

	return res
}

type ExchangeAPIResponse struct {
	Success    bool            `json:"success"`
	Timeseries bool            `json:"timeseries"`
	StartDate  string          `json:"start_date"`
	EndDate    string          `json:"end_date"`
	Base       string          `json:"base"`
	Rates      map[string]Rate `json:"rates"`
}

type Rate struct {
	USD float64 `json:"USD"`
	EUR float64 `json:"EUR"`
	GBP float64 `json:"GBP"`
}

func StoreDataInPostgress(db *sql.DB, res *http.Response) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request){
		if res.Body != nil {
			defer res.Body.Close()
		}
	
		ex := &ExchangeAPIResponse{}
		json.NewDecoder(res.Body).Decode(&ex)
		fmt.Println("Decoded:", ex.Rates)
	
		//
		for date, currencies := range ex.Rates {
			err := db.QueryRow("INSERT INTO postgress-exchange-data-store (date, usd, eur, gbp) VALUES ($1, $2, $3, $4) RETURNING id", date, currencies.USD, currencies.EUR, currencies.GBP).Scan()
			if err != nil {
				fmt.Println("error while insersion")
			}
			fmt.Println(date)
			fmt.Println(currencies)
		}
	
		json.NewEncoder(w).Encode(ex)

	}
	
	// //
	// stmt, err := db.Prepare("INSERT INTO your_table(date, USD, EUR, GBP) VALUES (?, ?, ?, ?)")
	// if err != nil {
	// 	// handle error
	// }

	// for date, rates := range yourMap {
	// 	_, err := stmt.Exec(date, rates["USD"], rates["EUR"], rates["GBP"])
	// 	if err != nil {
	// 		// handle error
	// 	}
	// }
	// //

	// err := db.QueryRow("INSERT INTO postgress-exchange-data-store (date, usd, eur, gbp) VALUES ($1, $2, $3, $4) RETURNING id", , u.Email).Scan(&u.ID)
	// if err != nil {
	// 	log.Fatal(err)
	// }


}


func MarshalExchangeApiResponseToStruct(res []byte) *ExchangeAPIResponse {
	exchangeApiResponse := &ExchangeAPIResponse{}

	err := json.Unmarshal(res, exchangeApiResponse)
	if err != nil {
		fmt.Println("error unmarshaling exchange api url response: " + err.Error())
	}
	fmt.Println("Exp", exchangeApiResponse, "StartDate", exchangeApiResponse.Rates)
	return exchangeApiResponse
}
