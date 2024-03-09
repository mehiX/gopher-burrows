package burrows

import (
	"os"
	"text/tabwriter"
)

func ExampleReport_ToTxt() {

	r := Report{
		TotalDepth:    10.23434,
		NumAvailable:  145,
		VolumeMin:     34.81231,
		VolumeMinName: "Burrow 3",
		VolumeMax:     78.312313,
		VolumeMaxName: "Burrow 123",
	}

	w := tabwriter.NewWriter(os.Stdout, 15, 0, 0, '.', tabwriter.AlignRight|tabwriter.Debug)
	r.ToTxt(w)
	w.Flush()

	// output:
	// .....TotalDepth|.........10.234|
	// ...NumAvailable|............145|
	// ..VolumeMinName|.......Burrow 3|
	// ..VolumeMaxName|.....Burrow 123|
}
