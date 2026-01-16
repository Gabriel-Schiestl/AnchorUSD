package external

import "os"

type IPriceFeedAPI interface {
	GetEthUsdPrice() (float64, error)
	GetBtcUsdPrice() (float64, error)
}

type PriceFeedAPI struct {
	baseUrl string
}

func NewPriceFeedAPI() *PriceFeedAPI {
	return &PriceFeedAPI{
		baseUrl: os.Getenv("PRICE_FEED_API_URL"),
	}
}

//TODO: implement falling back to another price feed if the primary fails
func (pfa *PriceFeedAPI) GetEthUsdPrice() (float64, error) {
	// Implementation to call the external price feed API and retrieve the ETH/USD price
	return 0, nil
}

func (pfa *PriceFeedAPI) GetBtcUsdPrice() (float64, error) {
	// Implementation to call the external price feed API and retrieve the BTC/USD price
	return 0, nil
}

