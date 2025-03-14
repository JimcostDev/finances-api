package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Income struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Concepto string             `bson:"concepto" json:"concepto"`
	Monto    float64            `bson:"monto" json:"monto"`
}

type Expense struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Concepto string             `bson:"concepto" json:"concepto"`
	Monto    float64            `bson:"monto" json:"monto"`
}

type Report struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID            primitive.ObjectID `bson:"user_id" json:"user_id"`
	Month             string             `bson:"month" json:"month"`
	Year              int                `bson:"year" json:"year"`
	Ingresos          []Income           `bson:"ingresos" json:"ingresos"`
	Gastos            []Expense          `bson:"gastos" json:"gastos"`
	PorcentajeOfrenda float64            `bson:"porcentaje_ofrenda" json:"porcentaje_ofrenda"`
	TotalIngresoBruto float64            `bson:"total_ingreso_bruto" json:"total_ingreso_bruto"`
	Diezmos           float64            `bson:"diezmos" json:"diezmos"`
	Ofrendas          float64            `bson:"ofrendas" json:"ofrendas"`
	Iglesia           float64            `bson:"iglesia" json:"iglesia"`
	IngresosNetos     float64            `bson:"ingresos_netos" json:"ingresos_netos"`
	TotalGastos       float64            `bson:"total_gastos" json:"total_gastos"`
	Liquidacion       float64            `bson:"liquidacion" json:"liquidacion"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}
