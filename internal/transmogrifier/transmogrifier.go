package transmogrifier

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/martinlindhe/unit"
	log "github.com/sirupsen/logrus"
)

type dataTransmogrifier func(data interface{}) interface{}

type wundergroundDataType struct {
	name         string
	format       string
	transmogrify dataTransmogrifier
}

func Start(ctx context.Context, wg *sync.WaitGroup, dataCh <-chan map[string]interface{}) <-chan map[string]string {
	out := make(chan map[string]string)

	wg.Add(1)

	go func() {
		defer close(out)
		defer wg.Done()

		for {
			select {
			case data, ok := <-dataCh:
				if !ok {
					return
				}

				log.Debug(data)

				outData, err := transmogrify(data)
				if err != nil {
					log.Error(err)
					continue
				}

				log.Debug(outData)

				out <- outData

			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func transmogrify(data map[string]interface{}) (map[string]string, error) {
	out := make(map[string]string)

	for k, v := range data {
		newK, newV, err := formatField(k, v)
		if err != nil {
			log.Error(err)
			continue
		}

		out[newK] = newV
	}

	return out, nil
}

func formatField(key string, value interface{}) (outKey string, outValue string, err error) {
	// https://support.weather.com/s/article/PWS-Upload-Protocol?language=en_US
	innerMap := map[string]wundergroundDataType{
		"dewpoint": {"dewptf", "%.1f", ctof},
		"hum":      {"humidity", "%.f", nil},
		"temp":     {"tempf", "%.1f", ctof},
		"solar":    {"solarradiation", "%.1f", nil},
		"uv":       {"UV", "%.1f", nil},

		"wdir":      {"winddir", "%.f", nil},
		"wdir_gust": {"windgustdir", "%.f", nil},
		"wdir_2m":   {"winddir_avg2m", "%.f", nil},

		"wspd":      {"windspeedmph", "%.1f", mpstomph},
		"wspd_2m":   {"windspdmph_avg2m", "%.1f", mpstomph},
		"wspd_gust": {"windgustmph", "%.1f", mpstomph},

		"rain_1hr":  {"rainin", "%.1f", mmtoin},
		"rain_24hr": {"dailyrainin", "%.1f", mmtoin},
	}

	err = nil

	key = strings.ToLower(key)

	dataType, ok := innerMap[key]

	if !ok {
		err = fmt.Errorf("unknown field %s", key)
		return
	}

	outKey = dataType.name

	if dataType.transmogrify != nil {
		value = dataType.transmogrify(value)
	}

	outValue = fmt.Sprintf(dataType.format, value)
	if strings.HasPrefix(outValue, "%!") {
		err = fmt.Errorf("invalid conversion for,\n\tfield: \"%s\"\n\tvalue \"%v\"\n\tformat \"%s\"", key, dataType.format, value)
		return
	}

	return
}

// Temperature [C] to [F]
func ctof(val interface{}) interface{} {
	return unit.FromCelsius(val.(float64)).Fahrenheit()
}

// Speed [m/s] to [mph]
func mpstomph(val interface{}) interface{} {
	mps := unit.Speed(val.(float64)) * unit.MetersPerSecond

	return mps.MilesPerHour()
}

// Sunlight [lux] to solar radiation [W/m2]
func stosr(val interface{}) interface{} {
	//https://help.ambientweather.net/help/why-is-the-lux-to-w-m-2-conversion-factor-126-7/
	return val.(float64) / 126.7
}

// Distance [mm] to [in]
func mmtoin(val interface{}) interface{} {
	mm := unit.Length(val.(float64)) * unit.Millimeter

	return mm.Inches()
}
