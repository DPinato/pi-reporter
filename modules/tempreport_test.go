package modules

import "testing"

func Test_getPITemperature(t *testing.T) {

	t.Run("Read temperature", func(t *testing.T) {
		got, err := getPITemperature()

		// check for error
		if err != nil {
			t.Errorf("Got error, %v\n", err)
		}

		// check that we got a valid value, these are in milli-Celsius
		if got < -20.0 || got > 110.0 {
			t.Errorf("Retrieved invalid temperature value, %f\n", got)
		}
	})
}
