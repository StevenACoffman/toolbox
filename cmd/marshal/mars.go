package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// A Month specifies a month of the year (January = 1, ...).
type Month int

const (
	January Month = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

var months = [...]string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

// String returns the English name of the month ("January", "February", ...).
func (m Month) String() string { return months[m-1] }

func main() {
	month := December
	if month == December {
		fmt.Println("Found a December")
	}

	month = month + Month(2)
	// %!v(PANIC=runtime error: index out of range)
	//fmt.Println(month)

	month = January + Month(2)
	fmt.Println(month)

	month++
	fmt.Println(month)

	day := 34
	month = Month(day % 31)
	fmt.Println(month)

	val := int(month) + 4
	fmt.Println(val)

	month = Month(val) + 1
	fmt.Println(month)

	b, err := month.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println("MARSHALLED",string(b))

	b = []byte(`{"month":"January"}`)

	m := MyMonth{}
	json.Unmarshal(b, &m)
	fmt.Println(m)

}

func (m Month) IsValid() error {
	switch m {
	case January, February, March, April, May, June, July, August, September, October, November, December:
		return nil
	}
	return errors.New("Inalid leave type")
}

func NewMonth(str String) (Month, error) {
	index := SliceIndex(len(months), func(i int) bool { return strings.EqualFold(months[i], value) })
	if index == -1 {
		return nil, errors.New("Inalid month type")
	}
	return Month(index+1)
}

type MyMonth struct {
	TheMonth Month `json:"month"`
}

func (m *Month) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`)
	*m = NewMonth(value)
	return nil
}

func (m *Month) MarshalJSON() ([]byte, error) {
	return []byte( months[*m-1]), nil
}

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
