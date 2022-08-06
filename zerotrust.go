/*
Copyright © 2022 Adedayo Adetoye (aka Dayo)
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/0-trust/service/cmd"
)

var (
	version = "0.0.0" // deployed version will be taken from release tags
)

func main() {

	cmd.Execute(version)

	// if file, err := os.Open("TestOTM.yaml"); err == nil {
	// 	if model, err := otm.Parse(file); err == nil {
	// 		otm_transform.OtmToGraphviz(model)
	// 	} else {
	// 		log.Printf("%v", err)
	// 	}
	// } else {
	// 	log.Printf("%v", err)
	// }
}

func readCSV(file string) ZoneFlows {

	zf := ZoneFlows{
		ZoneToFlows: make(map[string][]OutFlow),
	}

	if in, err := os.Open(file); err == nil {
		reader := csv.NewReader(in)
		start := time.Now()
		for {
			rec, err := reader.Read()

			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("%v", err)
				break
			}

			weight := 0
			if w, err := strconv.Atoi(rec[5]); err == nil {
				weight = w
			}

			sourceZone := rec[0]

			flow := OutFlow{
				Source:     toBasicHostName(rec[3]),
				Target:     toBasicHostName(rec[4]),
				SourceZone: sourceZone,
				TargetZone: rec[1],
				Label:      rec[2],
				Weight:     float32(weight),
			}

			if flows, exist := zf.ZoneToFlows[sourceZone]; exist {
				zf.ZoneToFlows[sourceZone] = append(flows, flow)
			} else {
				zf.ZoneToFlows[sourceZone] = []OutFlow{flow}
			}
		}

		end := time.Now()
		log.Printf("Time taken = %f", end.Sub(start).Seconds())

	}

	return zf
}

type ZoneFlows struct {
	ZoneToFlows map[string][]OutFlow
}

func (zf ZoneFlows) GenerateDotGraph() (out string) {

	connects := ""
	zones := ""
	zoneAssets := make(map[string]map[string]struct{})
	nothing := struct{}{}
	for zone, flows := range zf.ZoneToFlows {
		for _, of := range flows {
			if assets, exist := zoneAssets[zone]; exist {
				assets[of.Source] = nothing
				zoneAssets[zone] = assets
			} else {
				zoneAssets[zone] = map[string]struct{}{of.Source: nothing}
			}

			connects = fmt.Sprintf(`
         "%s" -> "%s" 
         %s
         `, of.Source, of.Target, connects)
		}
	}

	for zone, assets := range zoneAssets {
		zones = fmt.Sprintf(`
      %s
      %s
      `, zones, generateAssetSubgraph(zone, assets))
	}

	return fmt.Sprintf(`
   digraph {
      %s
      %s
   }
   `, connects, zones)

}

func toBasicHostName(in string) string {
	return strings.Split(in, ".")[0]

}

func generateAssetSubgraph(zone string, assets map[string]struct{}) string {

	assetList := "\n"
	for asset := range assets {
		assetList += fmt.Sprintf(`
           "%s"
      `, asset)
	}

	return fmt.Sprintf(`
   subgraph cluster_%s {
      label="%s"
      %s
   }
   `, zone, zone, assetList)

}

type OutFlow struct {
	Source, Target, Label  string
	SourceZone, TargetZone string
	Weight                 float32
}
