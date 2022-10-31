package models

import (
	"math"
	"sort"
	"time"
)

type Pagination struct {
	CurrentPage   int `json:"currentPage"`
	TotalPages    int `json:"totalPages"`
	PerPage       int `json:"perPage"`
	RecordsOnPage int `json:"recordsOnPage"`
	TotalRecords  int `json:"totalRecords"`
}

type PaginationData struct {
	Pagination Pagination  `json:"pagination"`
	Data       interface{} `json:"data"`
}

type SortablePaginationData interface {
	interface{}
	GetCreationTime() time.Time
}

func Paginate(collection []interface{}, page int, _perPage int) *PaginationData {
	var paginationData *PaginationData = new(PaginationData)
	var totalPages int = int(math.Ceil(float64(len(collection)) / float64(_perPage)))
	var recordsOnPage int = int(math.Min(float64(_perPage), float64(len(collection)-((page-1)*_perPage))))
	var data []interface{} = nil

	// If the page is out of bounds, return an empty page
	if page > 0 && recordsOnPage > 0 {
		// Sort elements by creation time
		sort.Slice(collection, func(i, j int) bool {
			return collection[i].(SortablePaginationData).GetCreationTime().Before(collection[j].(SortablePaginationData).GetCreationTime())
		})

		// Get elements from requested page
		data = collection[((page - 1) * _perPage) : ((page-1)*_perPage)+recordsOnPage]
	} else {
		recordsOnPage = 0
	}

	// Prepare pagination response
	paginationData.Data = data
	paginationData.Pagination.CurrentPage = page
	paginationData.Pagination.TotalPages = totalPages
	paginationData.Pagination.PerPage = _perPage
	paginationData.Pagination.RecordsOnPage = recordsOnPage
	paginationData.Pagination.TotalRecords = len(collection)
	return paginationData
}
