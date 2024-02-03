package server

// Represent an array of Isaiah hosts ([name, hostname])
type HostsArray [][]string

func (hosts HostsArray) ToStrings() []string {
	arr := make([]string, 0)

	for _, v := range hosts {
		arr = append(arr, v[0])
	}

	return arr
}
