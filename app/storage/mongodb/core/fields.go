package core

// Point is geo json type point
type Point struct {
	Type        string     `bson:"type"`
	Coordinates [2]float64 `bson:"coordinates"`
}
