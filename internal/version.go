package internal

import "fmt"

const (
	Version = "v1.0.0"
	Name    = "cmiyc"
)

func VersionedName() string {
	return fmt.Sprintf("%s %s", Version, Name)
}
