package slimarray

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/openacid/low/size"
	"github.com/openacid/testutil"
	"github.com/stretchr/testify/require"
)

var testNums []uint32 = []uint32{
	0, 16, 32, 48, 64, 79, 95, 111, 126, 142, 158, 174, 190, 206, 222, 236,
	252, 268, 275, 278, 281, 283, 285, 289, 296, 301, 304, 307, 311, 313, 318,
	321, 325, 328, 335, 339, 344, 348, 353, 357, 360, 364, 369, 372, 377, 383,
	387, 393, 399, 404, 407, 410, 415, 418, 420, 422, 426, 430, 434, 439, 444,
	446, 448, 451, 456, 459, 462, 465, 470, 473, 479, 482, 488, 490, 494, 500,
	506, 509, 513, 519, 521, 528, 530, 534, 537, 540, 544, 546, 551, 556, 560,
	566, 568, 572, 574, 576, 580, 585, 588, 592, 594, 600, 603, 606, 608, 610,
	614, 620, 623, 628, 630, 632, 638, 644, 647, 653, 658, 660, 662, 665, 670,
	672, 676, 681, 683, 687, 689, 691, 693, 695, 697, 703, 706, 710, 715, 719,
	722, 726, 731, 735, 737, 741, 748, 750, 753, 757, 763, 766, 768, 775, 777,
	782, 785, 791, 795, 798, 800, 806, 811, 815, 818, 821, 824, 829, 832, 836,
	838, 842, 846, 850, 855, 860, 865, 870, 875, 878, 882, 886, 890, 895, 900,
	906, 910, 913, 916, 921, 925, 929, 932, 937, 940, 942, 944, 946, 952, 954,
	956, 958, 962, 966, 968, 971, 975, 979, 983, 987, 989, 994, 997, 1000,
	1003, 1008, 1014, 1017, 1024, 1028, 1032, 1034, 1036, 1040, 1044, 1048,
	1050, 1052, 1056, 1058, 1062, 1065, 1068, 1072, 1078, 1083, 1089, 1091,
	1094, 1097, 1101, 1104, 1106, 1110, 1115, 1117, 1119, 1121, 1126, 1129,
	1131, 1134, 1136, 1138, 1141, 1143, 1145, 1147, 1149, 1151, 1153, 1155,
	1157, 1159, 1161, 1164, 1166, 1168, 1170, 1172, 1174, 1176, 1178, 1180,
	1182, 1184, 1186, 1189, 1191, 1193, 1195, 1197, 1199, 1201, 1203, 1205,
	1208, 1210, 1212, 1214, 1217, 1219, 1221, 1223, 1225, 1227, 1229, 1231,
	1233, 1235, 1237, 1239, 1241, 1243, 1245, 1247, 1249, 1251, 1253, 1255,
	1257, 1259, 1261, 1263, 1265, 1268, 1270, 1272, 1274, 1276, 1278, 1280,
	1282, 1284, 1286, 1288, 1290, 1292, 1294, 1296, 1298, 1300, 1302, 1304,
	1306, 1308, 1310, 1312, 1314, 1316, 1318, 1320, 1322, 1324, 1326, 1328,
	1330, 1332, 1334, 1336, 1338, 1340, 1342, 1344, 1346, 1348, 1350, 1352}

func TestMarginWidth(t *testing.T) {

	ta := require.New(t)

	cases := []struct {
		input int64
		want  uint32
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 4},
		{15, 4},
		{16, 8},
		{255, 8},
		{256, 16},
		{65535, 16},
		{65536, 32},
		{0x7fffffff, 32},
		{-1, 64},
	}

	for i, c := range cases {
		got := marginWidth(c.input)
		ta.Equal(c.want, got,
			"%d-th: input: %#v; want: %#v; got: %#v",
			i+1, c.input, c.want, got)
	}
}

func TestSlimArray_New(t *testing.T) {
	ta := require.New(t)

	cases := [][]uint32{
		{},
		{0},
		{1},
		{1, 2},
		testNums[:10],
		testNums[:50],
		testNums[:200],
		testNums,
	}

	for _, nums := range cases {

		a := NewU32(nums)
		testGet(ta, a, nums)

		// Stat() should work
		_ = a.Stat()
		fmt.Println(a.Stat())
	}
}

func TestSlimArray_eltWidthSmall(t *testing.T) {

	ta := require.New(t)

	n := 500
	nums := make([]uint32, n)
	for i := 0; i < n; i++ {
		nums[i] = uint32(15 * i)
	}

	a := NewU32(nums)
	fmt.Println(a.Stat())
	ta.True(a.Stat()["bits/elt"] <= 5)

}

func TestSlimArray_default(t *testing.T) {

	ta := require.New(t)

	a := NewU32(testNums)
	ta.Equal(int32(len(testNums)), a.N)

	fmt.Println(a.Stat())
	st := a.Stat()

	fmt.Println(size.Stat(a, 6, 3))
	ta.Equal(int32(3), st["elt_width"])

}

func TestSlimArray_Slice(t *testing.T) {

	ta := require.New(t)

	a := NewU32(testNums)

	for i := 0; i < len(testNums); i += 3 {
		for j := i; j < len(testNums)+10; j += 5 {
			e := j
			if e > len(testNums) {
				e = len(testNums)
			}

			rst := make([]uint32, e-i)
			a.Slice(int32(i), int32(e), rst)
			ta.Equal(testNums[i:e], rst)

		}
	}

}

func TestSlimArray_big(t *testing.T) {

	ta := require.New(t)

	n := int32(1024 * 1024)
	step := int32(64)
	ns := testutil.RandU32Slice(0, n, step)

	a := NewU32(ns)
	st := a.Stat()
	fmt.Println(st)

	testGet(ta, a, ns)
}

func TestSlimArray_bigResidual_lowhigh(t *testing.T) {

	ta := require.New(t)

	big := uint32(1<<31 - 1)

	ns := []uint32{
		0, 0, 0, 0,
		0, 0, 0, 0,
		big, big, big, big,
		big, big, big, big,
	}

	a := NewU32(ns)
	st := a.Stat()
	fmt.Println(st)

	testGet(ta, a, ns)
}

func TestSlimArray_bigResidual_zipzag(t *testing.T) {

	ta := require.New(t)

	big := uint32(0xffffffff)

	ns := []uint32{
		0, 0, 0, 0,
		0, big, 0, 0,
		0, 0, 0, 0,
		0, big, 0, 0,
	}

	a := NewU32(ns)
	st := a.Stat()
	fmt.Println(st)

	testGet(ta, a, ns)
}

func TestSlimArray_bigResidual_rand(t *testing.T) {

	// unsorted rand large array

	ta := require.New(t)

	big := uint32(0xffffffff)

	ns := []uint32{}
	n := 1024 * 1024
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < n; i++ {
		s := uint32(rnd.Float64() * float64(big))
		ns = append(ns, s)
	}

	a := NewU32(ns)
	st := a.Stat()
	fmt.Println(st)

	testGet(ta, a, ns)
}

func TestSlimArray_largenum(t *testing.T) {

	ta := require.New(t)

	n := int32(1024 * 1024)
	step := int32(64)
	ns := testutil.RandU32Slice(1<<30, n, step)

	a := NewU32(ns)
	testGet(ta, a, ns)
}

func TestSlimArray_Get_panic(t *testing.T) {
	ta := require.New(t)

	a := NewU32(testNums)
	ta.Panics(func() {
		a.Get(int32(len(testNums) + 64))
	})
	ta.Panics(func() {
		a.Get(int32(-1))
	})
}

func TestSlimArray_Get2(t *testing.T) {

	ta := require.New(t)

	n := int32(1024 * 1024)
	step := int32(32)
	nums := testutil.RandU32Slice(1<<30, n, step)

	a := NewU32(nums)

	for i, n := range nums {
		if i < len(nums)-1 {

			r, rnext := a.Get2(int32(i))
			ta.Equal(n, r, "i=%d expect: %v; but: %v", i, n, r)
			ta.Equal(nums[i+1], rnext, "i=%d expect: %v; but: %v", i, nums[i+1], rnext)
		}
	}
}

func TestSlimArray_Stat(t *testing.T) {

	ta := require.New(t)

	a := NewU32(testNums)

	st := a.Stat()
	want := map[string]int32{
		"seg_cnt":   1,
		"elt_width": 3,
		"mem_elts":  160,
		"n":         354,
		"mem_total": st["mem_total"], // do not compare this
		"spans/seg": 4,
		"span_cnt":  5,
		"bits/elt":  11,
	}

	ta.Equal(want, st)
}

func TestSlimArray_marshalUnmarshal(t *testing.T) {
	ta := require.New(t)

	a := NewU32(testNums)

	bytes, err := proto.Marshal(a)
	ta.Nil(err, "want no error but: %+v", err)

	b := &SlimArray{}

	err = proto.Unmarshal(bytes, b)
	ta.Nil(err, "want no error but: %+v", err)

	testGet(ta, b, testNums)
}

func testGet(ta *require.Assertions, a *SlimArray, nums []uint32) {
	for i, n := range nums {
		r := a.Get(int32(i))
		ta.Equal(n, r, "i=%d expect: %v; but: %v", i, n, r)
	}
	ta.Equal(len(nums), a.Len())
}

func TestSpan_String(t *testing.T) {

	ta := require.New(t)

	sp := span{
		poly:          []float64{1, 2, 3},
		residualWidth: 1,
		mem:           3,
		s:             1,
		e:             2,
	}

	s := sp.String()
	ta.Equal("1-2(1): width: 1, mem: 3, poly: [1 2 3]", s)
}

var Output int

func BenchmarkSlimArray_Get(b *testing.B) {

	n := int32(1024 * 1024)
	mask := int(n - 1)
	step := int32(128)
	ns := testutil.RandU32Slice(0, n, step)

	s := uint32(0)

	a := NewU32(ns)
	// fmt.Println(a.Stat())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s += a.Get(int32(i & mask))
	}

	Output = int(s)
}

func BenchmarkSlimArray_Get2(b *testing.B) {

	n := int32(1024*1024) + 1
	mask := int(1024*1024 - 1)
	step := int32(128)
	ns := testutil.RandU32Slice(0, n, step)

	s := uint32(0)

	a := NewU32(ns)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x, _ := a.Get2(int32(i & mask))
		s += x
	}

	Output = int(s)
}

func BenchmarkSlimArray_Slice(b *testing.B) {

	n := int32(1024 * 1024)
	mask := int(n - 1)
	step := int32(128)
	ns := testutil.RandU32Slice(0, n, step)

	s := uint32(0)

	a := NewU32(ns)
	// fmt.Println(a.Stat())

	for _, batchSize := range []int{1, 10, 100, 1000, 10000} {

		rst := make([]uint32, batchSize)
		b.Run(
			fmt.Sprintf(
				"Slice() n=:%d", batchSize,
			),
			func(b *testing.B) {

				for i := 0; i < b.N/batchSize; i++ {
					a.Slice(int32(i&mask), int32(i&mask+batchSize), rst)
					s += rst[0]
				}

				Output = int(s)
			})
	}
}

func BenchmarkNewU32(b *testing.B) {

	n := int32(1024 * 10)
	step := int32(128)
	ns := testutil.RandU32Slice(0, n, step)

	s := uint32(0)

	b.ResetTimer()
	var a *SlimArray
	for i := 0; i < b.N/int(n)+1; i++ {
		a = NewU32(ns)
		s += a.Get(int32(0))
	}

	// fmt.Println(a.Stat())
	Output = int(s)
}

func BenchmarkNewU32_multi(b *testing.B) {

	n := int32(1024 * 10)
	step := int32(128)
	ns := testutil.RandU32Slice(0, n, step)

	s := uint32(0)

	nthread := 8

	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < nthread; i++ {
		wg.Add(1)
		go func() {
			var a *SlimArray
			for i := 0; i < b.N/nthread/int(n)+1; i++ {
				a = NewU32(ns)
				s += a.Get(int32(0))
			}

			wg.Done()
		}()

	}

	wg.Wait()

	Output = int(s)
}
