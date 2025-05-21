package producto

import "fmt"

type Producto struct {
	ID          string  `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Precio      float64 `json:"precio"`
	Stock       int     `json:"stock"`
}

var Productos = make(map[string]*Producto)

func GenerarSiguienteID() string {
	return fmt.Sprintf("%d", len(Productos)+1)
}
