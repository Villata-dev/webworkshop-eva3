package usuario

import (
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// Usuario represents a user in the system
type Usuario struct {
	ID             string
	NombreUsuario  string
	HashContraseña []byte
	Rol            string
}

type Credenciales struct {
	NombreUsuario string `json:"username"`
	Contraseña    string `json:"password"`
}

var (
	Usuarios     = make(map[string]*Usuario)
	Sesiones     = make(map[string]string)
	SesionesLock sync.RWMutex
)

func init() {
	// Crear usuarios predefinidos
	passwordHash1, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	passwordHash2, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	// Agregar usuarios predefinidos
	Usuarios["1"] = &Usuario{
		ID:             "1",
		NombreUsuario:  "user",
		HashContraseña: passwordHash1,
		Rol:            "user",
	}
	Usuarios["2"] = &Usuario{
		ID:             "2",
		NombreUsuario:  "admin",
		HashContraseña: passwordHash2,
		Rol:            "admin",
	}
}

// ObtenerUsuarioIDPorSesion obtiene el ID del usuario asociado a una sesión
func ObtenerUsuarioIDPorSesion(sessionID string) (string, bool) {
	SesionesLock.RLock()
	defer SesionesLock.RUnlock()
	userID, existe := Sesiones[sessionID]
	return userID, existe
}

// ObtenerUsuarioPorID obtiene un usuario por su ID
func ObtenerUsuarioPorID(userID string) (*Usuario, bool) {
	usuario, existe := Usuarios[userID]
	return usuario, existe
}

// ObtenerUsuarioPorNombre obtiene un usuario por su nombre
func ObtenerUsuarioPorNombre(nombreUsuario string) (*Usuario, bool) {
	for _, u := range Usuarios {
		if u.NombreUsuario == nombreUsuario {
			return u, true
		}
	}
	return nil, false
}

// CrearSesion crea una nueva sesión
func CrearSesion(sessionID string, userID string) {
	SesionesLock.Lock()
	defer SesionesLock.Unlock()
	Sesiones[sessionID] = userID
}

// EliminarSesion elimina una sesión existente
func EliminarSesion(sessionID string) {
	SesionesLock.Lock()
	defer SesionesLock.Unlock()
	delete(Sesiones, sessionID)
}

// AgregarUsuario agrega un nuevo usuario al sistema
func AgregarUsuario(user *Usuario) error {
	// Verificar si ya existe un usuario con el mismo nombre
	for _, u := range Usuarios {
		if u.NombreUsuario == user.NombreUsuario {
			return fmt.Errorf("el usuario ya está registrado")
		}
	}
	Usuarios[user.ID] = user
	return nil
}
