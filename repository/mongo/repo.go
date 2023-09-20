package mongo

import (
	"context"

	"github.com/ekuu/dgo"
	"github.com/ekuu/dgo/internal"
	repo "github.com/ekuu/dgo/repository"
	"github.com/pkg/errors"
	pkgerr "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	mg "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate gogen option -n Repo -r db,agg,event,convert,reverse,newData,parseID --with-init
type Repo[I repo.ID, A dgo.AggBase, D any] struct {
	db               *mg.Database
	agg, event       *mg.Collection
	convert          func(context.Context, A) (D, error)
	reverse          func(context.Context, dgo.AggBase, D) (A, error)
	newData          func() D
	parseID          repo.ParseID[I]
	getVid           repo.NewVid[I]
	newEvent         func() repo.Event[I]
	newAggregate     func() Aggregate[I, D]
	versionFieldName string
	closeTransaction bool
}

func (r *Repo[I, A, D]) init() {
	if r.newEvent == nil {
		r.newEvent = NewEvent[I]
	}
	if r.newAggregate == nil {
		r.newAggregate = NewAggregate[I, D]
	}
	if r.getVid == nil {
		r.getVid = func(id dgo.ID, version uint64) (repo.Vid[I], error) {
			v, err := r.parseID(id)
			if err != nil {
				return nil, err
			}
			return NewVid[I](v, version), nil
		}
	}
	if r.versionFieldName == "" {
		r.versionFieldName = "version"
	}
}

func NewDefaultRepo[I repo.ID, A dgo.AggBase, D any](
	db *mg.Database,
	aggrName string,
	convert func(context.Context, A) (D, error),
	reverse func(context.Context, dgo.AggBase, D) (A, error),
	newData func() D,
	parseID repo.ParseID[I],
	_opts ...RepoOption[I, A, D],
) *Repo[I, A, D] {
	return NewRepo[I, A, D](
		db,
		db.Collection(aggrName),
		db.Collection(aggrName+"Event"),
		convert,
		reverse,
		newData,
		parseID,
		_opts...,
	)
}

func (r *Repo[I, A, D]) DB() *mg.Database {
	return r.db
}

func (r *Repo[I, A, D]) Agg() *mg.Collection {
	return r.agg
}

func (r *Repo[I, A, D]) Event() *mg.Collection {
	return r.event
}

func (r *Repo[I, A, D]) da2pa(ctx context.Context, da A) (Aggregate[I, D], error) {
	data, err := r.convert(ctx, da)
	if err != nil {
		return nil, err
	}
	pa, err := repo.ConvertAggBase(da, r.newAggregate(), r.parseID)
	if err != nil {
		return nil, err
	}
	pa.SetData(data)
	return pa, nil
}

func (r *Repo[I, A, D]) PA2DA(ctx context.Context, pa Aggregate[I, D]) (A, error) {
	return r.reverse(ctx, repo.ReverseAggBase[I](pa), pa.GetData())
}

func (r *Repo[I, A, D]) ParseID(id dgo.ID) (I, error) {
	return r.parseID(id)
}

func (r *Repo[I, A, D]) VersionFieldName() string {
	return r.versionFieldName
}

func (r *Repo[I, A, D]) Get(ctx context.Context, id dgo.ID) (a A, err error) {
	rid, err := r.parseID(id)
	if err != nil {
		return a, err
	}
	return r.FindDA(ctx, bson.M{"_id": rid})
}

func (r *Repo[I, A, D]) List(ctx context.Context, ids ...dgo.ID) (as []A, err error) {
	rids, err := internal.MapError(ids, func(i int, id dgo.ID) (I, error) {
		return r.parseID(id)
	})
	if err != nil {
		return nil, err
	}
	return r.FindDAs(ctx, bson.M{"_id": bson.M{"$in": rids}})
}

func (r *Repo[I, A, D]) Save(ctx context.Context, a A) error {
	pa, err := r.da2pa(ctx, a)
	if err != nil {
		return err
	}
	rs, err := r.agg.UpdateOne(ctx, bson.M{"_id": pa.GetID(), r.versionFieldName: a.OriginalVersion()}, bson.M{"$set": pa.GetContent()}, options.Update().SetUpsert(true))
	if err != nil {
		return pkgerr.WithStack(err)
	}
	if rs.UpsertedCount == 0 && rs.MatchedCount == 0 {
		return pkgerr.WithStack(dgo.ErrNotMatched)
	}
	return nil
}

func (r *Repo[I, A, D]) Delete(ctx context.Context, a A) error {
	id, err := r.parseID(a.ID())
	if err != nil {
		return err
	}
	_, err = r.agg.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		err = pkgerr.WithStack(err)
	}
	return err
}

func (r *Repo[I, A, D]) SaveEvents(ctx context.Context, unsaved dgo.Events) error {
	l := len(unsaved)
	if l == 0 {
		return nil
	}
	models := make([]mg.WriteModel, l)
	for i, v := range unsaved {
		e, err := repo.ConvertEvent(v, r.newEvent, r.getVid)
		if err != nil {
			return pkgerr.WithStack(err)
		}
		models[i] = mg.NewUpdateOneModel().SetFilter(bson.M{"_id": e.GetID()}).SetUpdate(bson.M{"$set": e}).SetUpsert(true)
	}
	_, err := r.event.BulkWrite(ctx, models)
	return err
}

func (r *Repo[I, A, D]) Transaction(ctx context.Context, fn func(ctx context.Context, r dgo.Repo[A]) error) error {
	if r.closeTransaction {
		return fn(ctx, r)
	}
	return r.db.Client().UseSession(ctx, func(sessCtx mg.SessionContext) error {
		if err := sessCtx.StartTransaction(); err != nil {
			return err
		}
		err := fn(sessCtx, r)
		if err != nil {
			abortErr := sessCtx.AbortTransaction(ctx)
			if abortErr != nil {
				return errors.Wrapf(err, abortErr.Error())
			}
			return err
		} else {
			return sessCtx.CommitTransaction(ctx)
		}
	})
}

func (r *Repo[I, A, D]) FindPA(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (Aggregate[I, D], error) {
	pa := r.newAggregate()
	pa.SetData(r.newData())
	err := ReplaceError(r.agg.FindOne(ctx, filter, opts...).Decode(pa))
	if err != nil {
		err = pkgerr.WithMessagef(err, "mongo collection %s, filter: %+v", r.agg.Name(), filter)
	}
	return pa, err
}

func (r *Repo[I, A, D]) FindDA(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) (d A, err error) {
	pa, err := r.FindPA(ctx, filter, opts...)
	if err != nil {
		return
	}
	return r.PA2DA(ctx, pa)
}

func (r *Repo[I, A, D]) FindPAs(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (pas []Aggregate[I, D], err error) {
	cursor, err := r.agg.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		pa := r.newAggregate()
		pa.SetData(r.newData())
		if err = cursor.Decode(pa); err != nil {
			return nil, err
		}
		pas = append(pas, pa)
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return pas, nil
}

func (r *Repo[I, A, D]) FindDAs(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (das []A, err error) {
	pas, err := r.FindPAs(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return internal.MapError(pas, func(i int, p Aggregate[I, D]) (A, error) {
		return r.PA2DA(ctx, p)
	})
}

func ReplaceError(err error) error {
	if err == nil {
		return nil
	}
	if mg.ErrNoDocuments == err {
		return pkgerr.WithStack(dgo.ErrNotFound)
	}
	return errors.WithStack(err)
}
