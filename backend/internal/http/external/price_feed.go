package external

import (
	"encoding/json"
	"net/http"
	"os"
)

type IPriceFeedAPI interface {
	GetEthUsdPrice() (string, error)
	GetBtcUsdPrice() (string, error)
}

type PriceFeedAPI struct {
	baseUrl string
}

type PriceResult struct {
	Data struct {
		Amount string `json:"amount"`
		Base  string `json:"base"`
		Currency string `json:"currency"`
	} `json:"data"`
}

func NewPriceFeedAPI() *PriceFeedAPI {
	return &PriceFeedAPI{
		baseUrl: os.Getenv("PRICE_FEED_API_URL"),
	}
}

// TODO: implement falling back to another price feed if the primary fails
func (pfa *PriceFeedAPI) GetEthUsdPrice() (string, error) {
	res, err := http.Get(pfa.baseUrl+"/ETH-USD/spot")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var priceResult PriceResult
	err = json.NewDecoder(res.Body).Decode(&priceResult)
	if err != nil {
		return "", err
	}

	return priceResult.Data.Amount, nil
}

func (pfa *PriceFeedAPI) GetBtcUsdPrice() (string, error) {
	res, err := http.Get(pfa.baseUrl+"/BTC-USD/spot")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var priceResult PriceResult
	err = json.NewDecoder(res.Body).Decode(&priceResult)
	if err != nil {
		return "", err
	}

	return priceResult.Data.Amount, nil
}
