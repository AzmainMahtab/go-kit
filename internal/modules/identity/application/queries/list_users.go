package queries

import (
	"context"

	"github.com/elite4print/elite4print-go/internal/modules/identity/application"
	"github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/shared/pagination"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
)

// ListUsers handles paginated user listing.
type ListUsers struct {
	repo domain.UserRepository
	v    validator.Validator
}

// NewListUsers creates the query handler.
func NewListUsers(repo domain.UserRepository, v validator.Validator) *ListUsers {
	return &ListUsers{repo: repo, v: v}
}

// Execute runs the query.
func (q *ListUsers) Execute(ctx context.Context, query application.ListUsersQuery) (pagination.Page[application.UserResult], error) {
	if err := q.v.ValidateStruct(query); err != nil {
		return pagination.Page[application.UserResult]{}, err
	}

	users, total, err := q.repo.List(ctx, query.Offset, query.Limit)
	if err != nil {
		return pagination.Page[application.UserResult]{}, err
	}

	items := make([]application.UserResult, len(users))
	for i, u := range users {
		items[i] = application.ToUserResult(u)
	}

	return pagination.Page[application.UserResult]{
		Items:  items,
		Total:  total,
		Offset: query.Offset,
		Limit:  query.Limit,
	}, nil
}
