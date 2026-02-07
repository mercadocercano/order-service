package criteria

import (
	"math"
)

// Criteria representa un conjunto de criterios para consultas
type Criteria struct {
	Filters Filters
	Order   Order
	Limit   *int
	Offset  *int
}

// NewCriteria crea un nuevo objeto Criteria
func NewCriteria(filters Filters, order Order, limit, offset *int) Criteria {
	return Criteria{
		Filters: filters,
		Order:   order,
		Limit:   limit,
		Offset:  offset,
	}
}

// IsEmpty verifica si el criteria está vacío
func (c Criteria) IsEmpty() bool {
	return len(c.Filters.Items) == 0 && c.Order.Field == "" && c.Limit == nil && c.Offset == nil
}

// Filters representa una colección de filtros
type Filters struct {
	Items []Filter
}

// Filter representa un filtro individual
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// NewFilter crea un nuevo filtro
func NewFilter(field, operator string, value interface{}) Filter {
	return Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

// NewFilters crea un conjunto de filtros
func NewFilters(filters ...Filter) Filters {
	return Filters{Items: filters}
}

// Add agrega un filtro a la colección
func (f *Filters) Add(filter Filter) {
	f.Items = append(f.Items, filter)
}

// Count retorna el número de filtros
func (f Filters) Count() int {
	return len(f.Items)
}

// IsEmpty verifica si no hay filtros
func (f Filters) IsEmpty() bool {
	return len(f.Items) == 0
}

// Order representa un orden para la consulta
type Order struct {
	Field     string
	OrderType OrderType
}

// OrderType representa el tipo de orden (ascendente o descendente)
type OrderType string

const (
	// ASC representa un orden ascendente
	ASC OrderType = "asc"
	// DESC representa un orden descendente
	DESC OrderType = "desc"
)

// NewOrder crea un nuevo orden
func NewOrder(field string, orderType OrderType) Order {
	return Order{
		Field:     field,
		OrderType: orderType,
	}
}

// IsEmpty verifica si el orden está vacío
func (o Order) IsEmpty() bool {
	return o.Field == ""
}

// Pagination representa los criterios de paginación
type Pagination struct {
	Page     int
	PageSize int
	Limit    int
	Offset   int
}

// NewPagination crea un nuevo criterio de paginación
func NewPagination(page, pageSize int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	return Pagination{
		Page:     page,
		PageSize: pageSize,
		Limit:    pageSize,
		Offset:   offset,
	}
}

// IsEmpty verifica si la paginación está vacía
func (p Pagination) IsEmpty() bool {
	return p.Limit == 0
}

// GetTotalPages calcula el número total de páginas
func (p Pagination) GetTotalPages(totalCount int) int {
	if p.PageSize == 0 {
		return 0
	}
	return int(math.Ceil(float64(totalCount) / float64(p.PageSize)))
}

// GetPageFromOffset calcula la página basada en offset y tamaño de página
func GetPageFromOffset(offset, pageSize int) int {
	if pageSize == 0 {
		return 1
	}
	return (offset / pageSize) + 1
}

// GetTotalPagesFromLimit calcula el número total de páginas basado en limit
func GetTotalPagesFromLimit(totalCount, limit int) int {
	if limit == 0 {
		return 0
	}
	return int(math.Ceil(float64(totalCount) / float64(limit)))
}

// ListResponse representa una respuesta de listado genérica
type ListResponse[T any] struct {
	Items      []*T `json:"items"`
	TotalCount int  `json:"total_count"`
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalPages int  `json:"total_pages"`
}

// NewListResponse crea una nueva respuesta de listado
func NewListResponse[T any](items []*T, totalCount int, criteria Criteria) *ListResponse[T] {
	var page, pageSize, totalPages int

	if criteria.Limit != nil {
		pageSize = *criteria.Limit
		totalPages = GetTotalPagesFromLimit(totalCount, pageSize)

		if criteria.Offset != nil {
			page = GetPageFromOffset(*criteria.Offset, pageSize)
		} else {
			page = 1
		}
	} else {
		page = 1
		pageSize = len(items)
		totalPages = 1
	}

	return &ListResponse[T]{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
