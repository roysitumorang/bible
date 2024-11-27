package models

type (
	Pagination struct {
		Links struct {
			First    string `json:"first"`
			Previous string `json:"previous"`
			Current  string `json:"current"`
			Next     string `json:"next"`
		} `json:"links"`
		Info struct {
			PerPage int `json:"per_page"`
			Pages   int `json:"pages"`
			Total   int `json:"total"`
		} `json:"info"`
	}

	RootCategory struct {
		ID           int64  `json:"-"`
		UID          string `json:"id"`
		CategoryID   int64  `json:"category_id"`
		CategoryName string `json:"category_name"`
	}
)

const (
	SourceEproc       int = 0
	SourceGokobiz     int = 1
	SourceVirtual     int = 2
	SourceDirectTC    int = 3
	SourceRepeatOrder int = 4

	SourceSubcategoryVirtualEproc    = 20
	SourceSubcategoryVirtualEcatalog = 21
	SourceSubcategoryVirtualSales    = 22

	RoleBuyer      int = 0
	RoleSeller     int = 1
	RoleSourcing   int = 2
	RoleCommercial int = 3
)

var (
	MapPerPageRowsCount = map[int]int{1: 1, 10: 1, 25: 1, 50: 1, 100: 1}
	PerPageRowsCount    = []int{1, 10, 25, 50, 100}
)
