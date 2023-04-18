package stat

import (
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCountAggregator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStater := NewMockStat(ctrl)
	tests := []struct {
		name            string
		ExecuteStatting func(stater CountAggregator, mockStater *MockStat)
	}{
		{
			name: "simple aggregation",
			ExecuteStatting: func(stater CountAggregator, mockStater *MockStat) {
				mockStater.EXPECT().Count("stat1", float64(15), []string{"tag1"})
				go stater.Count("stat1", 9, "tag1")
				go stater.Count("stat1", 6, "tag1")
				time.Sleep(200 * time.Millisecond)
			},
		},
		{
			name: "tag ordering aggregation",
			ExecuteStatting: func(stater CountAggregator, mockStater *MockStat) {
				mockStater.EXPECT().Count("stat1", float64(10), []string{"aTag", "bTab"})
				go func() {
					stater.Count("stat1", 9, "aTag", "bTab")
					stater.Count("stat1", 1, "bTab", "aTag")
				}()
				time.Sleep(200 * time.Millisecond)
			},
		},
		{
			name: "complex stat aggregation",
			ExecuteStatting: func(stater CountAggregator, mockStater *MockStat) {
				mockStater.EXPECT().Count("stat1", float64(15), []string{"tag1"})
				mockStater.EXPECT().Count("stat2", float64(4), []string{"tag2"})
				go func() {
					stater.Count("stat1", 9, "tag1")
					stater.Count("stat2", 3, "tag2")
					stater.Count("stat1", 6, "tag1")
					stater.Count("stat2", 1, "tag2")
				}()
				time.Sleep(200 * time.Millisecond)
			},
		},
		{
			name: "multiple stat aggregation",
			ExecuteStatting: func(stater CountAggregator, mockStater *MockStat) {
				mockStater.EXPECT().Count("stat1", float64(15), []string{"tag1"})
				mockStater.EXPECT().Count("stat2", float64(4), []string{"tag2"})
				go func() {
					stater.Count("stat1", 9, "tag1")
					stater.Count("stat2", 3, "tag2")
					stater.Count("stat1", 6, "tag1")
					stater.Count("stat2", 1, "tag2")
				}()
				time.Sleep(200 * time.Millisecond)
				mockStater.EXPECT().Count("stat1", float64(2), []string{"tag1"})
				mockStater.EXPECT().Count("stat2", float64(2), []string{"tag2"})
				go func() {
					stater.Count("stat1", 1, "tag1")
					stater.Count("stat2", 1, "tag2")
					stater.Count("stat1", 1, "tag1")
					stater.Count("stat2", 1, "tag2")
				}()
				time.Sleep(200 * time.Millisecond)
			},
		},
	}
	for _, test := range tests {

		t.Run(test.name, func(tt *testing.T) {
			countAggregator := CountAggregator{
				Stater:        mockStater,
				Bucket:        make(map[StatTagKey]float64),
				lock:          &sync.Mutex{},
				flushJob:      make(chan int, 1),
				FlushInterval: 100 * time.Millisecond,
			}
			test.ExecuteStatting(countAggregator, mockStater)
		})

	}
}
func TestGauge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStater := NewMockStat(ctrl)
	stater := CountAggregator{
		Stater: mockStater,
	}
	mockStater.EXPECT().Gauge("stat", float64(1), []string{"tag"})
	stater.Gauge("stat", 1, "tag")
}

func TestHistogram(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStater := NewMockStat(ctrl)
	stater := CountAggregator{
		Stater: mockStater,
	}
	mockStater.EXPECT().Histogram("stat", float64(1), []string{"tag"})
	stater.Histogram("stat", 1, "tag")
}
func TestTiming(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStater := NewMockStat(ctrl)
	stater := CountAggregator{
		Stater: mockStater,
	}
	mockStater.EXPECT().Timing("stat", 1*time.Millisecond, []string{"tag"})
	stater.Timing("stat", 1*time.Millisecond, "tag")
}

func TestAddTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStater := NewMockStat(ctrl)
	stater := CountAggregator{
		Stater: mockStater,
	}
	mockStater.EXPECT().AddTags([]string{"tag1", "tag2"})
	stater.AddTags("tag1", "tag2")
}

func TestGetTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStater := NewMockStat(ctrl)
	stater := CountAggregator{
		Stater: mockStater,
	}
	mockStater.EXPECT().GetTags().Return([]string{"tag1", "tag2"})
	tags := stater.GetTags()
	assert.Equal(t, []string{"tag1", "tag2"}, tags)
}
