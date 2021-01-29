package main

import "strconv"

var checkInfo = []func(answer string) bool{
	func(answer string) bool {
		switch answer {
		case `male`, `female`:
			return true
		}
		return false
	},
	func(answer string) bool {
		_, err := strconv.Atoi(answer)
		return err == nil
	},
}

var checkRating = []func(answer string) bool{
	func(answer string) bool {
		switch answer {
		case `friendly`, `unfriendly`, `scam`, `return`:
			return true
		}
		return false
	},
}
