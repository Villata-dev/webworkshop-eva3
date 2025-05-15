package producto

import "strconv" // Necesario para convertir int a string para el ID simple

// Producto representa un artículo disponible en el inventario
type Producto struct {
	ID          string  `json:"id"`          // Identificador único del producto
	Nombre      string  `json:"nombre"`      // Nombre del producto
	Descripcion string  `json:"descripcion"` // Descripción breve del producto
	Precio      float64 `json:"precio"`      // Precio del producto
	Stock       int     `json:"stock"`       // Cantidad disponible en inventario
}

// Almacenamiento en memoria para simular la base de datos de productos
var Productos = make(map[string]*Producto) // Mapa para buscar productos por ID
var siguienteID = 1                        // Contador simple para IDs incrementales (mejor usar UUIDs en real)

// Función auxiliar para generar IDs simples y convertir el contador a string
func GenerarSiguienteID() string {
	id := strconv.Itoa(siguienteID) // Convierte el entero a string
	siguienteID++
	return id
}

// Nota: En un proyecto real, estos mapas y la función de ID serían reemplazados
// por una conexión a base de datos y la lógica de persistencia.
