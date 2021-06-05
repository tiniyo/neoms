package helper

import "regexp"

/*
 We will keep all phone number related utility here
 */
func NumberSanity(number string) string {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return number
	}
	return reg.ReplaceAllString(number, "")
}

func RemovePlus(number string) string {
	reg, err := regexp.Compile("[^0-9]+")
	if err != nil {
		return number
	}
	return reg.ReplaceAllString(number, "")
}
