package model

import "testing"

func TestWithTimeParse(t *testing.T) {
	got := withTimeParse("tom:123321@tcp(127.0.0.1:3306)/launch")
	want := "tom:123321@tcp(127.0.0.1:3306)/launch?parseTime=true&loc=Local"
	if got != want {
		t.Fatalf("got %s", got)
	}
	got = withTimeParse("user:pass@tcp(127.0.0.1:3306)/launch?charset=utf8mb4")
	want = "user:pass@tcp(127.0.0.1:3306)/launch?charset=utf8mb4&parseTime=true&loc=Local"
	if got != want {
		t.Fatalf("got %s", got)
	}
	kept := "user:pass@tcp(127.0.0.1:3306)/launch?parseTime=True&loc=Local"
	if withTimeParse(kept) != kept {
		t.Fatalf("changed existing params: %s", withTimeParse(kept))
	}
}
