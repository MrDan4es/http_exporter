package internal

import (
	"fmt"
	"net/http"
)

type apiResponse struct {
	BalanceData struct {
		Balance      float64 `json:"balance"`
		BonusBalance float64 `json:"bonus_balance"`
		DaysLeft     int     `json:"days_left"`
		Detalization []struct {
			Linked []struct {
				Plan       string `json:"plan"`
				Price      string `json:"price"`
				PriceMonth string `json:"price_month"`
				ResourceID int    `json:"resource_id"`
				Type       string `json:"type"`
			} `json:"linked"`
			Name       string `json:"name"`
			Plan       string `json:"plan"`
			Price      string `json:"price"`
			PriceMonth string `json:"price_month"`
			ResourceID int    `json:"resource_id"`
			State      string `json:"state"`
			Type       string `json:"type"`
		} `json:"detalization"`
		HourlyCost  float64 `json:"hourly_cost"`
		HoursLeft   int     `json:"hours_left"`
		MonthlyCost float64 `json:"monthly_cost"`
	} `json:"balance_data"`
}

func GetRequest(token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://api.cloudvps.reg.ru/v1/balance_data", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return req, nil
}
