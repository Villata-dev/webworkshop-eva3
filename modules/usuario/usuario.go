package usuario

import (
	"fmt"
	"log" // Importar para logs dentro del módulo
	"sync"

	// No necesitas time aquí, solo en el handler para la cookie Expires
	// No necesitas bcrypt aquí
	"github.com/google/uuid" // Importar para generar UUIDs de sesión
)

// Usuario representa un usuario registrado en el sistema
type Usuario struct {
	ID             string `json:"id"`       // Identificador único del usuario (UUIDs)
	NombreUsuario  string `json:"username"` // Nombre de usuario para login
	HashContraseña []byte `json:"-"`        // Hash seguro de la contraseña. El tag `json:"-"` evita que se exponga.
	Rol            string `json:"role"`     // Rol del usuario (ej: "usuario", "admin"). Por defecto será "usuario".
}

// Credenciales es una struct temporal para decodificar las credenciales de registro/login
type Credenciales struct {
	NombreUsuario string `json:"username"`
	Contraseña    string `json:"password"`
}

// --- Almacenamiento en memoria con Mutex para Usuarios ---
var (
	muUsuario         sync.Mutex                  // Mutex para proteger el acceso a los mapas de usuarios
	Usuarios          = make(map[string]*Usuario) // Mapa [ID de Usuario] -> *Usuario
	UsuariosPorNombre = make(map[string]*Usuario) // Mapa [NombreUsuario] -> *Usuario
)

// --- Almacenamiento en memoria con Mutex para Sesiones ---
var (
	muSesion sync.Mutex // Mutex para proteger el acceso al mapa de sesiones
	// El mapa de Sesiones asocia el ID de Sesión con el ID del Usuario autenticado
	Sesiones = make(map[string]string) // Mapa [ID de Sesión] -> ID de Usuario
	// Podríamos almacenar más datos de sesión si fuera necesario (ej: tiempo de creación para expiración real en el server)
)

// --- Funciones seguras para gestionar Usuarios (Asegúrate de tener estas con los mutexes) ---

// AgregarUsuario agrega un usuario al almacenamiento de forma segura
func AgregarUsuario(user *Usuario) error {
	muUsuario.Lock()         // Bloqueamos el mutex antes de acceder a los mapas
	defer muUsuario.Unlock() // Aseguramos que se desbloquee al salir de la función

	log.Printf("🔒 Intentando agregar usuario '%s'", user.NombreUsuario)

	// Verificar si el nombre de usuario ya existe
	if _, existe := UsuariosPorNombre[user.NombreUsuario]; existe {
		log.Printf("❌ Usuario '%s' ya está registrado.", user.NombreUsuario)
		return fmt.Errorf("El nombre de usuario '%s' ya está registrado", user.NombreUsuario)
	}

	// Asignar un ID único si no lo tiene (debería generarse antes de llamar a esta función)
	if user.ID == "" {
		user.ID = uuid.New().String() // Generar UUID si no tiene
	} else if _, existe := Usuarios[user.ID]; existe {
		// Si ya tiene ID pero ese ID ya existe en el mapa, es un error inesperado
		log.Printf("❌ Error interno: El ID de usuario '%s' ya existe al agregar.", user.ID)
		return fmt.Errorf("Error interno: El ID de usuario '%s' ya existe", user.ID)
	}

	Usuarios[user.ID] = user                     // Guardar por ID
	UsuariosPorNombre[user.NombreUsuario] = user // Guardar por NombreUsuario

	log.Printf("✅ Usuario '%s' agregado con ID %s", user.NombreUsuario, user.ID)
	return nil
}

// ObtenerUsuarioPorNombre busca un usuario por nombre de forma segura
func ObtenerUsuarioPorNombre(nombreUsuario string) (*Usuario, bool) {
	muUsuario.Lock()         // Bloqueamos el mutex antes de acceder al mapa
	defer muUsuario.Unlock() // Aseguramos que se desbloquee al salir

	log.Printf("🔒 Buscando usuario por nombre: '%s'", nombreUsuario)
	user, existe := UsuariosPorNombre[nombreUsuario]
	if existe {
		log.Printf("✅ Usuario '%s' encontrado.", nombreUsuario)
	} else {
		log.Printf("❌ Usuario '%s' no encontrado.", nombreUsuario)
	}
	return user, existe
}

// ObtenerUsuarioPorID busca un usuario por ID de forma segura
func ObtenerUsuarioPorID(id string) (*Usuario, bool) {
	muUsuario.Lock()         // Bloqueamos el mutex
	defer muUsuario.Unlock() // Desbloqueamos

	log.Printf("🔒 Buscando usuario por ID: '%s'", id)
	user, existe := Usuarios[id]
	if existe {
		log.Printf("✅ Usuario con ID '%s' encontrado.", id)
	} else {
		log.Printf("❌ Usuario con ID '%s' no encontrado.", id)
	}
	return user, existe
}

// --- Funciones seguras para gestionar Sesiones ---

// CrearSesion genera un ID de sesión único y lo asocia con un ID de usuario
func CrearSesion(userID string) (string, error) {
	muSesion.Lock()         // Bloqueamos el mutex de sesiones
	defer muSesion.Unlock() // Aseguramos que se desbloquee

	log.Printf("🔒 Intentando crear sesión para usuario ID '%s'", userID)

	sessionID := uuid.New().String() // Generar un UUID para la sesión

	// Opcional: Verificar si el ID de sesión ya existe (muy improbable con UUID)
	if _, existe := Sesiones[sessionID]; existe {
		log.Printf("❌ Error interno: ID de sesión '%s' generado ya existe.", sessionID)
		return "", fmt.Errorf("Error interno: ID de sesión generado ya existe")
	}

	Sesiones[sessionID] = userID // Guardar la asociación ID de Sesión -> ID de Usuario
	log.Printf("✅ Sesión creada: '%s' para usuario ID '%s'", sessionID, userID)
	return sessionID, nil
}

// ObtenerUsuarioIDPorSesion busca el ID de usuario asociado a un ID de sesión
func ObtenerUsuarioIDPorSesion(sessionID string) (string, bool) {
	muSesion.Lock()         // Bloqueamos el mutex de sesiones
	defer muSesion.Unlock() // Aseguramos que se desbloquee

	log.Printf("🔒 Buscando usuario ID por sesión: '%s'", sessionID)
	userID, existe := Sesiones[sessionID]
	if existe {
		log.Printf("✅ Sesión '%s' encontrada, usuario ID '%s'.", sessionID, userID)
	} else {
		log.Printf("❌ Sesión '%s' no encontrada.", sessionID)
	}
	return userID, existe
}

// EliminarSesion elimina la asociación de una sesión
func EliminarSesion(sessionID string) {
	muSesion.Lock()         // Bloqueamos el mutex de sesiones
	defer muSesion.Unlock() // Aseguramos que se desbloquee

	log.Printf("🔒 Intentando eliminar sesión: '%s'", sessionID)
	// No hay error si el ID de sesión no existe
	delete(Sesiones, sessionID)
	log.Printf("✅ Sesión eliminada: '%s'", sessionID)
}

// TODO: Considerar la limpieza de sesiones expiradas si almacenamos tiempo de creación.
