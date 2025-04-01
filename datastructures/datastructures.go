package datastructures

import (
	"fmt"
	"slices"
)

func IsValueInSlice(value interface{}, slice []interface{}) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func StringToInterfaceSlice(arr []string) []interface{} {
	result := make([]interface{}, len(arr))
	for i, v := range arr {
		result[i] = v
	}
	return result
}

func RemoveSliceValue(value interface{}, slice []interface{}) []interface{} {
	for i, item := range slice {
		if item == value {
			// slice = append(slice[:i], slice[i+1:]...)
			slice = slices.Delete(slice, i, i+1)
			break
		}
	}
	return slice
}

func GetIntSliceSum(values []int) int {
	rs := 0

	for _, value := range values {
		rs += value
	}

	return rs
}

func RemoveSliceValueTwoD(value int, slice [][]interface{}) [][]interface{} {
	for i := range slice {
		if i == value {
			// slice = append(slice[:i], slice[i+1:]...)
			slice = slices.Delete(slice, i, i+1)
			break
		}
	}
	return slice
}

func TwoDStringToInterfaceSlice(arr [][]string) [][]interface{} {
	var interfaceSlice [][]interface{}

	for _, innerSlice := range arr {
		interfaceSlice = append(interfaceSlice, StringToInterfaceSlice(innerSlice))
	}

	return interfaceSlice
}

func InterfaceToStringSlice(arr []interface{}) []string {
	var interfaceSlice []string

	for _, innerSlice := range arr {
		interfaceSlice = append(interfaceSlice, fmt.Sprintf("%v", innerSlice))
	}

	return interfaceSlice
}

func InterfaceToTwoDStringSlice(arr [][]interface{}) [][]string {
	var stringSlice [][]string

	for _, innerSlice := range arr {
		stringSlice = append(stringSlice, InterfaceToStringSlice(innerSlice))
	}

	return stringSlice
}
