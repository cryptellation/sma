package sma

import (
	"time"

	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/cryptellation/candlesticks/pkg/period"
)

// PointParameters is the params for the Point function.
type PointParameters struct {
	Candlesticks *candlestick.List
	PriceType    candlestick.PriceType
}

// Point is a point of the SMA.
type Point struct {
	Exchange  string
	Pair      string
	Period    period.Symbol
	PeriodNb  int
	PriceType candlestick.PriceType
	Time      time.Time
	Price     float64
}

// NewPoint creates a new point from the given parameters.
func NewPoint(params PointParameters) Point {
	var total float64

	// Generate point
	p := Point{
		Exchange:  params.Candlesticks.Metadata.Exchange,
		Pair:      params.Candlesticks.Metadata.Pair,
		Period:    params.Candlesticks.Metadata.Period,
		PeriodNb:  params.Candlesticks.Data.Len(),
		PriceType: params.PriceType,
	}

	// Get count of candlesticks
	count := params.Candlesticks.Data.Len()

	// Get total from the timeserie
	_ = params.Candlesticks.Loop(func(cs candlestick.Candlestick) (bool, error) {
		price := cs.Price(params.PriceType)

		// Reduce the count if the price is 0
		if price == 0 {
			count--
			return false, nil
		}

		total += price

		return false, nil
	})

	// Get point time
	last, ok := params.Candlesticks.Last()
	if ok {
		p.Time = last.Time
		if count > 0 {
			p.Price = total / float64(count)
		} else {
			p.Price = 0
		}
	} else {
		p.Price = 0
	}

	return p
}
