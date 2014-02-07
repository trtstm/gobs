package zone

type Zone struct {
	name string
}

func NewZone(name string) *Zone {
	zone := Zone{name: name}

	return &zone
}
