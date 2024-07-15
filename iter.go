package ghiter

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/url"
	"reflect"
	"strings"

	"github.com/TomNomNom/linkheader"
	"github.com/google/go-github/v63/github"
)

type Iterator[T, O any] struct {
	opt O

	ctx  context.Context
	args []string

	fn  func(ctx context.Context, opt O) ([]T, *github.Response, error)
	fn1 func(ctx context.Context, arg1 string, opt O) ([]T, *github.Response, error)
	fn2 func(ctx context.Context, arg1, arg2 string, opt O) ([]T, *github.Response, error)

	raw *github.Response
	err error
}

func NewFromFn[T, O any](
	fn func(ctx context.Context, opt O) ([]T, *github.Response, error),
) *Iterator[T, O] {
	return &Iterator[T, O]{
		fn: fn,
	}
}

func NewFromFn1[T, O any](
	fn1 func(ctx context.Context, arg1 string, opt O) ([]T, *github.Response, error),
	arg1 string,
) *Iterator[T, O] {
	return &Iterator[T, O]{
		fn1:  fn1,
		args: []string{arg1},
	}
}

func NewFromFn2[T, O any](
	fn2 func(ctx context.Context, arg1, arg2 string, opt O) ([]T, *github.Response, error),
	arg1 string,
	arg2 string,
) *Iterator[T, O] {
	return &Iterator[T, O]{
		fn2:  fn2,
		args: []string{arg1, arg2},
	}
}

func (it *Iterator[T, O]) Ctx(ctx context.Context) *Iterator[T, O] {
	it.ctx = ctx
	return it
}

func (it *Iterator[T, O]) Args(args ...string) *Iterator[T, O] {
	it.args = args
	return it
}

func (it *Iterator[T, O]) Opts(opt O) *Iterator[T, O] {
	it.opt = opt
	return it
}

func (it *Iterator[T, O]) Raw() *github.Response {
	return it.raw
}

func (it *Iterator[T, O]) Err() error {
	return it.err
}

func (it *Iterator[T, O]) All() iter.Seq[T] {
	initialize(it)

	return func(yield func(T) bool) {
		if err := validate(it); err != nil {
			it.err = err
			return
		}

		for {
			parts, resp, err := it.do()
			it.raw = resp

			links := linkheader.Parse(resp.Header.Get("link"))

			var nextLink string
			for _, link := range links {
				if link.Rel == "next" {
					nextLink = link.URL
					break
				}
			}

			nextURL, err := url.Parse(nextLink)
			if err != nil {
				it.err = err
				return
			}

			vals := nextURL.Query()
			for k, v := range vals {
				fmt.Println(k, v)
			}

			val := reflect.ValueOf(it.opt).Elem()

			for i := 0; i < val.NumField(); i++ {
				field := reflect.TypeOf(it.opt).Elem().Field(i)
				fieldType := val.Field(i).Kind()
				if fieldType == reflect.Struct {
					fmt.Println(field.Name, "nested")

					nested := val.Field(i)
					for i := 0; i < nested.NumField(); i++ {
						nestedfieldVal := reflect.ValueOf(nested.Field(i))
						nestedfield := reflect.TypeOf(nested).Field(i)
						urlTAG := nestedfield.Tag.Get("url")
						fmt.Println("##", nestedfield.Name, nestedfieldVal, nestedfield.Anonymous, nestedfield.Tag, urlTAG)
					}
				}

				jsonTag := field.Tag.Get("url")
				fmt.Println(field.Name, field.Anonymous, field.Tag, jsonTag)
			}

			if err != nil {
				it.err = err
				return
			}

			for _, p := range parts {
				if !yield(p) || contextErr(it) {
					return
				}
			}

			// no more results, break
			if resp.NextPage == 0 {
				return
			}

			// iterate to the next page
			val = reflect.ValueOf(it.opt).Elem()
			field := val.FieldByName("ListOptions")
			if field.IsValid() && field.CanSet() {

				listOpts := field.Interface().(github.ListOptions)
				listOpts.Page = resp.NextPage
				field.Set(reflect.ValueOf(listOpts))
			}

			// Users.ListAll
			field = val.FieldByName("Since")
			if field.IsValid() && field.CanSet() {
				fmt.Println(resp.Header.Get("link"))
				field.SetInt(int64(resp.NextPage))
			}
		}
	}
}

func initialize[T, O any](it *Iterator[T, O]) {
	// initialize context if nil
	if it.ctx == nil {
		it.ctx = context.Background()
	}

	// initialize options if nil
	if reflect.ValueOf(it.opt).IsNil() {
		optionPointerType := reflect.TypeOf(it.opt)
		optionValue := reflect.New(optionPointerType.Elem())
		if opt, ok := optionValue.Interface().(O); ok {
			it.opt = opt
		}
	}
}

func validate[T, O any](it *Iterator[T, O]) error {
	if it.fn == nil && it.fn1 == nil && it.fn2 == nil {
		return errors.New("no func provided")
	}

	numOfArgs := len(it.args)
	if it.fn1 != nil {
		if numOfArgs != 1 {
			args := strings.Join(it.args, ",")
			return fmt.Errorf("wrong number of arguments: expected 1, got %d [%s]", numOfArgs, args)
		}

		if it.args[0] == "" {
			return errors.New("empty argument[0]")
		}
	}

	if it.fn2 != nil {
		if numOfArgs != 2 {
			args := strings.Join(it.args, ",")
			return fmt.Errorf("wrong number of arguments: expected 2, got %d [%s]", numOfArgs, args)
		}

		if it.args[0] == "" {
			return errors.New("empty argument[0]")
		}

		if it.args[1] == "" {
			return errors.New("empty argument[1]")
		}
	}

	return nil
}

func contextErr[T, O any](it *Iterator[T, O]) bool {
	if err := it.ctx.Err(); err != nil {
		it.err = err
		return true
	}
	return false
}

func (it *Iterator[T, O]) do() ([]T, *github.Response, error) {
	if it.fn != nil {
		return it.fn(it.ctx, it.opt)
	} else if it.fn1 != nil {
		return it.fn1(it.ctx, it.args[0], it.opt)
	} else if it.fn2 != nil {
		return it.fn2(it.ctx, it.args[0], it.args[1], it.opt)
	}

	return nil, nil, errors.New("no func provided")
}
