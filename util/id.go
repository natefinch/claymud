package util

// Id is a type that allows for unique identification of an object
type Id uint64

var Ids <-chan (Id) = idGenerator()

// idGenerator creates a channel that will continuously generate new unique Ids
//
// Thanks to Andrew Rolfe of WolfMUD fame for this snippet of code -
// http://www.wolfmud.org
func idGenerator() <-chan (Id) {
	// TODO: initialize at startup with highest Id created previously
	next := make(chan Id)
	id := Id(0)
	go func() {
		for {
			next <- id
			id++
		}
	}()
	return next
}
