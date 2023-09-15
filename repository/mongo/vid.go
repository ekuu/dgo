package mongo

import repo "github.com/ekuu/dgo/repository"

type vid[I repo.ID] struct {
	ID      I      `bson:"aid"`
	Version uint64 `bson:"av"`
}

func NewVid[I repo.ID](id I, version uint64) repo.Vid[I] {
	return &vid[I]{ID: id, Version: version}
}

func (v *vid[I]) GetID() I {
	return v.ID
}

func (v *vid[I]) SetID(id I) {
	v.ID = id
}

func (v *vid[I]) GetVersion() uint64 {
	return v.Version
}

func (v *vid[I]) SetVersion(version uint64) {
	v.Version = version
}
