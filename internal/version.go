package internal

import "fmt"

const (
	Version = "v0.1.0"
	Name    = "cmiyc"
)

func VersionedName() string {
	return fmt.Sprintf("%s %s", Version, Name)
}
