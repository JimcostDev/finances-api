package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Category struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Nombre string             `bson:"nombre" json:"nombre"`
	Tipo   string             `bson:"tipo" json:"tipo"` // "ingreso" | "gasto"
}

