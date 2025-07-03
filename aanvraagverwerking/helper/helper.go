package helper

import (
	"math"
	"math/rand"
	"time"
)

// GenerateRandomBudget genereert een random budget tussen min en max (in hele euro's)
func RandomFloat64Between(min, max float64) float64 {
    if min >= max {
        return min
    }

    // Zorg ervoor dat de random generator een unieke seed heeft
    // om willekeurige resultaten te krijgen bij elke aanroep
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    raw := min + r.Float64()*(max-min)

    // Afronden op 2 decimalen
    return math.Round(raw*100) / 100
}