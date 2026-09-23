package arabic

import "testing"

func TestShape(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "isolated", in: "ب", want: "ﺏ"},
		{name: "initial final", in: "بب", want: "ﺑﺐ"},
		{name: "medial", in: "ببب", want: "ﺑﺒﺐ"},
		{name: "right joining breaks", in: "بدب", want: "ﺑﺪﺏ"},
		{name: "marks are transparent", in: "بَب", want: "ﺑَﺐ"},
		{name: "lam alef ligature", in: "لا", want: "ﻻ"},
		{name: "lam alef after join", in: "بلا", want: "ﺑﻼ"},
		{name: "lam mark alef", in: "لَا", want: "ﻻَ"},
		{name: "space separates", in: "ب ب", want: "ﺏ ﺏ"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Shape(c.in)
			if got != c.want {
				t.Fatalf("Shape(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestVisual(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "reverses words", in: "ب د", want: "ﺩ ﺏ"},
		{name: "keeps marks after base", in: "بَ د", want: "ﺩ ﺏَ"},
		{name: "digits keep order", in: "ب ١٢", want: "١٢ ﺏ"},
		{name: "latin keeps order", in: "ب abc", want: "abc ﺏ"},
		{name: "brackets mirror", in: "(ب)", want: "(ﺏ)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Visual(c.in)
			if got != c.want {
				t.Fatalf("Visual(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
