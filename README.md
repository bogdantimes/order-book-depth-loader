# order-book-depth-loader
A tool that loads order book depth historical data for back-testing. Loads 1 minute frequency depth data provided by Crypto Chassis project.

## Installation

```go
go get github.com/bogdantimes/order-book-depth-loader@latest
```

## Docs

```
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
```
