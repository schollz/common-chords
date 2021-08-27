package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	log "github.com/schollz/logger"
	"github.com/schollz/progressbar/v3"
)

func main() {
	log.SetLevel("debug")
	err := Analyze()
	if err != nil {
		log.Error(err)
	}
}

func Analyze() (err error) {
	log.Debug("loading chords")
	chordIndex := make(map[string]map[string]float64)
	b, err := ioutil.ReadFile("chordIndexInC.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &chordIndex)
	if err != nil {
		return
	}

	allowOnlyFourDifferentChords := true

	log.Debug("generating markov chains")
	rand.Seed(time.Now().UTC().UnixNano())
	fourChords := make(map[string]float64)
	numIterations := int64(100000)
	bar := progressbar.Default(numIterations)
	for j := int64(0); j < numIterations; j++ {
		bar.Add(1)
		chordList := ""
		badChords := false
		for i := 0; i < 4; i++ {
			if i == 0 {
				chordList += randomWeightedChoice(chordIndex["init"])
			} else {
				if _, ok := chordIndex[chordList]; !ok {
					badChords = true
					break
				}
				choice := randomWeightedChoice(chordIndex[chordList])
				if allowOnlyFourDifferentChords && strings.Contains(chordList, choice) {
					badChords = true
					break
				}
				chordList += " " + choice
			}
		}
		if badChords {
			continue
		}
		if _, ok := fourChords[chordList]; !ok {
			fourChords[chordList] = 0
		}
		fourChords[chordList]++
	}

	// normalize fourchords
	sum := 0.0
	for k := range fourChords {
		sum += fourChords[k]
	}

	for k := range fourChords {
		fourChords[k] = math.Round(fourChords[k]/sum*100000) / 1000
	}

	type kv struct {
		Key   string
		Value float64
	}
	var ss []kv
	for k, v := range fourChords {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for i := 0; i < 1000; i++ {
		fmt.Println(ss[i].Key, ss[i].Value)
	}

	return
}

func randomWeightedChoice(m map[string]float64) string {
	type kv struct {
		Key   string
		Value float64
	}
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Key > ss[j].Key
	})

	curSum := 0.0
	target := rand.Float64() * 100
	for _, kv := range ss {
		curSum += kv.Value
		if curSum >= target {
			return kv.Key
		}
	}
	return ""
}
