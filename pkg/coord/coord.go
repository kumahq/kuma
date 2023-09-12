package coord

import (
	"math"
	"strconv"

	"github.com/pkg/errors"
)

type Coord struct {
	Lat  float64
	Long float64
}

func NewCoord(in []string) (Coord, error) {
	c := Coord{}

	if len(in) != 2 {
		return c, errors.New("must have 2 elements")
	}

	lat, err := strconv.ParseFloat(in[0], 64)
	if err != nil {
		return c, errors.New("failed to convert latitude to float")
	}
	long, err := strconv.ParseFloat(in[1], 64)
	if err != nil {
		return c, errors.New("failed to convert longitude to float")
	}

	c.Lat = lat * math.Pi / 180
	c.Long = long * math.Pi / 180
	return c, nil
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//
//	a given longitude and latitude relatively accurately (using a spherical
//	approximation of the Earth) through the Haversin Distance Formula for
//	great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func (c Coord) Distance(c2 Coord) float64 {
	var r float64 = 6378100 // Earth radius in METERS
	// calculate
	h := hsin(c2.Lat-c.Lat) + math.Cos(c.Lat)*math.Cos(c2.Lat)*hsin(c2.Long-c.Long)

	return 2 * r * math.Asin(math.Sqrt(h))
}
