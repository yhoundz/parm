/*
Copyright © 2025 Alexander Wang
*/

package parmver

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

type Channel int

const (
	Unknown Channel = iota - 1
	Dev
	Stable
)

type Version struct {
	major   uint
	minor   uint
	patch   uint
	channel Channel
}

var (
	StringVersion string
	Owner         string = "yhoundz"
	Repo          string = "parm"
	AppVersion    Version
)

func init() {
	if StringVersion == "" {
		return
	}

	ver, err := semver.NewVersion(StringVersion)
	if err != nil {
		return
	}

	chstr := ver.Prerelease()
	var channel Channel
	switch chstr {
	case "stable", "":
		channel = Stable
	case "dev":
		channel = Dev
	default:
		channel = Unknown
	}

	AppVersion = Version{
		major:   uint(ver.Major()),
		minor:   uint(ver.Minor()),
		patch:   uint(ver.Patch()),
		channel: channel,
	}
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
		return fmt.Sprintf("v%d.%d.%d-%s", v.major, v.minor, v.patch, v.channel.String())
	case Stable:
		return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
	default:
		return fmt.Sprintf("v%d.%d.%d-%s", v.major, v.minor, v.patch, Unknown.String())
	}
}
