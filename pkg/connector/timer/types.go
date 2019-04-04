package timer

const RequestTypeTicker = "ticker"

type RequestTicker struct {
	TickerID string `json:"ticker_id"`
	Now      int64  `json:"now"`
}

const RequestTypeTimer = "timer"

type RequestTimer struct {
	TimerID string                 `json:"timer_id"`
	Now     int64                  `json:"now"`
	Data    map[string]interface{} `json:"json"`
}
