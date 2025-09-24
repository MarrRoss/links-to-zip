package model

import "github.com/google/uuid"

type ID uuid.UUID

func NewID() ID {
	return ID(uuid.New())
}

func (id ID) ToRaw() uuid.UUID {
	return uuid.UUID(id)
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func UUIDtoID(guid uuid.UUID) ID {
	return ID(guid)
}
