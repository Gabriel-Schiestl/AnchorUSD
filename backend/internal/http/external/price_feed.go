package external

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IPriceFeedAPI interface {
    GetEthUsdPrice() (string, error)
    GetBtcUsdPrice() (string, error)
}

type PriceFeedAPI struct {
    baseUrl         string
    fallbackBaseUrl string

    mu               sync.Mutex
    failureCount     int
    failureThreshold int
    cooldown         time.Duration
    openUntil        time.Time

    client *http.Client
}

type PriceResult struct {
    Data struct {
        Amount   string `json:"amount"`
        Base     string `json:"base"`
        Currency string `json:"currency"`
    } `json:"data"`
}

type FallbackPriceResult struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func NewPriceFeedAPI() *PriceFeedAPI {
    base := strings.TrimRight(os.Getenv("PRICE_FEED_API_URL"), "/")
    fallback := strings.TrimRight(os.Getenv("PRICE_FEED_FALLBACK_API_URL"), "/")

    threshold := 5
    if v := os.Getenv("PRICE_FEED_FAILURE_THRESHOLD"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            threshold = n
        }
    }

    coolSec := 60
    if v := os.Getenv("PRICE_FEED_COOLDOWN_SECONDS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            coolSec = n
        }
    }

    return &PriceFeedAPI{
        baseUrl:          base,
        fallbackBaseUrl:  fallback,
        failureThreshold: threshold,
        cooldown:         time.Duration(coolSec) * time.Second,
        client:           &http.Client{Timeout: 10 * time.Second},
    }
}

func (pfa *PriceFeedAPI) fetch(url, token string, isFallback bool) (string, error) {
    res, err := pfa.client.Get(url)
    if err != nil {
        return "", err
    }
    defer res.Body.Close()

    if res.StatusCode < 200 || res.StatusCode >= 300 {
        return "", fmt.Errorf("unexpected status %d", res.StatusCode)
    }

    if isFallback {
		var fallbackResult []FallbackPriceResult
		err = json.NewDecoder(res.Body).Decode(&fallbackResult)
		if err != nil {
			return "", err
		}
		for _, fr := range fallbackResult {
			if fr.Symbol == token {
				return fr.Price, nil
			}
		}
	}

	var priceResult PriceResult
    err = json.NewDecoder(res.Body).Decode(&priceResult)
    if err != nil {
        return "", err
    }

    return priceResult.Data.Amount, nil
}

func (pfa *PriceFeedAPI) callWithCircuit(path string) (string, error) {
    primary := strings.Replace(pfa.baseUrl, "[TOKEN]", path, 1)

    now := time.Now()

    pfa.mu.Lock()
    isOpen := now.Before(pfa.openUntil)
    pfa.mu.Unlock()

	fallbackToken := ""
	switch path {
	case "/ETH-USD":
		fallbackToken = "ETHUSD1"
	case "/BTC-USD":
		fallbackToken = "BTCUSD1"
	}

    if isOpen && pfa.fallbackBaseUrl != "" {
        if val, err := pfa.fetch(pfa.fallbackBaseUrl, fallbackToken, true); err == nil {
            return val, nil
        } else {
            return "", fmt.Errorf("circuit open - fallback failed: %w", err)
        }
    }

    val, err := pfa.fetch(primary, "", false)
    if err == nil {
        pfa.mu.Lock()
        pfa.failureCount = 0
        pfa.mu.Unlock()
        return val, nil
    }

    pfa.mu.Lock()
    pfa.failureCount++
    if pfa.failureCount >= pfa.failureThreshold {
        pfa.openUntil = time.Now().Add(pfa.cooldown)
    }
    pfa.mu.Unlock()

    if pfa.fallbackBaseUrl != "" {
        if val2, err2 := pfa.fetch(pfa.fallbackBaseUrl, fallbackToken, true); err2 == nil {
            return val2, nil
        } else {
            return "", fmt.Errorf("primary error: %v; fallback error: %v", err, err2)
        }
    }

    return "", err
}

func (pfa *PriceFeedAPI) GetEthUsdPrice() (string, error) {
    return pfa.callWithCircuit("/ETH-USD")
}

func (pfa *PriceFeedAPI) GetBtcUsdPrice() (string, error) {
    return pfa.callWithCircuit("/BTC-USD")
}