/*
Copyright © 2025 Alexander Wang
*/

package parmver

import "fmt"

type Channel int

const (
	Dev Channel = iota
	Stable
)

type Version struct {
	major   int
	minor   int
	patch   int
	channel Channel
}

func (c Channel) String() string {
	switch c {
	case Stable:
		return "stable"
	case Dev:
		return "dev"
	default:
		return "unknown?"
	}
}

func (v Version) String() string {
	switch v.channel {
	case Dev:
		return fmt.Sprintf("%d.%d.%d-%s", v.major, v.minor, v.patch, v.channel.String())
	default:
		return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	}
}

var AppVersion = Version{}
