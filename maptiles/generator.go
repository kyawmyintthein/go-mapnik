package maptiles

import (
	"fmt"
	"github.com/kyawmyintthein/go-mapnik/mapnik"
	"io/ioutil"
	"log"
	"math"
	"os"
)

type Generator struct {
	MapFile string
	TileDir string
	Threads int
}

func ensureDirExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
}

// Generates tile files as a <zoom>/<x>/<y>.png file hierarchy in the current
// work directory.
func (g *Generator) Run(lowLeft, upRight mapnik.Coord, minZ, maxZ uint64, name string) {
	c := make(chan TileCoord)
	q := make(chan bool)

	log.Println("starting job", name)

	// i := uint64(23473824)
	// newGoogleProjection(maxZ + i)

	for i := 0; i < g.Threads; i++ {
		go func(id int, ctc <-chan TileCoord, q chan bool) {
			requests := NewTileRendererChan(g.MapFile)
			results := make(chan TileFetchResult)
			for t := range ctc {
				requests <- TileFetchRequest{t, results}
				r := <-results
				ioutil.WriteFile(r.Coord.OSMFilename(), r.BlobPNG, 0644)
			}
			q <- true
		}(i, c, q)
	}

	ll0 := [2]float64{lowLeft.X, upRight.Y}
	ll1 := [2]float64{upRight.X, lowLeft.Y}

	for z := minZ; z <= (maxZ + 1); z++ {
		px0 := fromLLtoPixel(ll0, z)
		px1 := fromLLtoPixel(ll1, z)
		ensureDirExists(fmt.Sprintf("%s/%d", g.TileDir,z))

		for x := uint64(px0[0] / 256.0); x <= uint64(px1[0]/256.0)+1; x++ {
			if (x < 0) || (float64(x) >= math.Pow(float64(2),float64(z))){
                continue
			}
			ensureDirExists(fmt.Sprintf("%s/%d/%d", g.TileDir, z, x))
			
			for y := uint64(px0[1] / 256.0); y <= uint64(px1[1]/256.0); y++ {
				if (y < 0) || (float64(y) >= math.Pow(float64(2),float64(z))){
                	continue
				}
				c <- TileCoord{x, (uint64((math.Pow(float64(2),float64(z))-float64(1)) - float64(y))), z, false, "", g.TileDir}
			}
		}
	}
	close(c)
	log.Println("Done job", name)
	for i := 0; i < g.Threads; i++ {
		<-q
	}
}
