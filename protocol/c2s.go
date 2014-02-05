package protocol

import (
	"strings"
	"strconv"
)

type protocolError struct {
	msg string
}

func (e protocolError) Error() string {
	return e.msg
}

type Version struct {
	Major uint
	Minor uint
	Patch uint
}

type Connect struct {
	Version
	Swname string
	Zonename string
	Hostname string
	Password string
}

type Plogin struct {
	Pid uint // Pid given by game server
	Flag bool // Register new name?
	Name string
	Pw string
	Ip string
	Macid uint // ?
	Contid string // 128 characters if using continuum, else empty
	
}

func parseVersion(versionField []string) (Version, error) {
	var version = Version{}

	major, err := strconv.Atoi(versionField[0])
	if err != nil {
		return version, protocolError{"parseVersion: Major is not an int"}
	}

	minor, err := strconv.Atoi(versionField[1])
	if err != nil {
		return version, protocolError{"parseVersion: Minor is not an int"}
	}

	patch, err := strconv.Atoi(versionField[2])
	if err != nil {
		return version, protocolError{"parseVersion: Patch is not an int"}
	}

	version.Major = uint(major)
	version.Minor = uint(minor)
	version.Patch = uint(patch)

	return version, nil
}

func ParseConnect(fields []string) (Connect, error) {
	msg := Connect{}

	if len(fields) != 6 {
		return msg, protocolError{"ParseConnect: Not enough fields"}
	}
	
	versionField := strings.Split(fields[1], ".")
	if len(versionField) != 3 {
		return msg, protocolError{"ParseConnect: Version does not contain 3 values"}
	}

	var err error
	msg.Version, err = parseVersion(versionField)
	if err != nil {
		return msg, err
	}

	msg.Swname = fields[2]
	msg.Zonename = fields[3]
	msg.Hostname = fields[4]
	msg.Password = fields[5]

	return msg, nil
}

func ParsePlogin(fields []string) (Plogin, error) {
	msg := Plogin{}

	if len(fields) != 8 {
		return msg, protocolError{"ParsePlogin: Not enough fields"}
	}
	
	pid, err := strconv.Atoi(fields[1])
	if err != nil {
		return msg, protocolError{"ParsePlogin: Invalid pid"}
	}
	msg.Pid = uint(pid)

	flag := false
	if fields[2] == "1" {
		flag = true
	}
	msg.Flag = flag

	msg.Name = fields[3]
	msg.Pw = fields[4]
	msg.Ip = fields[5]

	macid, err := strconv.Atoi(fields[6])
	if err != nil {
		return msg, protocolError{"ParsePlogin: Invalid macid"}
	}
	msg.Macid = uint(macid)

	msg.Contid = fields[7]

	return msg, nil
}
