package misc

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPointer(t *testing.T) {
	cases := []struct {
		Name  string
		Input any
	}{
		{
			Name:  "string pointer",
			Input: "foo",
		},
		{
			Name:  "int pointer",
			Input: 1,
		},
		{
			Name:  "float pointer",
			Input: 1.10001,
		},
		{
			Name:  "bool pointer",
			Input: true,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			v := c.Input
			p := Pointer(v)
			if v != *p {
				t.Fatalf("given %s, got: %s \n", v, *p)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	{ // filter integer array
		out := Filter[int]([]int{1, 2, 3}, func(f int) bool {
			return f < 3
		})
		require.Len(t, out, 2)
		require.Equal(t, 1, out[0])
		require.Equal(t, 2, out[1])
	}

	{ // filter string array
		out := Filter[string]([]string{"foo", "bar", "bar", "zar"}, func(f string) bool {
			// filter bar out!
			return f != "bar"
		})
		require.Len(t, out, 2)
		require.Equal(t, "foo", out[0])
		require.Equal(t, "zar", out[1])
	}
	{ // filter base on custom object properties
		type foo struct {
			Bar int
		}
		out := Filter[foo]([]foo{{Bar: 1}, {Bar: 2}, {Bar: 3}, {Bar: 4}}, func(f foo) bool {
			return f.Bar%2 == 0
		})
		require.Len(t, out, 2)
		require.Equal(t, foo{Bar: 2}, out[0])
		require.Equal(t, foo{Bar: 4}, out[1])
	}
}

func TestExtract(t *testing.T) {
	{ // extract
		type foo struct {
			A int
		}

		type bar struct {
			A string
		}

		fl := []foo{{A: 1}, {A: 2}, {A: 3}}

		bl := Extract[foo, bar](fl, func(i foo) bar {
			return bar{A: strconv.Itoa(i.A)}
		})

		require.Len(t, bl, len(fl))
		require.Equal(t, strconv.Itoa(fl[0].A), bl[0].A)
		require.Equal(t, strconv.Itoa(fl[1].A), bl[1].A)
		require.Equal(t, strconv.Itoa(fl[2].A), bl[2].A)
	}
}

func TestContain(t *testing.T) {
	{ // check if a string array contain target value
		v := Contain[string]([]string{"foo", "bar", "zar"}, "foo")
		require.True(t, v)

		v = Contain[string]([]string{"foo", "bar", "zar"}, "oof")
		require.False(t, v)
	}

	{ // check if an integer array contain target value
		v := Contain[int]([]int{1, 2, 3, 4}, 4)
		require.True(t, v)

		v = Contain[int]([]int{1, 2, 3, 4}, 5)
		require.False(t, v)
	}
}

func TestIndex(t *testing.T) {
	{ // find the index of item in integer array
		i := Index[int]([]int{1, 2, 3, 4, 5}, 4)
		require.Equal(t, 3, i)

		i = Index[int]([]int{1, 2, 3, 4, 5}, 10)
		require.Equal(t, -1, i)
	}
}
