package reaper

import (
	"modi/calculator/config"
	"modi/utils"
	"modi/utils/sorted"
	"sort"
	"strings"
)

func AggregationPath(pathCount map[string]int) []string {
	for {
		paths := utils.MapKeyToList(pathCount)
		sort.Sort(sorted.ByCharCount(paths, "/"))
		aggregated := 0
		for i := range paths {
			path := paths[len(paths)-1-i] // reverse loop
			if path == "/" {
				continue
			}

			oldPath := path
			countNum := pathCount[path]
			lastAggreNum := aggregated

			if strings.Index(path, "http://") == 0 {
				path = utils.CutHttpHead(path)
				aggregated += 1

			}

			parts := strings.Split(path, "/")
			depth := len(parts) - 1
			lastLen := len(parts[depth])

			if depth > config.PathMaxDepth {
				path = strings.Join(parts[:config.PathMaxDepth+1], "/")
				aggregated += 1
				depth = config.PathMaxDepth
			}

			if countNum < config.PathAggreNumber || lastLen > config.PathLastlength || len(path) > config.PathAggrelength {
				if depth == 1 {
					path = "/"
				} else {
					path = strings.Join(parts[:len(parts)-1], "/")
				}
				aggregated += 1
			}

			if aggregated > lastAggreNum {
				if _, ok := pathCount[path]; ok {
					pathCount[path] += countNum

				} else {
					pathCount[path] = countNum
				}
				delete(pathCount, oldPath)
			}
		}
		if aggregated == 0 {
			break
		}
	}
	final := utils.MapKeyToList(pathCount)
	sStruct := sorted.ByCharCount(final, "/")
	sort.Sort(sStruct)
	return final
}
