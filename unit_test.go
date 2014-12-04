package main

import (
	"testing"
)

func TestCapitalCase(t *testing.T) {
	cases := [][]string{
		[]string{"cp_user_124_jiu", "CpUser124Jiu"},
		[]string{"Cp_u___test", "CpUTest"},
		[]string{"hello23World", "Hello23World"},
		[]string{"CP_test_USer", "CpTestUser"},
		[]string{"USER", "User"},
	}
	for _, cs := range cases {
		target := toCapitalCase(cs[0])
		if target != cs[1] {
			t.Errorf("src %s, expected %s, got %s", cs[0], cs[1], target)
		}
	}
}