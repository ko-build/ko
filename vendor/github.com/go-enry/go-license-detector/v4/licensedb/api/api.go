package api

// Match is a detection result of a license with a confidence (0.0 - 1.0)
// and a mapping of files to confidence.
type Match struct {
	Files      map[string]float32
	Confidence float32
	File       string
}
