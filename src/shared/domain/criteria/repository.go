package criteria

import (
	"context"
)

// CriteriaRepository define una interfaz genérica para repositorios que soportan criteria
type CriteriaRepository[T any] interface {
	// SearchByCriteria busca entidades usando los criterios especificados
	SearchByCriteria(ctx context.Context, criteria Criteria) ([]*T, error)

	// CountByCriteria cuenta las entidades que coinciden con los criterios
	CountByCriteria(ctx context.Context, criteria Criteria) (int, error)
}

// ListRepository define una interfaz para operaciones de listado con criteria
type ListRepository[T any] interface {
	CriteriaRepository[T]

	// ListByCriteria combina búsqueda y conteo para generar respuesta de listado
	ListByCriteria(ctx context.Context, criteria Criteria) (*ListResponse[T], error)
}

// BaseListRepository implementación base que puede ser embebida por repositorios concretos
type BaseListRepository[T any] struct {
	criteriaRepo CriteriaRepository[T]
}

// NewBaseListRepository crea una nueva instancia del repositorio base
func NewBaseListRepository[T any](criteriaRepo CriteriaRepository[T]) *BaseListRepository[T] {
	return &BaseListRepository[T]{
		criteriaRepo: criteriaRepo,
	}
}

// ListByCriteria implementa la lógica común para listado con criteria
func (r *BaseListRepository[T]) ListByCriteria(ctx context.Context, criteria Criteria) (*ListResponse[T], error) {
	// Obtener elementos
	items, err := r.criteriaRepo.SearchByCriteria(ctx, criteria)
	if err != nil {
		return nil, err
	}

	// Obtener conteo total
	total, err := r.criteriaRepo.CountByCriteria(ctx, criteria)
	if err != nil {
		return nil, err
	}

	// Crear respuesta
	return NewListResponse(items, total, criteria), nil
}
