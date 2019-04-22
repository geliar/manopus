package timer

const requestTypeTicker = "ticker"

type requestTicker struct {
	TickerID string `json:"ticker_id"`
	Now      int64  `json:"now"`
}

const requestTypeTimer = "timer"

type requestTimer struct {
	TimerID string                 `json:"timer_id"`
	Now     int64                  `json:"now"`
	Data    map[string]interface{} `json:"json"`
}
