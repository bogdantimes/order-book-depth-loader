package depth

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/life4/genesis/slices"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Loader downloads the depth data from the crypto-chassis API
//
// It takes a Pair and startDate, and endDate.
// Then it iterates from start to end date, and downloads the depth data for each day.
// To download the depth data, it uses the public crypto-chassis API. Example to load 1 day:
// https://api.cryptochassis.com/v1/market-depth/binance/btc-busd?startTime=2021-10-10
// The response is in the following format:
//
//	{
//	 "urls": [
//	   {
//	     "startTime": {
//	       "seconds": 1633824000,
//	       "iso": "2021-10-10T00:00:00.000Z"
//	     },
//	     "endTime": {
//	       "seconds": 1633910400,
//	       "iso": "2021-10-11T00:00:00.000Z"
//	     },
//	     "url": "https://marketdata-e0323a9039add2978bf5b49550572c7c.s3.amazonaws.com/v2/market_depth/bn/btc-busd/1-1633824000.csv.gz?AWSAccessKeyId=ASIATPNB7YZIWD73JVDZ&Expires=1669398105&Signature=G6t5obQajMeD4w%2BffdIn4vEaABE%3D&x-amz-security-token=IQoJb3JpZ2luX2VjEOn%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaCXVzLWVhc3QtMSJHMEUCIQD8Xb6ethZ1G8E36g6kBGD%2Frz8kFpi42Ih9IwnFjQ4RcQIgdm1RG87cGCIfhVH6biZChXCk%2Fa9W2GaTjoISPN4Fvscq1QQI8f%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FARADGgwyMzkyNDcyNzk2OTciDJgEsOkTOI6bSvsO3CqpBAzzDxbMqFqN0dWh4fuqb8TEeiV0A4qNvEMR%2BrWhHnLiIAVCjxaBc5dspa3Ap%2Br3bj8clwVU0Uh%2F9YRWJb6tT3iX%2FcGc9EQnJ1R2y6%2FZzjzzF8Zv3ofMbudf2VgV4RYHHs2kNB4z700S%2BfFiA%2FFHfGq4pV7I5KCtmMBhmMH2oAvDQ02YHXnqbdGaBCQLEfMwISGqPgsqTdC8GpNyhS1RJEj29AOZAy%2Bg8%2BTmmijSO3NSx332q5oXbIzPuizBLdlZHDseLNsp%2B1e%2FPgOHMJZkCBAExLC5ax3W48Ko6MS1xiBnQclU9hvabBLTrvDdSStZ8lCY%2F3TndBn3Gvkyb9EhFhj6Xg63MG6EIZBARsEWZ6FzRRNySC8TCFEiArCKTn7WYyi%2FCNhZPLKaVB4FHA6n%2BAZEvJNwTLR1%2FUVvrYgBAnH4S3pIu4ZvyMoORx2Y0t1FuT8Dbc3V8n5LScp%2FqdoWhybaIVU4ZKhnUW1N7jgS%2BZhF2at%2BZ%2B2%2BYX%2Bkc8jn2WFWgefoJwE%2BKq2B5l3%2FH4I39ZLFidRH%2FHqoIPCfMYgnOAO%2FIb%2Bo8ANgDlguYo4P7nzvaY681HjboYJlpHxCXrwuxP5V3Cq5RdUMlZSiasuAyUrBJEOWDSG2oBevbiO0adBP%2FIg9UGPuz4m4hjX9zbHIDF%2B6jRbIn8j0WCR4pk58%2BfCIZvzCpm7EUDnvdy%2FnxFsv6UQEozYywjKe0%2FfrVDjD76ptECmn%2BsgD85kw89aDnAY6qQHMuPRvmUnle%2B1UR58vvwf0psmC%2FoEmESqOSNKTKHk6vU4842T3iDQTH71OMQb8cKEXFevBf7z2iA8xTroGMoC9Vytf5gi%2B5pgbGebEMfQRymrib7XocCnESrLvLvv%2BEtq4B7s7ycWUQyH9uAnBxmoPjlKEpLJ1GtH%2FCVkVVpvTBIaWcbXuCnIlPeOHAlov9Dv7KZgiGmukOAaF5sjbWgXOo5heo7kmOoIW"
//	   }
//	 ],
//	 "expiration": "300 seconds"
//	}
//
// Using the url, it downloads the csv.gz file, unzips it, and appends the data to the depth data file.
//
// The resulting CSV content format:
//
//	#,<Pair1>,<Pair2>...
//	<Pair1>,<BidPrice>,<BidSize>,<AskPrice>,<AskSize>,<BidPrice>,<BidSize>,<AskPrice>,<AskSize>,...
//	<Pair2>,<BidPrice>,<BidSize>,<AskPrice>,<AskSize>,<BidPrice>,<BidSize>,<AskPrice>,<AskSize>,...
//	...
//
// Each group of 4 values represents a single 1 minute record.
// There are as many 4-value groups as there are minutes in the startDate-endDate time-range.
// Example:
//
//	#,BTC-BUSD
//	BTC-BUSD,1633824000,54968.99_1.52092,54969_0.00001,1633824001,54968.99_0.00477,54969_0.15224
type Loader interface {
	// Load loads the market depth data for the given pair and time range.
	// It creates the file if it doesn't exist, and appends the data to the file if it does.
	// It returns the full content of the file after the load.
	Load(pairs []Pair, startDate time.Time, endDate time.Time) map[Pair][]string
	// Tick can be used to iterate the data after it has been loaded.
	// With each call, it moves the pointer to the next minute in the loaded data time range.
	Tick()
	// GetDepth returns the current depth record for the given pair.
	// To proceed to the next minute, call Tick().
	GetDepth(pair Pair) Record
}

func NewCCDepthLoader(market Market) Loader {
	return &CCDepthLoader{
		market:  market,
		records: make(map[Pair][]string),
	}
}

type Market string

const (
	MarketBitfinex               Market = "bitfinex"
	MarketBitmex                 Market = "bitmex"
	MarketBinance                Market = "binance"
	MarketBinanceCoinFutures     Market = "binance-coin-futures"
	MarketBinanceUsdsFutures     Market = "binance-usds-futures"
	MarketBinanceUs              Market = "binance-us"
	MarketBitstamp               Market = "bitstamp"
	MarketCoinbase               Market = "coinbase"
	MarketDeribit                Market = "deribit"
	MarketFtx                    Market = "ftx"
	MarketFtxUs                  Market = "ftx-us"
	MarketGateio                 Market = "gateio"
	MarketGateioPerpetualFutures Market = "gateio-perpetual-futures"
	MarketGemini                 Market = "gemini"
	MarketHuobi                  Market = "huobi"
	MarketHuobiCoinSwap          Market = "huobi-coin-swap"
	MarketHuobiUsdtSwap          Market = "huobi-usdt-swap"
	MarketKucoin                 Market = "kucoin"
	MarketKraken                 Market = "kraken"
	MarketKrakenFutures          Market = "kraken-futures"
	MarketOkex                   Market = "okex"
)

type Pair string

// Pair is a string like BTC-BUSD.
func (p Pair) String() string {
	return string(p)
}

// Base returns the base currency of the pair.
func (p Pair) Base() string {
	if !p.Valid() {
		return ""
	}
	return strings.Split(p.String(), "-")[0]
}

// Quote returns the quote currency of the pair.
func (p Pair) Quote() string {
	if !p.Valid() {
		return ""
	}
	return strings.Split(p.String(), "-")[1]
}

// Valid checks if the pair is valid.
func (p Pair) Valid() bool {
	return strings.Contains(p.String(), "-")
}

type CCDepthLoader struct {
	market  Market
	records map[Pair][]string
	index   int
}

func (l *CCDepthLoader) Load(pairs []Pair, startDate time.Time, endDate time.Time) map[Pair][]string {
	path := "data/" + startDate.Format("2006-01-02") + "_" + endDate.Format("2006-01-02") + "_depth.csv"
	// historyLength is number of minutes between start and end date
	historyLength := int(endDate.Sub(startDate).Minutes())

	var pairsToLoad []Pair

	if len(pairs) == 0 {
		pairsToLoad = defaultPairs[0:]
	}

	fileExists := false

	if _, err := os.Stat(path); err == nil {
		// open read mode
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		fileExists = true
		testPairs := pairs[0:]
		fileHistoryLength := l.readDepthRecordsFromFile(file, testPairs)

		if fileHistoryLength != 0 && math.Abs(float64(fileHistoryLength)-float64(historyLength)) >= 1400 {
			panic("file history length does not match the range for more than 1 day")
		}

		pairsToLoad = testPairs[0:]
		if len(pairsToLoad) == 0 {
			pairsToLoad = l.readPairNamesFromHeader(file)
		}
		pairsToLoad = slices.Filter(pairsToLoad, func(s Pair) bool {
			return l.records[s] == nil
		})
		if len(pairsToLoad) > 0 {
			fmt.Println("Missing prices will be fetched and appended to the file")
		}
	}

	// make sure the directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		panic(err)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	if !fileExists {
		// Put pairs in the file header as a comment
		_, err = file.WriteString(fmt.Sprintf("#,%s\n", slices.Join(defaultPairs, ",")))
		if err != nil {
			panic(err)
		}
	}

	// load data for missing pairs
	slices.Each(pairsToLoad, func(pair Pair) {
		var days []time.Time
		for date := startDate; date.Before(endDate); date = date.AddDate(0, 0, 1) {
			days = append(days, date)
		}
		recordsForEachDay := slices.MapAsync(days, 30, func(date time.Time) []string {
			fmt.Println("Downloading depth for", pair, date)
			return l.downloadDay(pair, date)
		})
		var fullRecord = slices.Concat(recordsForEachDay...)
		if len(fullRecord) == 0 {
			return
		}
		l.records[pair] = fullRecord
		_, err = file.WriteString(fmt.Sprintf("%s,%s\n", pair, slices.Join(fullRecord, ",")))
		if err != nil {
			panic(err)
		}
	})

	if len(pairsToLoad) > 0 {
		fmt.Println("Depth data written to", path)
	}

	return l.records
}

func (l *CCDepthLoader) downloadDay(pair Pair, date time.Time) (S []string) {
	url := l.getURL(pair.String(), date)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(err)
	}
	defer gz.Close()

	// Parse CSV into structure and keep in memory
	reader := csv.NewReader(gz)
	reader.FieldsPerRecord = -1

	// date is for every second, but we need only each minute
	var prevRecord []string
	var prevRecordTime time.Time
	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if record[0] == "time_seconds" {
			continue
		}
		// Parse time seconds into time
		s, _ := strconv.ParseInt(record[0], 10, 64)
		timeSeconds := time.Unix(s, 0)

		// if the gap between two records is more than 1 second, we should reuse the previous record
		if !prevRecordTime.IsZero() && timeSeconds.Sub(prevRecordTime) > time.Second {
			// add previous record for each missing minute
			for prevRecordTime.Add(time.Minute).Before(timeSeconds) {
				prevRecordTime = prevRecordTime.Add(time.Minute)
				records = append(records, prevRecord)
			}
		}

		if timeSeconds.Second() == 0 {
			bidPriceAndSize := strings.Split(record[1], "_")
			askPriceAndSize := strings.Split(record[2], "_")
			record = []string{
				bidPriceAndSize[0],
				bidPriceAndSize[1],
				askPriceAndSize[0],
				askPriceAndSize[1],
			}

			records = append(records, record)
			prevRecord = record
			prevRecordTime = timeSeconds
		}
	}
	if len(records) > 0 {
		// join records into one line
		fullRec := slices.Concat(records...)
		numbersPerRecord := 4
		minutesInADay := 1440
		if len(fullRec) != numbersPerRecord*minutesInADay {
			panic("wrong number of records: " + strconv.Itoa(len(fullRec)))
		}
		return fullRec
	}
	return nil
}

func (l *CCDepthLoader) getURL(pair string, date time.Time) string {
	url := "https://api.cryptochassis.com/v1/market-depth/" +
		string(l.market) + "/" +
		pair +
		"?startTime=" + date.Format("2006-01-02")

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		// check if error is Timeout then repeat the request after 1 second
		if strings.Contains(string(body), "Too many requests, please try again later.") {
			return l.getURL(pair, date)
		}
		panic(err)
	}
	urls := result["urls"].([]interface{})
	if len(urls) > 0 {
		return urls[0].(map[string]interface{})["url"].(string)
	}
	panic(err)
}

func (l *CCDepthLoader) readPairNamesFromHeader(file *os.File) []Pair {
	firstLine := l.readFirstLine(file)
	pairNames := strings.Split(firstLine, ",")
	if pairNames[0] == "#" {
		return slices.Map(pairNames[1:], func(s string) Pair {
			return Pair(s)
		})
	}
	return nil
}

func (l *CCDepthLoader) readFirstLine(file *os.File) string {
	_, _ = file.Seek(0, 0)
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	return scanner.Text()
}

func (l *CCDepthLoader) readDepthRecordsFromFile(file *os.File, pairs []Pair) uint {
	historyLength := uint(0)

	_, _ = file.Seek(0, 0)
	csvParser := csv.NewReader(file)
	csvParser.FieldsPerRecord = 0
	csvParser.TrimLeadingSpace = true
	csvParser.Comment = '#'

	foundPairs := make(map[Pair]bool)

	for {
		record, err := csvParser.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		pair := Pair(record[0])
		if len(pairs) > 0 && !slices.Contains(pairs, pair) {
			continue
		}
		depths := record[1:]
		historyLength = uint(math.Max(float64(historyLength), float64(len(depths)/4)))
		if len(depths) > 0 && len(depths)/4 != int(historyLength) {
			panic("file is corrupted: history length is not consistent at pair " + string(pair))
		}

		l.records[pair] = depths

		if len(pairs) > 0 {
			foundPairs[pair] = true
			if len(foundPairs) == len(pairs) {
				break
			}
		}
	}
	return historyLength
}

type Record struct {
	pair     Pair
	BidPrice float64
	BidSize  float64
	AskPrice float64
	AskSize  float64
}

func (r Record) SpreadPercentage() float64 {
	return (r.AskPrice - r.BidPrice) / r.BidPrice
}

func (r Record) Imbalance() float64 {
	return (r.BidSize - r.AskSize) / (r.BidSize + r.AskSize)
}

func (l *CCDepthLoader) Tick() {
	l.index += 4
}

func (l *CCDepthLoader) GetDepth(pair Pair) Record {
	if l.index >= len(l.records[pair]) {
		panic("index out of range")
	}
	record := l.records[pair][l.index : l.index+4]
	return Record{
		pair:     pair,
		BidPrice: mustParseFloat(record[0]),
		BidSize:  mustParseFloat(record[1]),
		AskPrice: mustParseFloat(record[2]),
		AskSize:  mustParseFloat(record[3]),
	}
}

func mustParseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return f
}

// set known available pairs from Crypto Chassis
// todo: do not hardcode
var defaultPairs = []Pair{
	"ADA-BUSD",
	"BCH-BUSD",
	"BNB-BUSD",
	"BTC-BUSD",
	"DOGE-BUSD",
	"DOT-BUSD",
	"EOS-BUSD",
	"ETH-BUSD",
	"LTC-BUSD",
	"SOL-BUSD",
	"UNI-BUSD",
	"XRP-BUSD",
}
