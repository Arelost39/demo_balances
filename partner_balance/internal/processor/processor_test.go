package processor

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestAverageArrow(t *testing.T) {
    data := []float64{1, 2, 3, 3, 4, 1, 2, 0, 4, 5, 6, 1}
    got := AverageArrow(data)
    want := []float64{1, 1, 1, 1, 1, 1}
    assert.Equal(t, want, got)
}
func TestAverageArrow2(t *testing.T) {
    data := []float64{}
    got := AverageArrow(data)
    want := []float64{30, 30 ,30}
    assert.Equal(t, want, got)
}

func TestAverageValue(t *testing.T) {
    data := []float64{25, 30, 35}
    got := AverageValue(data)
    want := 120.00
    assert.Equal(t, want, got)
}
func TestAverageValue2(t *testing.T) {
    data := []float64{-7.67, 8.96, 35.73, 62.84, 101.02, 133.2, 169.63, 190.69, 228.72, 286.39, 313.86}
    got := AverageValue(AverageArrow(data))
    want := 128.61
    assert.Equal(t, want, got)
}

func TestAverageArrow_Empty(t *testing.T) {
    assert.Empty(t, AverageArrow([]float64{}))
}

func TestAverageArrow_Single(t *testing.T) {
    assert.Empty(t, AverageArrow([]float64{42}))
}

func TestAverageArrow_NoPositive(t *testing.T) {

    assert.Empty(t, AverageArrow([]float64{5, 5, 5}))
    assert.Empty(t, AverageArrow([]float64{10, 8, 6, 4}))
}


func TestAverageArrow_PositiveDifferences(t *testing.T) {
    data := []float64{1, 3, 2, 5}
    got := AverageArrow(data)

    want := []float64{2, 3}
    assert.Equal(t, want, got)
}

