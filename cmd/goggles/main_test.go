package main

import "testing"

func TestProductName(t *testing.T) {
	if productName != "goggles" {
		t.Fatalf("product name = %q, want %q", productName, "goggles")
	}
}
