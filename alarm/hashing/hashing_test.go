package hashing

import "testing"

func TestHash(t *testing.T) {


	for i:= 0; i<25; i++ {

		hashedValue :=Hash("10baeb6b-199f-4300-8db7-1acbe744e3fc")

		if hashedValue != 1532899497 {
			t.FailNow()
		}

	}

}
