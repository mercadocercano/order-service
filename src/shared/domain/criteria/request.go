package criteria

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Operadores de filtro soportados
const (
	OpEqual              = "="
	OpNotEqual           = "!="
	OpGreaterThan        = ">"
	OpGreaterThanOrEqual = ">="
	OpLessThan           = "<"
	OpLessThanOrEqual    = "<="
	OpLike               = "LIKE"
	OpIn                 = "IN"
	OpIsNull             = "NULL"
	OpIsNotNull          = "NOT NULL"
	OpArrayContains      = "ARRAY_CONTAINS"
)

// BaseListRequest representa la estructura base para requests de listado
type BaseListRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	SortBy   string `form:"sort_by" json:"sort_by"`
	SortDir  string `form:"sort_dir" json:"sort_dir"`
}

// ToCriteria convierte el request base a un criteria básico
func (r BaseListRequest) ToCriteria() Criteria {
	// Validar y ajustar valores por defecto
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 {
		r.PageSize = 10
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	if r.SortBy == "" {
		r.SortBy = "created_at"
	}
	if r.SortDir == "" {
		r.SortDir = "DESC"
	}

	// Convertir dirección de ordenamiento
	var orderType OrderType
	if strings.ToUpper(r.SortDir) == "ASC" {
		orderType = ASC
	} else {
		orderType = DESC
	}

	// Calcular limit y offset
	limit := r.PageSize
	offset := (r.Page - 1) * r.PageSize

	return NewCriteria(
		NewFilters(),
		NewOrder(r.SortBy, orderType),
		&limit,
		&offset,
	)
}

// CriteriaBuilder facilita la construcción de criterios usando el patrón builder
type CriteriaBuilder struct {
	filters    []Filter
	orderField string
	orderDir   OrderType
	page       int
	pageSize   int
}

// NewCriteriaBuilder crea un nuevo builder
func NewCriteriaBuilder() *CriteriaBuilder {
	return &CriteriaBuilder{
		filters:    make([]Filter, 0),
		page:       1,
		pageSize:   10,
		orderField: "created_at",
		orderDir:   DESC,
	}
}

// FromURLValues inicializa el builder desde url.Values
func (b *CriteriaBuilder) FromURLValues(values url.Values) *CriteriaBuilder {
	// Paginación
	if page := values.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			b.page = p
		}
	}

	if pageSize := values.Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			b.pageSize = ps
		}
	}

	// Ordenamiento
	if sortBy := values.Get("sort_by"); sortBy != "" {
		b.orderField = sortBy
	}

	if sortDir := values.Get("sort_dir"); sortDir != "" {
		if strings.ToUpper(sortDir) == "ASC" {
			b.orderDir = ASC
		} else {
			b.orderDir = DESC
		}
	}

	return b
}

// AddFilter agrega un filtro genérico
func (b *CriteriaBuilder) AddFilter(field, operator string, value interface{}) *CriteriaBuilder {
	if value != nil && value != "" {
		b.filters = append(b.filters, NewFilter(field, operator, value))
	}
	return b
}

// AddEqualFilter agrega un filtro de igualdad
func (b *CriteriaBuilder) AddEqualFilter(field string, value interface{}) *CriteriaBuilder {
	return b.AddFilter(field, OpEqual, value)
}

// AddNotEqualFilter agrega un filtro de desigualdad
func (b *CriteriaBuilder) AddNotEqualFilter(field string, value interface{}) *CriteriaBuilder {
	return b.AddFilter(field, OpNotEqual, value)
}

// AddLikeFilter agrega un filtro LIKE para búsquedas de texto
func (b *CriteriaBuilder) AddLikeFilter(field string, value interface{}) *CriteriaBuilder {
	if str, ok := value.(string); ok && str != "" {
		// Si no tiene wildcards, los agregamos
		if !strings.Contains(str, "%") {
			value = "%" + str + "%"
		}
		return b.AddFilter(field, OpLike, value)
	}
	return b
}

// AddGreaterThanFilter agrega un filtro mayor que
func (b *CriteriaBuilder) AddGreaterThanFilter(field string, value interface{}) *CriteriaBuilder {
	return b.AddFilter(field, OpGreaterThan, value)
}

// AddGreaterThanOrEqualFilter agrega un filtro mayor o igual que
func (b *CriteriaBuilder) AddGreaterThanOrEqualFilter(field string, value interface{}) *CriteriaBuilder {
	return b.AddFilter(field, OpGreaterThanOrEqual, value)
}

// AddLessThanFilter agrega un filtro menor que
func (b *CriteriaBuilder) AddLessThanFilter(field string, value interface{}) *CriteriaBuilder {
	return b.AddFilter(field, OpLessThan, value)
}

// AddLessThanOrEqualFilter agrega un filtro menor o igual que
func (b *CriteriaBuilder) AddLessThanOrEqualFilter(field string, value interface{}) *CriteriaBuilder {
	return b.AddFilter(field, OpLessThanOrEqual, value)
}

// AddInFilter agrega un filtro IN para arrays
func (b *CriteriaBuilder) AddInFilter(field string, values []interface{}) *CriteriaBuilder {
	if len(values) > 0 {
		return b.AddFilter(field, OpIn, values)
	}
	return b
}

// AddUUIDFilter agrega un filtro para UUID validando el formato
func (b *CriteriaBuilder) AddUUIDFilter(field string, value interface{}) *CriteriaBuilder {
	if str, ok := value.(string); ok && str != "" {
		if _, err := uuid.Parse(str); err == nil {
			return b.AddEqualFilter(field, str)
		}
	}
	return b
}

// AddBoolFilter agrega un filtro booleano
func (b *CriteriaBuilder) AddBoolFilter(field string, value interface{}) *CriteriaBuilder {
	if str, ok := value.(string); ok {
		if str == "true" {
			return b.AddEqualFilter(field, true)
		} else if str == "false" {
			return b.AddEqualFilter(field, false)
		}
	}
	if boolVal, ok := value.(bool); ok {
		return b.AddEqualFilter(field, boolVal)
	}
	return b
}

// SetOrder establece el ordenamiento
func (b *CriteriaBuilder) SetOrder(field string, direction OrderType) *CriteriaBuilder {
	if field != "" {
		b.orderField = field
		b.orderDir = direction
	}
	return b
}

// SetPagination establece la paginación
func (b *CriteriaBuilder) SetPagination(page, pageSize int) *CriteriaBuilder {
	if page > 0 {
		b.page = page
	}
	if pageSize > 0 && pageSize <= 100 {
		b.pageSize = pageSize
	}
	return b
}

// Build construye el criteria final
func (b *CriteriaBuilder) Build() Criteria {
	filters := NewFilters(b.filters...)

	order := NewOrder(b.orderField, b.orderDir)

	// Calcular limit y offset
	limit := b.pageSize
	offset := (b.page - 1) * b.pageSize

	return NewCriteria(filters, order, &limit, &offset)
}

// AddArrayContainsFilter agrega un filtro para verificar si un array contiene un valor específico
func (b *CriteriaBuilder) AddArrayContainsFilter(field string, value interface{}) *CriteriaBuilder {
	if value != nil && value != "" {
		return b.AddFilter(field, OpArrayContains, value)
	}
	return b
}
