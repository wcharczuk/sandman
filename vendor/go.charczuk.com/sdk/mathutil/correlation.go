package mathutil

import (
	"math/big"
)

// Correlation computes the r^2 correlation from a given set of actual
// values to a given set of prediction values.
//
// An r^2 of +1.0 means perfectly correlated, and 0 means completely uncorrelated.
//
// Because the r^2 is "squared" you will detect inverse and positive correlation with the same output value.
func Correlation[T Operatable](actual, prediction []T) float64 {
	if len(actual) == 0 {
		return 0
	}
	if len(actual) != len(prediction) {
		return 0
	}

	actualMeanAccum := zeroBig()
	predictionMeanAccum := zeroBig()
	for x := 0; x < len(actual); x++ {
		actualMeanAccum = addBig(actualMeanAccum, valBig(actual[x]))
		predictionMeanAccum = addBig(predictionMeanAccum, valBig(prediction[x]))
	}
	actualMean := divBig(actualMeanAccum, valBig(len(actual)))
	predictionMean := divBig(predictionMeanAccum, valBig(len(prediction)))

	numerator := zeroBig()
	denominatorActual := zeroBig()
	denominatorPrediction := zeroBig()

	for x := 0; x < len(actual); x++ {
		actualValue := valBig(actual[x])
		predictionValue := valBig(prediction[x])

		actualVariance := subBig(actualValue, actualMean)
		predictionVariance := subBig(predictionValue, predictionMean)

		numerator = addBig(numerator, mulBig(actualVariance, predictionVariance))

		denominatorActual = addBig(denominatorActual, mulBig(actualVariance, actualVariance))
		denominatorPrediction = addBig(denominatorPrediction, mulBig(predictionVariance, predictionVariance))
	}

	r := divBig(numerator, sqrtBig(mulBig(denominatorActual, denominatorPrediction)))
	r2, _ := mulBig(r, r).Float64()
	return r2
}

func zeroBig() *big.Float {
	output := big.NewFloat(0)
	return output
}

func valBig[T Operatable](value T) *big.Float {
	output := big.NewFloat(float64(value))
	return output
}

func addBig(a, b *big.Float) *big.Float {
	output := zeroBig()
	return output.Add(a, b)
}

func subBig(a, b *big.Float) *big.Float {
	output := zeroBig()
	return output.Sub(a, b)
}

func mulBig(a, b *big.Float) *big.Float {
	output := zeroBig()
	return output.Mul(a, b)
}

func divBig(a, b *big.Float) *big.Float {
	output := zeroBig()
	return output.Quo(a, b)
}

func sqrtBig(a *big.Float) *big.Float {
	output := zeroBig()
	return output.Sqrt(a)
}
