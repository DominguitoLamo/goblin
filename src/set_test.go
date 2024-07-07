package main

import (
    "testing"
)

func TestSet(t *testing.T) {
    s := createSet()
    s.add("a")
    s.add("b")
    s.add("c")
    if s.size() != 3 {
        t.Errorf("Expected size 3, got %d", s.size())
    }

    arr := []string{"d", "e", "f"}
    s.addArr(arr)

    if s.size() != 6 {
        t.Errorf("Expected size 6, got %d", s.size())
    }

    if !s.contains("a") {
        t.Errorf("Expected to contain 'a'")
    }

    if !s.contains("f") {
        t.Errorf("Expected to contain 'f'")
    }

    s.remove("a")

    if s.contains("a") {
        t.Errorf("Expected not to contain 'a'")
    }

    s1 := createSet()
    s1.addArr(arr)

    s2 := createSet()
    arr2 := []string{"d", "e", "f", "g"}
    s2.addArr(arr2)

    if s1.equal(s2) {
        t.Errorf("Expected sets to be not equal")
    }
}