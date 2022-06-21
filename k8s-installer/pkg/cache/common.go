package cache

import (
	"errors"
	"fmt"
	"k8s-installer/schema"
	"sort"
)

func pageUp(resultLen, pageSize, pageIndex int64) (totalPage, pageStart, pageEnd int64) {
	if pageSize >= resultLen {
		totalPage = 1
	} else {
		totalPage = resultLen / pageSize
		if resultLen%pageSize > 0 {
			totalPage += 1
		}
	}
	// get start index
	pageStart = pageIndex * pageSize
	// get end index
	pageEnd = pageSize + pageStart
	if pageSize > (resultLen - pageStart) {
		pageEnd = resultLen
	}
	return
}

func getAvailableNodesMap(fullNodeCollection schema.NodeInformationCollection, pageIndex, pageSize int64,
	filter func(information schema.NodeInformation) *schema.NodeInformation) (schema.NodeInformationCollection, int64, error) {

	var nodeIdList []string

	if filter == nil {
		for _, node := range fullNodeCollection {
			nodeIdList = append(nodeIdList, node.Id)
		}
	} else {
		for _, node := range fullNodeCollection {
			selectedNode := filter(node)
			if selectedNode != nil {
				nodeIdList = append(nodeIdList, node.Id)
			}
		}
	}

	if len(nodeIdList) == 0 {
		return nil, 0, nil
	}

	sort.Strings(nodeIdList)

	result := schema.NodeInformationCollection{}

	totalPage, pageStart, pageEnd := pageUp(int64(len(nodeIdList)), pageSize, pageIndex)

	if pageStart >= int64(len(nodeIdList)) {
		// page index do not exists set data to empty
		return nil, totalPage, errors.New(fmt.Sprintf("Page %d doest not exist", pageIndex))
	} else {
		// return ranged data

		for i := pageStart; i < pageEnd; i++ {
			result[nodeIdList[i]] = fullNodeCollection[nodeIdList[i]]
		}
	}

	return result, totalPage, nil
}
