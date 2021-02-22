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
		age, err := strconv.Atoi(answer)
		return err == nil && age > 15 && age <= 35
	},
}

var checkRating = []func(answer string) bool{
	func(answer string) bool {
		switch answer {
		case `friendly`, `unfriendly`, `scam`:
			return true
		}
		return false
	},
}

func checkBanned(user User) bool {
	return user.Scam >= 2 || user.Unfriendly >= 5
}
