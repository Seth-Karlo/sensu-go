package transformers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/types"
)

// OpenTSDBList contains a list of OpenTSDB metrics
type OpenTSDBList []OpenTSDB

// OpenTSDB contains values of an OpenTSDB metric
type OpenTSDB struct {
	Name      string
	Value     float64
	TagSet    []*types.MetricTag
	Timestamp int64
}

// Transform transforms metrics in OpenTSDB format to Sensu Metric Format
func (o OpenTSDBList) Transform() []*types.MetricPoint {
	var points []*types.MetricPoint
	for _, metric := range o {
		mp := &types.MetricPoint{
			Name:      metric.Name,
			Value:     metric.Value,
			Timestamp: metric.Timestamp,
			Tags:      metric.TagSet,
		}
		points = append(points, mp)
	}
	return points
}

// ParseOpenTSDB parses OpenTSDB metrics into a list of OpenTSDB structs
func ParseOpenTSDB(output string) (OpenTSDBList, error) {
	openTSDBList := OpenTSDBList{}

	// Split each line of the output into its own metric
	output = strings.TrimSpace(output)
	metrics := strings.Split(output, "\n")

	for _, metric := range metrics {
		parts := strings.Split(metric, " ")

		// Ensure we have all the required components. A single metric requires a
		// name, timestamp, value and at least one tag.
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid opentsdb metric, at least 4 arguments are required: %s", metric)
		}

		name := parts[0]

		// Convert the timestamp to a unix timestamp with second resolution
		timestamp, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid opentsdb metric timestamp, must be an integer: %s", parts[1])
		}
		if len(parts[1]) == 13 {
			timestamp = timestamp / 1000
		}

		// Parse the value as a float64
		value, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid opentsdb metric value, must be an integer or a floating point value: %s", parts[1])
		}

		// Create a OpenTSDB metric with what we have so far
		o := OpenTSDB{
			Name:      name,
			TagSet:    []*types.MetricTag{},
			Timestamp: timestamp,
			Value:     value,
		}

		// Extract the tag(s)
		for i := 3; i < len(parts); i++ {
			t := strings.Split(parts[i], "=")

			if len(t) != 2 {
				return nil, fmt.Errorf("invalid opentsdb metric tag: %s", parts[i])
			}

			tag := &types.MetricTag{
				Name:  t[0],
				Value: t[1],
			}

			// Add this tag to our metric
			o.TagSet = append(o.TagSet, tag)
		}

		// Add this metric to our list
		openTSDBList = append(openTSDBList, o)
	}

	return openTSDBList, nil
}
