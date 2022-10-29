package models

import "math"

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

func Paginate(collection []interface{}, page int, _perPage int) *PaginationData {
	var paginationData *PaginationData = new(PaginationData)
	var totalPages int = int(math.Ceil(float64(len(collection)) / float64(_perPage)))
	var recordsOnPage int = int(math.Min(float64(_perPage), float64(len(collection)-((page-1)*_perPage))))
	var data []interface{} = nil

	for i := ((page - 1) * _perPage); i < ((page-1)*_perPage)+recordsOnPage; i++ {
		data = append(data, collection[i])
	}

	paginationData.Data = data
	paginationData.Pagination.CurrentPage = page
	paginationData.Pagination.TotalPages = totalPages
	paginationData.Pagination.PerPage = _perPage
	paginationData.Pagination.RecordsOnPage = recordsOnPage
	paginationData.Pagination.TotalRecords = len(collection)
	return paginationData
}
