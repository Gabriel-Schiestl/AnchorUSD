package external

import "os"

type IPriceFeedAPI interface {
	GetEthUsdPrice() (string, error)
	GetBtcUsdPrice() (string, error)
}

type PriceFeedAPI struct {
	baseUrl string
}

func NewPriceFeedAPI() *PriceFeedAPI {
	return &PriceFeedAPI{
		baseUrl: os.Getenv("PRICE_FEED_API_URL"),
	}
}

// TODO: implement falling back to another price feed if the primary fails
func (pfa *PriceFeedAPI) GetEthUsdPrice() (string, error) {
	// Implementation to call the external price feed API and retrieve the ETH/USD price
	return "", nil
}

func (pfa *PriceFeedAPI) GetBtcUsdPrice() (string, error) {
	// Implementation to call the external price feed API and retrieve the BTC/USD price
	return "", nil
}
