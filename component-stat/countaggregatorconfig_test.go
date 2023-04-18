package stat

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	countAggregatorConfig := CountAggregatorConfig{}
	assert.Equal(t, "statcountaggregator", countAggregatorConfig.Name())
}

func TestComponentDefaultConfig(t *testing.T) {
	component := &CountAggregatorComponent{}
	config := component.Settings()
	assert.Equal(t, config.FlushInterval, 10*time.Second)
}

func TestNewWithValues(t *testing.T) {
	statComponent := &DatadogComponent{}
	component := &CountAggregatorComponent{StatComponent: statComponent}
	config := component.Settings()
	_, err := component.New(context.Background(), config)
	assert.Nil(t, err)

}
