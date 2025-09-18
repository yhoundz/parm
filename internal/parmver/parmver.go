/*
Copyright © 2025 A.W. <EMAIL ADDRESS>
*/

package parmver

import "fmt"

type ReleaseChannel int

const (
	Dev ReleaseChannel = iota
	Stable
)

type Version struct {
	major   int
	minor   int
	patch   int
	channel ReleaseChannel
}

func (c ReleaseChannel) String() string {
	switch c {
	case Stable:
		return "stable"
	default:
		return "ch?"
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

var AppVersion = Version{
	major:   0,
	minor:   0,
	patch:   0,
	channel: Dev,
}
