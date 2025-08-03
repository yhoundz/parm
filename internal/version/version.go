/*
Copyright © 2025 A.W. <EMAIL ADDRESS>
*/

package version

import "fmt"

type ReleaseChannel int // channel
type ReleaseStage int   // stage

const (
	Dev ReleaseChannel = iota
	Nightly
	Stable
)

const (
	Alpha ReleaseStage = iota
	Beta
	Rc
	Release
)

type version struct {
	major   int
	minor   int
	patch   int
	channel ReleaseChannel
	stage   ReleaseStage
}

func (c ReleaseChannel) String() string {
	switch c {
	case Dev:
		return "dev"
	case Nightly:
		return "nightly"
	case Stable:
		return "stable"
	default:
		return "ch?"
	}
}

func (s ReleaseStage) String() string {
	switch s {
	case Alpha:
		return "alpha"
	case Beta:
		return "beta"
	case Rc:
		return "rc"
	case Release:
		return "release"
	default:
		return "st?"
	}
}

func (v version) String() string {
	return fmt.Sprintf("%d.%d.%d-%s/%s", v.major, v.minor, v.patch, v.channel.String(), v.stage)
}

var Version = version{
	major:   0,
	minor:   1,
	patch:   0,
	channel: Dev,
	stage:   Alpha,
}
