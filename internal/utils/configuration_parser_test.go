package utils

import (
	"fmt"
	"testing"

	"github.com/CESARBR/knot-mqtt/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestGivenValidApplicationFilepathReturnConfiguration(t *testing.T) {
	expectedPertinentTags := make(map[int]string)
	expectedPertinentTags[1] = "GR-11-TIT-0410TE001-01"
	expectedPertinentTags[2] = "GR-11-TIT-0410TE001-02"
	expectedPertinentTags[3] = "GR-11-TIT-0410TE001-03"
	expectedPertinentTags[4] = "GR-11-TIT-0410TE001-04"
	expectedPertinentTags[5] = "GR-11-TIT-0410TE001-05"
	expectedConfiguration := entities.Application{
		IntervalBetweenRequestInSeconds: 30,
		PertinentTags:                   expectedPertinentTags,
	}

	var applicationConfig entities.Application

	applicationConfiguration, err := ConfigurationParser("application_configuration_test.yaml", applicationConfig)
	fmt.Println(applicationConfiguration)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfiguration.IntervalBetweenRequestInSeconds, applicationConfiguration.IntervalBetweenRequestInSeconds)
	assert.Equal(t, expectedConfiguration.PertinentTags[1], applicationConfiguration.PertinentTags[1])
	assert.Equal(t, expectedConfiguration.PertinentTags[2], applicationConfiguration.PertinentTags[2])
}

func TestGivenInvalidFilepathReturnError(t *testing.T) {
	var applicationConfig entities.Application

	_, err := ConfigurationParser("invalid_filepath.yaml", applicationConfig)
	assert.NotNil(t, err)
}
