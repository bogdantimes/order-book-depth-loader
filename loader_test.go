package order_book_depth_loader_test

import (
	"github.com/kaz-yamam0t0/go-timeparser/timeparser"
	"github.com/stretchr/testify/assert"
	"order-book-depth-loader/depth"
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
	if len(result) == 0 {
		t.Error("result is empty")
	}

	minutesInDay := 24 * 60
	if len(result["BTC-BUSD"]) != minutesInDay*4 {
		t.Error("wrong result")
	}

	result = depthLoader.Load([]depth.Pair{"ETH-BUSD"}, ParseOrDie("11-24-2022"), ParseOrDie("11-25-2022"))
	if len(result) == 0 {
		t.Error("result is empty")
	}

	if len(result["ETH-BUSD"]) != minutesInDay*4 {
		t.Error("wrong result")
	}

	if len(result["BTC-BUSD"]) != minutesInDay*4 {
		t.Error("wrong result")
	}

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
}
