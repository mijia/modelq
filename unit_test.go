package main

import (
	"testing"
	"github.com/mijia/modelq/gmq"
	"log"
)

func TestOptions(t *testing.T) {
	var o gmq.OptionInt
	o = gmq.SomeInt(5)
	if n, err := o.Get(); err != nil || n != 5 {
		t.Errorf("Option get is not working, got err or wrong value, %s, %d", err, n)
	}
	if !o.IsDefined() {
		t.Errorf("Option isDefined is not working, expectd defined")
	}

	o = gmq.NoneInt()
	if n, err := o.Get(); err == nil {
		t.Errorf("Option get is not working, should get error for NoneInt, %s, %d", err, n)
	}
	if o.IsDefined() {
		t.Errorf("Option isDefined is not working, expectd not defined")
	}
}

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

func TestGmqFilters(t *testing.T) {
	left := gmq.UnitFilter("id", "=", 1)
	log.Println(left.SqlString("User"), left.Params())
	
	right := gmq.UnitFilter("name", "LIKE", "hello%")
	log.Println(right.SqlString("User"), right.Params())

	and := left.And(right)
	log.Println(and.SqlString("User"), and.Params())

	in := gmq.InFilter("id", []interface{}{10, 20, 30})
	log.Println(in.SqlString("User"), in.Params())

	or := and.Or(in)
	log.Println(or.SqlString("User"), or.Params())
}

func init() {
	gmq.Debug = true
}

var _ = log.Println