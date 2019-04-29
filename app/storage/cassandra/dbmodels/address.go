package dbmodels

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/thatique/kuade/app/model"
)

type Address struct {
	Address  string  `cql:"address"`
	Address2 string  `cql:"address2"`
	City     string  `cql:"city"`
	State    string  `cql:"state"`
	Zipcode  string  `cql:"zipcode"`
	Lat      float64 `cql:"lat"`
	Lon      float64 `cql:"lon"`
}

func FromDomainAddress(address model.Address) *Address {
	addr := &Address{
		Address:  address.Address,
		Address2: address.Address2,
		City:     address.City,
		State:    address.State,
		Zipcode:  address.Zipcode,
	}
	if address.Point != nil {
		addr.Lat = address.Point.GetLatitude()
		addr.Lon = address.Point.GetLongitude()
	}
	return addr
}

func (a *Address) ToDomain() model.Address {
	return model.Address{
		Address:  a.Address,
		Address2: a.Address2,
		City:     a.City,
		State:    a.State,
		Zipcode:  a.Zipcode,
		Point: &model.Point{
			Latitude:  a.Lat,
			Longitude: a.Lon,
		},
	}
}

func (a *Address) MarshalUDT(name string, info gocql.TypeInfo) ([]byte, error) {
	switch name {
	case "address":
		return gocql.Marshal(info, a.Address)
	case "address2":
		return gocql.Marshal(info, a.Address2)
	case "city":
		return gocql.Marshal(info, a.City)
	case "state":
		return gocql.Marshal(info, a.State)
	case "zipcode":
		return gocql.Marshal(info, a.Zipcode)
	case "lat":
		return gocql.Marshal(info, a.Lat)
	case "lon":
		return gocql.Marshal(info, a.Lon)
	default:
		return nil, fmt.Errorf("unknown column for position: %q", name)
	}
}

// UnmarshalUDT handles unmarshalling a Tag.
func (a *Address) UnmarshalUDT(name string, info gocql.TypeInfo, data []byte) error {
	switch name {
	case "address":
		return gocql.Unmarshal(info, data, &a.Address)
	case "address2":
		return gocql.Unmarshal(info, data, &a.Address2)
	case "city":
		return gocql.Unmarshal(info, data, &a.City)
	case "state":
		return gocql.Unmarshal(info, data, &a.State)
	case "zipcode":
		return gocql.Unmarshal(info, data, &a.Zipcode)
	case "lat":
		return gocql.Unmarshal(info, data, &a.Lat)
	case "lon":
		return gocql.Unmarshal(info, data, &a.Lon)
	default:
		return fmt.Errorf("unknown column for position: %q", name)
	}
}
