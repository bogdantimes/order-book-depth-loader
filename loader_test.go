package order_book_depth_loader_test

import (
	"github.com/bogdantimes/order-book-depth-loader/depth"
	"github.com/kaz-yamam0t0/go-timeparser/timeparser"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const DateFmt = "m-d-Y"

func ParseOrDie(s string) time.Time {
	rangeStart, err := timeparser.ParseFormat(DateFmt, s)
	if err != nil {
		panic(err)
	}
	return *rangeStart
}

func TestLoader(t *testing.T) {
	depthLoader := depth.NewCCDepthLoader(depth.MarketBinance)

	result := depthLoader.Load([]depth.Pair{"BTC-BUSD"}, ParseOrDie("11-24-2022"), ParseOrDie("11-25-2022"))
	assert.NotEmpty(t, result["BTC-BUSD"])
	assert.Empty(t, result["ETH-BUSD"])

	minutesInDay := 24 * 60
	assert.Len(t, result["BTC-BUSD"], minutesInDay*4)

	result = depthLoader.Load([]depth.Pair{"ETH-BUSD"}, ParseOrDie("11-24-2022"), ParseOrDie("11-25-2022"))
	assert.NotEmpty(t, result["ETH-BUSD"])

	assert.Len(t, result["ETH-BUSD"], minutesInDay*4)
	assert.Len(t, result["BTC-BUSD"], minutesInDay*4)

	record1 := depthLoader.GetDepth("BTC-BUSD")
	assert.NotEmpty(t, record1.BidPrice)
	assert.NotEmpty(t, record1.AskPrice)
	assert.NotEmpty(t, record1.BidSize)
	assert.NotEmpty(t, record1.AskSize)

	depthLoader.Tick()

	record2 := depthLoader.GetDepth("BTC-BUSD")
	assert.NotEmpty(t, record2.BidPrice)
	assert.NotEmpty(t, record2.AskPrice)
	assert.NotEmpty(t, record2.BidSize)
	assert.NotEmpty(t, record2.AskSize)

	assert.NotEqual(t, record1.BidPrice, record2.BidPrice)

	result = depthLoader.Load([]depth.Pair{}, ParseOrDie("11-24-2022"), ParseOrDie("11-25-2022"))
	assert.Greater(t, len(result), 3)

	assert.FileExists(t, "data/2022-11-24_2022-11-25_depth.csv")
	assert.NoError(t, os.Remove("data/2022-11-24_2022-11-25_depth.csv"))
}
