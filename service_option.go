// Code generated by "gogen option -n Service -r repo,aggConstruct --with-init"; DO NOT EDIT.

package dgo

import "fmt"

// NewService constructor
func NewService[A AggBase](repo Repo[A], aggConstruct func() A, _opts ...ServiceOption[A]) *Service[A] {
	_s := new(Service[A])

	_s.repo = repo
	_s.aggConstruct = aggConstruct

	_s.SetOptions(_opts...)

	_s.init()
	return _s
}

// ServiceOption[A AggBase] option interface
type ServiceOption[A AggBase] interface {
	apply(*Service[A])
}

// ServiceOption[A AggBase] option function
type serviceOptionFunc[A AggBase] func(*Service[A])

func (f serviceOptionFunc[A]) apply(_s *Service[A]) {
	f(_s)
}

func (_s *Service[A]) SetOptions(_opts ...ServiceOption[A]) *Service[A] {
	for _, _opt := range _opts {
		_opt.apply(_s)
	}
	return _s
}

func SkipServiceOption[A AggBase]() ServiceOption[A] {
	return serviceOptionFunc[A](func(_s *Service[A]) {
		return
	})
}

func WithServiceOptions[A AggBase](o *serviceOptions[A]) ServiceOption[A] {
	return serviceOptionFunc[A](func(_s *Service[A]) {
		_s.SetOptions(o.opts...)
	})
}

// serviceOptions[A AggBase] options struct
type serviceOptions[A AggBase] struct {
	opts []ServiceOption[A]
}

// NewServiceOptions[A AggBase] new options struct
func NewServiceOptions[A AggBase]() *serviceOptions[A] {
	return new(serviceOptions[A])
}

func (_o *serviceOptions[A]) Options() []ServiceOption[A] {
	return _o.opts
}

func (_o *serviceOptions[A]) Append(_opts ...ServiceOption[A]) *serviceOptions[A] {
	_o.opts = append(_o.opts, _opts...)
	return _o
}

// Bus bus option of Service
func (_o *serviceOptions[A]) Bus(bus Bus) *serviceOptions[A] {
	_o.opts = append(_o.opts, WithServiceBus[A](bus))
	return _o
}

// IdGenerator idGenerator option of Service
func (_o *serviceOptions[A]) IdGenerator(idGenerator IDGenerator) *serviceOptions[A] {
	_o.opts = append(_o.opts, WithServiceIdGenerator[A](idGenerator))
	return _o
}

// SnapshotSaveStrategy snapshotSaveStrategy option of Service
func (_o *serviceOptions[A]) SnapshotSaveStrategy(snapshotSaveStrategy SnapshotSaveStrategy[A]) *serviceOptions[A] {
	_o.opts = append(_o.opts, WithServiceSnapshotSaveStrategy[A](snapshotSaveStrategy))
	return _o
}

// WithServiceBus bus option of Service
func WithServiceBus[A AggBase](bus Bus) ServiceOption[A] {
	return serviceOptionFunc[A](func(_s *Service[A]) {
		_s.bus = bus
	})
}

// WithServiceIdGenerator idGenerator option of Service
func WithServiceIdGenerator[A AggBase](idGenerator IDGenerator) ServiceOption[A] {
	return serviceOptionFunc[A](func(_s *Service[A]) {
		_s.idGenerator = idGenerator
	})
}

// WithServiceSnapshotSaveStrategy snapshotSaveStrategy option of Service
func WithServiceSnapshotSaveStrategy[A AggBase](snapshotSaveStrategy SnapshotSaveStrategy[A]) ServiceOption[A] {
	return serviceOptionFunc[A](func(_s *Service[A]) {
		_s.snapshotSaveStrategy = snapshotSaveStrategy
	})
}

func PrintServiceOptions(packageName string) {
	opts := []string{
		"WithServiceBus()",
		"WithServiceIdGenerator()",
		"WithServiceSnapshotSaveStrategy()",
	}
	if packageName == "" {
		fmt.Printf("opts := []ServiceOption{ \n")
		for _, v := range opts {
			fmt.Printf("	%s,\n", v)
		}
	} else {
		fmt.Printf("opts := []%s.ServiceOption{ \n", packageName)
		for _, v := range opts {
			fmt.Printf("	%s.%s,\n", packageName, v)
		}
	}
	fmt.Println("}")
}
