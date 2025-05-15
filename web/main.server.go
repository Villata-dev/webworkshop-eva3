package main

import (
	"context"
	"encoding/json"
	"fmt" // Importar fmt si se usa para Printf, etc.
	"io"
	"log"
	"net/http"
	"strings"

	// Remover "sync" si mueves sesiones fuera de main
	"time"

	"web-workshop-eval3/modules/producto"
	"web-workshop-eval3/modules/usuario" // Asegúrate que la ruta es correcta y que incluye la lógica de sesiones

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Constantes para el manejo de sesiones (mantener aquí)
const (
	cookieNombreSesion = "session_id"
	duracionSesion     = 24 * time.Hour // Duración de la sesión (ej: 24 horas)
)

// Definir tipo y constante para la clave de contexto del usuario autenticado
type contextKey string

const ContextKeyUsuarioAutenticado contextKey = "usuarioAutenticado"

// Remover variables y funciones de sesión si se mueven a modules/usuario
/*
var (
	sesiones     = make(map[string]usuario.Usuario)
	sesionesLock sync.RWMutex
)

func guardarSesion(sessionID string, user usuario.Usuario) { ... }
func obtenerSesion(sessionID string) (usuario.Usuario, bool) { ... }
func eliminarSesion(sessionID string) { ... } // Necesitas esta función para logout!
*/

func main() {
	// Creamos un multiplexer (router) básico de net/http
	mux := http.NewServeMux()

	// --- Rutas Estáticas para el Cliente Web ---
	fs := http.FileServer(http.Dir("./web/public"))
	mux.Handle("/", fs)

	// --- Datos de Ejemplo (Simulación de DB en memoria) ---
	log.Println("⏳ Inicializando datos de ejemplo de productos...")
	// ... (tu código para inicializar productos) ...
	id1 := producto.GenerarSiguienteID()
	producto.Productos[id1] = &producto.Producto{ /* ... */ }
	id2 := producto.GenerarSiguienteID()
	producto.Productos[id2] = &producto.Producto{ /* ... */ }
	id3 := producto.GenerarSiguienteID()
	producto.Productos[id3] = &producto.Producto{ /* ... */ }
	log.Printf("✅ Inicializados %d productos de ejemplo.", len(producto.Productos))

	// --- Usuarios de Ejemplo (Para poder probar login) ---
	// Asegúrate de que modules/usuario/usuario.go tiene las funciones con Mutex
	log.Println("⏳ Registrando usuarios de ejemplo 'admin' y 'user'...")
	hashedPasswordAdmin, errAdmin := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if errAdmin != nil {
		log.Fatalf("Fatal: No se pudo hashear contraseña de admin de ejemplo: %v", errAdmin)
	}
	usuarioAdmin := &usuario.Usuario{
		ID:             uuid.New().String(), // UUID para el admin
		NombreUsuario:  "admin",
		HashContraseña: hashedPasswordAdmin,
		Rol:            "admin", // Rol de administrador
	}
	// Usamos la función segura del paquete usuario para agregar al usuario
	if err := usuario.AgregarUsuario(usuarioAdmin); err != nil {
		log.Printf("⚠️ No se pudo registrar usuario admin de ejemplo: %v (quizás ya existe)", err)
	} else {
		log.Println("✅ Usuario de ejemplo 'admin' registrado.")
	}

	hashedPasswordUser, errUser := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	if errUser != nil {
		log.Fatalf("Fatal: No se pudo hashear contraseña de user de ejemplo: %v", errUser)
	}
	usuarioNormal := &usuario.Usuario{
		ID:             uuid.New().String(), // UUID para el usuario normal
		NombreUsuario:  "user",
		HashContraseña: hashedPasswordUser,
		Rol:            "usuario", // Rol normal
	}
	// Usamos la función segura del paquete usuario para agregar al usuario
	if err := usuario.AgregarUsuario(usuarioNormal); err != nil {
		log.Printf("⚠️ No se pudo registrar usuario user de ejemplo: %v (quizás ya existe)", err)
	} else {
		log.Println("✅ Usuario de ejemplo 'user' registrado.")
	}
	log.Printf("✅ Usuarios de ejemplo registrados.")

	// --- Rutas de la API ---

	// Endpoint para listar todos los productos (GET /api/v1/productos) - NO requiere auth
	// Llama a configurarCORS al inicio del handler
	mux.HandleFunc("GET /api/v1/productos", func(w http.ResponseWriter, r *http.Request) {
		configurarCORS(w) // Configura CORS para esta ruta
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed) // 405
			return
		}
		listarProductosHandler(w, r)
	})

	// Endpoint para crear un nuevo producto (POST /api/v1/productos) - REQUIERE auth (cualquier usuario logueado)
	// Aplica el middleware requireAuth
	mux.HandleFunc("POST /api/v1/productos", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		configurarCORS(w) // Configura CORS antes de la lógica del handler
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed) // 405
			return
		}
		crearProductoHandler(w, r) // Llama al handler real
	}))

	// Endpoint para obtener un producto por ID (GET /api/v1/productos/{id}) - NO requiere auth
	// Usa patrón {id} y llama a configurarCORS
	mux.HandleFunc("GET /api/v1/productos/{id}", func(w http.ResponseWriter, r *http.Request) {
		configurarCORS(w)               // Configura CORS
		if r.Method != http.MethodGet { // Redundante si se registra "GET /...", pero robusto
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed) // 405
			return
		}
		obtenerProductoHandler(w, r) // r.PathValue("id") funcionará aquí
	})

	// Endpoint para actualizar un producto existente (PUT /api/v1/productos/{id}) - REQUIERE auth (cualquier usuario logueado)
	// Usa patrón {id}, aplica middleware requireAuth y llama a configurarCORS
	mux.HandleFunc("PUT /api/v1/productos/{id}", requireAuth(func(w http.ResponseWriter, r *http.Request) {
		configurarCORS(w)               // Configura CORS
		if r.Method != http.MethodPut { // Redundante
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed) // 405
			return
		}
		actualizarProductoHandler(w, r) // r.PathValue("id") funcionará aquí
	}))

	// Endpoint para eliminar un producto (DELETE /api/v1/productos/{id}) - REQUIERE auth Y PERMISO admin
	// Usa patrón {id}, aplica requireAuth Y requireRole("admin") y llama a configurarCORS
	mux.HandleFunc("DELETE /api/v1/productos/{id}", requireAuth(requireRole("admin")(func(w http.ResponseWriter, r *http.Request) {
		configurarCORS(w)                  // Configura CORS
		if r.Method != http.MethodDelete { // Redundante
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed) // 405
			return
		}
		eliminarProductoHandler(w, r) // r.PathValue("id") funcionará aquí
	})))

	// --- Rutas de Autenticación ---
	// IMPORTANTE: Usar el prefijo /api/auth/ según la evaluación
	// Llama a configurarCORS en cada handler de autenticación
	mux.HandleFunc("POST /api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		configurarCORS(w) // Configura CORS
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed) // 405
			return
		}
		registrarUsuarioHandler(w, r)
	})

	mux.HandleFunc("POST /api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		configurarCORS(w) // Configura CORS
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed) // 405
			return
		}
		iniciarSesionHandler(w, r)
	})

	mux.HandleFunc("POST /api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		configurarCORS(w) // Configura CORS
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed) // 405
			return
		}
		cerrarSesionHandler(w, r)
	})

	// NOTA sobre OPTIONS (Preflight CORS): HandleFunc("METHOD /path", handler) de net/http
	// no maneja automáticamente los métodos OPTIONS para preflight.
	// Si el frontend hace peticiones PUT/DELETE/POST con cabeceras o cuerpos complejos
	// a un origen distinto, el navegador enviará un OPTIONS antes.
	// La forma correcta de manejar OPTIONS es tener un middleware de CORS que responda
	// a OPTIONS antes que tus handlers. Como no lo tenemos completo, las cabeceras
	// "Access-Control-Allow-Methods" y "Access-Control-Allow-Headers" en configurarCORS
	// son la forma más simple para que el navegador sepa qué está permitido DESPUÉS
	// de un preflight OPTIONS (asumiendo que el navegador cachea la respuesta OPTIONS
	// o que no se envía un OPTIONS si la petición es simple). Para esta evaluación,
	// puede ser suficiente. Un middleware CORS completo es más robusto.

	// --- Configuración del Servidor ---
	server := &http.Server{
		Addr:         ":8080", // El puerto en el que escuchará el servidor
		Handler:      mux,     // Usamos nuestro multiplexer
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚀 Servidor iniciado en http://localhost%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

// --- Handlers de la API para Productos ---
// Mantener estos handlers. ASEGURARSE de que obtienen el ID usando r.PathValue("id")
// donde aplica (obtenerProductoHandler, actualizarProductoHandler, eliminarProductoHandler).
// Remover la lógica de Method check dentro de estos handlers si se registran con metodo especifico (ej: "GET /...").
// Remover la obtención del ID del contexto en estos handlers si se usa r.PathValue.

func listarProductosHandler(w http.ResponseWriter, r *http.Request) {
	// Referenciar 'r' para evitar error de parámetro no usado
	_ = r.Method
	log.Println("📝 Ejecutando listarProductosHandler")
	// ... (tu lógica existente) ...
	// Remover Method check si se registra como "GET /..."

	productosSlice := []producto.Producto{}
	for _, p := range producto.Productos {
		productosSlice = append(productosSlice, *p)
	}

	// Remover cabeceras CORS si se llaman en configurarCORS al inicio del handler wrap
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(productosSlice); err != nil {
		log.Printf("❌ Error al codificar respuesta JSON: %v", err)
		// No llamar http.Error si cabecera ya escrita
	}
	log.Println("✅ listarProductosHandler completado")
}

func obtenerProductoHandler(w http.ResponseWriter, r *http.Request) {
	// configurarCORS(w) // Ya llamada en el main func wrap
	log.Println("📝 Ejecutando obtenerProductoHandler")
	// Remover Method check
	// --- Obtener ID del patrón de la ruta (CORRECCIÓN) ---
	// Ya NO se obtiene del contexto si usas registro con {id}
	id := r.PathValue("id") // Obtener ID usando r.PathValue

	if id == "" { // Esta verificación sigue siendo útil aunque r.PathValue con {id} patrón no debería dar vacío
		log.Println("❌ ID de producto no proporcionado en la ruta (PathValue vacío)")
		http.Error(w, "ID no proporcionado en la ruta", http.StatusBadRequest) // 400
		return
	}
	log.Printf("Buscando producto con ID: %s", id)

	productoEncontrado, existe := producto.Productos[id]
	if !existe {
		log.Printf("❌ Producto con ID %s no encontrado", id)
		http.Error(w, "Producto no encontrado", http.StatusNotFound) // 404
		return
	}
	log.Printf("✅ Producto encontrado: %s", productoEncontrado.Nombre)

	// Remover cabeceras CORS
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(productoEncontrado); err != nil {
		log.Printf("❌ Error al codificar respuesta JSON para producto %s: %v", id, err)
	}
	log.Println("✅ obtenerProductoHandler completado")
}

func crearProductoHandler(w http.ResponseWriter, r *http.Request) {
	// configurarCORS(w) // Ya llamada en el middleware wrap
	log.Println("📝 Ejecutando crearProductoHandler")
	// Remover Method check

	// Opcional: Obtener usuario del contexto si necesitas saber quién creó el producto
	// user, ok := r.Context().Value(ContextKeyUsuarioAutenticado).(*usuario.Usuario)
	// if ok { log.Printf("Producto siendo creado por usuario: %s", user.NombreUsuario) }

	var nuevoProducto producto.Producto
	lectorLimitado := io.LimitReader(r.Body, 1048576)
	if err := json.NewDecoder(lectorLimitado).Decode(&nuevoProducto); err != nil {
		log.Printf("❌ Error al decodificar cuerpo de la petición: %v", err)
		http.Error(w, "Error al decodificar el producto JSON. Asegúrate de enviar JSON válido.", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	nuevoProducto.ID = "" // Asegurarse de que el servidor asigna el ID

	if strings.TrimSpace(nuevoProducto.Nombre) == "" {
		log.Println("❌ Intento de crear producto con nombre vacío")
		http.Error(w, "El nombre del producto no puede estar vacío", http.StatusBadRequest)
		return
	}
	if nuevoProducto.Precio < 0 {
		log.Println("❌ Intento de crear producto con precio negativo")
		http.Error(w, "El precio del producto no puede ser negativo", http.StatusBadRequest)
		return
	}

	// Usar la función GenerarSiguienteID del paquete producto (asumiendo que está allí)
	idGenerado := producto.GenerarSiguienteID()
	nuevoProducto.ID = idGenerado

	// Usar el mapa global del paquete producto (asumiendo que se accede de forma segura si tiene Mutex, aunque para Productos no añadimos Mutex)
	producto.Productos[nuevoProducto.ID] = &nuevoProducto

	log.Printf("✅ Producto creado con ID: %s", nuevoProducto.ID)

	// Remover cabeceras CORS
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created

	if err := json.NewEncoder(w).Encode(nuevoProducto); err != nil {
		log.Printf("❌ Error al codificar respuesta JSON para nuevo producto: %v", err)
	}
	log.Println("✅ crearProductoHandler completado")
}

func actualizarProductoHandler(w http.ResponseWriter, r *http.Request) {
	// configurarCORS(w) // Ya llamada en el middleware wrap
	log.Println("📝 Ejecutando actualizarProductoHandler")
	// Remover Method check

	// --- Obtener ID del patrón de la ruta (CORRECCIÓN) ---
	// Ya NO se obtiene del contexto si usas registro con {id}
	idProductoAActualizar := r.PathValue("id") // Obtener ID usando r.PathValue

	if idProductoAActualizar == "" { // Verificación necesaria
		log.Println("❌ ID de producto a actualizar no proporcionado en la ruta (PathValue vacío)")
		http.Error(w, "ID no proporcionado en la ruta", http.StatusBadRequest) // 400
		return
	}
	log.Printf("Intentando actualizar producto con ID: %s", idProductoAActualizar)

	// Opcional: Obtener usuario del contexto para posibles cheques de permiso adicionales (ej: solo el creador puede actualizar)
	// user, ok := r.Context().Value(ContextKeyUsuarioAutenticado).(*usuario.Usuario)
	// if ok { log.Printf("Actualización solicitada por usuario: %s", user.NombreUsuario) }

	var datosActualizados producto.Producto
	lectorLimitado := io.LimitReader(r.Body, 1048576)
	if err := json.NewDecoder(lectorLimitado).Decode(&datosActualizados); err != nil {
		log.Printf("❌ Error al decodificar cuerpo de la petición para actualizar: %v", err)
		http.Error(w, "Error al decodificar datos de actualización. Asegúrate de enviar JSON válido.", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Buscar el producto existente por el ID de la RUTA
	productoExistente, existe := producto.Productos[idProductoAActualizar]
	if !existe {
		log.Printf("❌ Producto con ID %s no encontrado para actualizar", idProductoAActualizar)
		http.Error(w, "Producto no encontrado para actualizar", http.StatusNotFound) // 404
		return
	}
	log.Printf("✅ Producto existente encontrado para actualizar: %s", productoExistente.Nombre)

	// Opcional: Validaciones de los datos actualizados (mantener si las tenías)
	if strings.TrimSpace(datosActualizados.Nombre) == "" {
		log.Println("❌ Intento de actualizar producto con nombre vacío")
		http.Error(w, "El nombre del producto no puede estar vacío", http.StatusBadRequest)
		return
	}
	if datosActualizados.Precio < 0 {
		log.Println("❌ Intento de actualizar producto con precio negativo")
		http.Error(w, "El precio del producto no puede ser negativo", http.StatusBadRequest)
		return
	}
	// Puedes copiar campos uno por uno si no quieres reemplazar todo el struct
	// productoExistente.Nombre = datosActualizados.Nombre
	// productoExistente.Descripcion = datosActualizados.Descripcion
	// productoExistente.Precio = datosActualizados.Precio
	// productoExistente.Stock = datosActualizados.Stock

	// Opción 1 (Simple): Reemplazar el producto existente con los datos decodificados
	// Aseguramos que el ID en el struct actualizado sea el de la ruta
	datosActualizados.ID = idProductoAActualizar
	producto.Productos[idProductoAActualizar] = &datosActualizados // Reemplazamos el puntero en el mapa

	log.Printf("✅ Producto con ID %s actualizado exitosamente.", idProductoAActualizar)

	// Remover cabeceras CORS
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK

	if err := json.NewEncoder(w).Encode(producto.Productos[idProductoAActualizar]); err != nil { // Enviamos el producto *desde* el mapa
		log.Printf("❌ Error al codificar respuesta JSON para producto actualizado %s: %v", idProductoAActualizar, err)
	}
	log.Println("✅ actualizarProductoHandler completado")
}

func eliminarProductoHandler(w http.ResponseWriter, r *http.Request) {
	// configurarCORS(w) // Ya llamada en el middleware wrap
	log.Println("📝 Ejecutando eliminarProductoHandler")
	// Remover Method check

	// --- Obtener ID del patrón de la ruta (CORRECCIÓN) ---
	// Ya NO se obtiene del contexto si usas registro con {id}
	idProductoAEliminar := r.PathValue("id") // Obtener ID usando r.PathValue

	if idProductoAEliminar == "" { // Verificación necesaria
		log.Println("❌ ID de producto a eliminar no proporcionado en la ruta (PathValue vacío)")
		http.Error(w, "ID no proporcionado en la ruta", http.StatusBadRequest) // 400
		return
	}
	log.Printf("Intentando eliminar producto con ID: %s", idProductoAEliminar)

	// --- Obtener usuario del contexto (¡YA LO HACE requireAuth y requireRole!) ---
	// user, ok := r.Context().Value(ContextKeyUsuarioAutenticado).(*usuario.Usuario)
	// if !ok || user == nil { /* Error, no deberia pasar */ }
	// log.Printf("Eliminación solicitada por usuario: '%s' con Rol '%s'", user.NombreUsuario, user.Rol)

	// --- ¡¡El chequeo de rol "admin" YA LO HACE el middleware requireRole!! ---
	// No necesitas repetir el chequeo de rol aquí dentro del handler
	log.Println("✅ Permiso (rol admin) verificado por middleware.")

	// Verificar si el producto existe antes de intentar eliminar
	if _, existe := producto.Productos[idProductoAEliminar]; !existe {
		log.Printf("❌ Producto con ID %s no encontrado para eliminar", idProductoAEliminar)
		http.Error(w, "Producto no encontrado para eliminar", http.StatusNotFound) // 404
		return
	}
	log.Printf("✅ Producto encontrado para eliminar: %s", idProductoAEliminar)

	// Eliminar el producto del mapa (simulando eliminar de DB)
	delete(producto.Productos, idProductoAEliminar)

	log.Printf("✅ Producto con ID %s eliminado exitosamente.", idProductoAEliminar)

	// Remover cabeceras CORS
	// La respuesta 204 No Content no tiene cuerpo.
	w.WriteHeader(http.StatusNoContent) // 204 No Content

	log.Println("✅ eliminarProductoHandler completado (204 No Content)")
}

// --- Handlers de Autenticación ---
// Mantener estos handlers, ajustar CORS y la respuesta de logout sin cookie.
// Asegurarse que llaman a las funciones seguras de usuario (AgregarUsuario)
// y sesiones (CrearSesion, EliminarSesion, ObtenerUsuarioIDPorSesion) que estarán en modules/usuario.

func registrarUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	// configurarCORS(w) // Ya llamada en el main func wrap
	log.Println("📝 Ejecutando registrarUsuarioHandler")
	// Remover Method check

	var credenciales usuario.Credenciales
	lectorLimitado := io.LimitReader(r.Body, 1048576)
	if err := json.NewDecoder(lectorLimitado).Decode(&credenciales); err != nil {
		log.Printf("❌ Error al decodificar JSON de registro: %v", err)
		http.Error(w, "Error al decodificar credenciales. Asegúrate de enviar JSON válido con 'username' y 'password'.", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(credenciales.NombreUsuario) == "" || strings.TrimSpace(credenciales.Contraseña) == "" {
		log.Println("❌ Intento de registro con usuario o contraseña vacíos.")
		http.Error(w, "Nombre de usuario y contraseña no pueden estar vacíos.", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credenciales.Contraseña), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ Error al hashear contraseña: %v", err)
		http.Error(w, "Error interno del servidor al procesar contraseña.", http.StatusInternalServerError)
		return
	}

	nuevoUsuario := &usuario.Usuario{
		ID:             uuid.New().String(),
		NombreUsuario:  credenciales.NombreUsuario,
		HashContraseña: hashedPassword,
		Rol:            "usuario", // Rol por defecto
	}

	// Usar la función del paquete usuario
	if err := usuario.AgregarUsuario(nuevoUsuario); err != nil {
		log.Printf("❌ Error al agregar usuario '%s': %v", credenciales.NombreUsuario, err)
		if strings.Contains(err.Error(), "ya está registrado") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Error interno del servidor al guardar usuario.", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("✅ Usuario registrado exitosamente: '%s' con ID %s", nuevoUsuario.NombreUsuario, nuevoUsuario.ID)

	// Respuesta de éxito (sin el hash ni session_id)
	respuestaExito := map[string]string{
		"message":  "Usuario registrado exitosamente",
		"id":       nuevoUsuario.ID,
		"username": nuevoUsuario.NombreUsuario,
	}

	// Remover cabeceras CORS
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created

	if err := json.NewEncoder(w).Encode(respuestaExito); err != nil {
		log.Printf("❌ Error al codificar respuesta JSON de registro: %v", err)
	}
	log.Println("✅ registrarUsuarioHandler completado")
}

func iniciarSesionHandler(w http.ResponseWriter, r *http.Request) {
	// configurarCORS(w) // Ya llamada en el main func wrap
	log.Println("📝 Ejecutando iniciarSesionHandler")
	// Remover Method check

	var credenciales usuario.Credenciales
	lectorLimitado := io.LimitReader(r.Body, 1048576)
	if err := json.NewDecoder(lectorLimitado).Decode(&credenciales); err != nil {
		log.Printf("❌ Error al decodificar JSON de inicio de sesión: %v", err)
		http.Error(w, "Error al decodificar credenciales. Asegúrate de enviar JSON válido con 'username' y 'password'.", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if strings.TrimSpace(credenciales.NombreUsuario) == "" || strings.TrimSpace(credenciales.Contraseña) == "" {
		log.Println("❌ Intento de inicio de sesión con usuario o contraseña vacíos.")
		http.Error(w, "Nombre de usuario y contraseña no pueden estar vacíos.", http.StatusBadRequest)
		return
	}

	// Usar la función del paquete usuario
	usuarioEncontrado, existe := usuario.ObtenerUsuarioPorNombre(credenciales.NombreUsuario)
	if !existe {
		log.Printf("❌ Intento de login fallido: usuario '%s' no encontrado.", credenciales.NombreUsuario)
		http.Error(w, "Credenciales inválidas.", http.StatusUnauthorized) // 401
		return
	}

	err := bcrypt.CompareHashAndPassword(usuarioEncontrado.HashContraseña, []byte(credenciales.Contraseña))
	if err != nil {
		log.Printf("❌ Intento de login fallido para usuario '%s': contraseña incorrecta.", credenciales.NombreUsuario)
		http.Error(w, "Credenciales inválidas.", http.StatusUnauthorized) // 401
		return
	}

	log.Printf("✅ Autenticación exitosa para usuario: '%s'", usuarioEncontrado.NombreUsuario)

	// --- Autenticación Exitosa: Crear Sesión y Setear Cookie ---

	// Usar la función del paquete usuario
	sessionID, err := usuario.CrearSesion(usuarioEncontrado.ID) // Asociar el usuario ID con un nuevo ID de sesión
	if err != nil {
		log.Printf("❌ Error al crear sesión para usuario '%s': %v", usuarioEncontrado.NombreUsuario, err)
		http.Error(w, "Error interno del servidor al crear sesión.", http.StatusInternalServerError) // 500
		return
	}

	expires := time.Now().Add(duracionSesion)
	cookie := http.Cookie{
		Name:     cookieNombreSesion,
		Value:    sessionID,
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,                // DEBE ser 'true' en producción con HTTPS
		SameSite: http.SameSiteLaxMode, // O StrictMode
	}
	http.SetCookie(w, &cookie) // Setear la cookie en la respuesta

	log.Printf("✅ Cookie de sesión seteada para usuario '%s'", usuarioEncontrado.NombreUsuario)

	// --- Respuesta Exitosa ---
	respuestaExito := map[string]interface{}{ // Usar interface{} si mezclas tipos
		"message":  fmt.Sprintf("Inicio de sesión exitoso para %s", usuarioEncontrado.NombreUsuario),
		"username": usuarioEncontrado.NombreUsuario,
		"id":       usuarioEncontrado.ID,  // Puedes incluir el ID
		"rol":      usuarioEncontrado.Rol, // Puedes incluir el rol
		// ¡¡IMPORTANTE!! NO INCLUIR "session_id"
	}

	// Remover cabeceras CORS

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK

	if err := json.NewEncoder(w).Encode(respuestaExito); err != nil {
		log.Printf("❌ Error al codificar respuesta JSON de login: %v", err)
	}
	log.Println("✅ iniciarSesionHandler completado")
}

func cerrarSesionHandler(w http.ResponseWriter, r *http.Request) {
	// configurarCORS(w) // Ya llamada en el main func wrap
	log.Println("📝 Ejecutando cerrarSesionHandler")
	// Remover Method check

	// 1. Intentar obtener la cookie de sesión
	cookie, err := r.Cookie(cookieNombreSesion)
	if err != nil {
		// Si no hay cookie, el usuario no estaba logueado.
		if err == http.ErrNoCookie {
			log.Println("ℹ️ Intento de logout sin cookie de sesión activa.")
			// Responder 204 No Content
			w.WriteHeader(http.StatusNoContent) // 204 No Content
			log.Println("✅ cerrarSesionHandler completado (no había sesión activa)")
			return
		}
		// Otro error al obtener la cookie (raro)
		log.Printf("❌ Error inesperado al obtener cookie de sesión: %v", err)
		http.Error(w, "Error interno al procesar la petición.", http.StatusInternalServerError) // 500
		return
	}

	// 2. Si la cookie existe, obtener el ID de sesión y eliminar la sesión del almacenamiento en memoria
	sessionID := cookie.Value
	// Usar la función del paquete usuario
	usuario.EliminarSesion(sessionID) // Esta función ya maneja el Mutex y el log

	// 3. Invalidar la cookie en el navegador del cliente
	expiredCookie := http.Cookie{
		Name:     cookieNombreSesion,
		Value:    "",
		Expires:  time.Now().Add(-24 * time.Hour), // Un día en el pasado
		Path:     "/",                             // Mismo Path que la cookie original
		HttpOnly: true,
		Secure:   false,                // DEBE ser 'true' en producción con HTTPS
		SameSite: http.SameSiteLaxMode, // O StrictMode
	}
	http.SetCookie(w, &expiredCookie) // Setear la cookie expirada

	log.Printf("✅ Sesión y cookie eliminadas para ID de sesión: %s", sessionID)

	// --- Respuesta Exitosa de Logout ---
	// Usar 204 No Content es estándar
	// Remover cabeceras CORS
	w.WriteHeader(http.StatusNoContent) // 204 No Content

	// NO llamar a json.NewEncoder(w).Encode() para 204
	log.Println("✅ cerrarSesionHandler completado (204 No Content)")
}

// --- Middleware de Autenticación ---

// requireAuth es un middleware que verifica si el usuario está autenticado (tiene una sesión válida).
// Si es válido, añade el usuario al contexto y llama al siguiente handler. Responde 401 si no es válido.
// Este middleware DEBE llamarse ANTES de cualquier middleware de permisos (como requireRole).
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔒 Middleware requireAuth para %s %s", r.Method, r.URL.Path)

		// CORS headers para respuestas 401 (aunque mejor un middleware CORS general)
		configurarCORS(w) // Configura CORS incluso para 401

		// 1. Intentar obtener la cookie de sesión
		cookie, err := r.Cookie(cookieNombreSesion)
		if err != nil {
			if err == http.ErrNoCookie {
				log.Println("❌ requireAuth: Cookie de sesión no encontrada.")
				http.Error(w, "Autenticación requerida. No se encontró cookie de sesión.", http.StatusUnauthorized) // 401
				return
			}
			// Otro error al obtener la cookie (raro)
			log.Printf("❌ requireAuth: Error inesperado al obtener cookie de sesión: %v", err)
			http.Error(w, "Error interno al procesar autenticación.", http.StatusInternalServerError) // 500
			return
		}

		// 2. Obtener el ID de sesión de la cookie
		sessionID := cookie.Value
		if sessionID == "" {
			log.Println("❌ requireAuth: Cookie de sesión encontrada, pero el valor (ID de sesión) está vacío.")
			http.Error(w, "Sesión inválida.", http.StatusUnauthorized) // 401
			// Opcional: Invalidar la cookie aquí
			return
		}
		log.Printf("requireAuth: Cookie encontrada con ID de sesión: %s", sessionID)

		// 3. --- ¡¡VALIDAR EL ID DE SESIÓN CONTRA EL ALMACENAMIENTO EN MEMORIA!! ---
		// Usar la función del paquete usuario para buscar el ID de usuario asociado al ID de sesión
		// Tu lógica actual ObtenerSesion(cookie.Value) que devuelve usuario.Usuario y bool funciona si mueves sesiones a usuario.
		// Si mantienes sesiones en main, usa obtenerSesion de main.
		// Asumiendo que la lógica de sesiones está en modules/usuario:
		userID, existe := usuario.ObtenerUsuarioIDPorSesion(sessionID) // Función del paquete usuario
		if !existe {
			log.Printf("❌ requireAuth: ID de sesión '%s' no encontrado o expirado en el almacenamiento.", sessionID)
			// Opcional: Invalidar la cookie en el cliente
			setExpiredCookie(w)                                                                                      // Necesitarías implementar setExpiredCookie
			http.Error(w, "Sesión expirada o inválida. Por favor, inicia sesión de nuevo.", http.StatusUnauthorized) // 401
			return
		}
		log.Printf("✅ requireAuth: Sesión ID %s válida para usuario ID %s", sessionID, userID)

		// 4. --- OBTENER EL STRUCT COMPLETO DEL USUARIO AUTENTICADO ---
		// Ahora que sabemos el ID del usuario, obtenemos su struct completo usando la función del paquete usuario.
		user, existe := usuario.ObtenerUsuarioPorID(userID) // Función del paquete usuario
		if !existe {
			log.Printf("❌ requireAuth: Error interno. Usuario ID %s encontrado en sesiones, pero no en almacenamiento de usuarios.", userID)
			// Opcional: Eliminar la sesión ya que apunta a un usuario inexistente
			usuario.EliminarSesion(sessionID)                                                // Función del paquete usuario
			setExpiredCookie(w)                                                              // Necesitarías implementar setExpiredCookie
			http.Error(w, "Error interno de autenticación.", http.StatusInternalServerError) // 500
			return
		}
		log.Printf("requireAuth: Usuario autenticado: '%s' con Rol '%s'", user.NombreUsuario, user.Rol)

		// 5. --- PASAR EL USUARIO AUTENTICADO AL SIGUIENTE HANDLER USANDO CONTEXTO ---
		// Creamos un nuevo contexto con el usuario autenticado bajo la clave ContextKeyUsuarioAutenticado
		ctx := context.WithValue(r.Context(), ContextKeyUsuarioAutenticado, user)
		// Creamos una NUEVA petición con este contexto modificado
		rWithContext := r.WithContext(ctx)

		// 6. Llamar al siguiente handler (o al próximo middleware) con la petición que contiene el contexto
		log.Println("requireAuth: Autenticación exitosa. Pasando a siguiente handler.")
		next.ServeHTTP(w, rWithContext) // ¡Importante usar rWithContext!
	}
}

// Middleware de Permisos
// requireRole es una fábrica de middleware que verifica si el usuario autenticado tiene uno de los roles permitidos.
// DEBE usarse DESPUÉS de requireAuth, ya que asume que el usuario está en el contexto.
func requireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		// La función interna es el middleware real
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("🔑 Middleware requireRole para %s %s (Roles requeridos: %v)", r.Method, r.URL.Path, roles)

			// El middleware requireAuth ya puso el usuario en el contexto si la sesión era válida.
			// Lo obtenemos del contexto.
			user, ok := r.Context().Value(ContextKeyUsuarioAutenticado).(*usuario.Usuario)
			if !ok || user == nil {
				// Esto NO DEBERÍA PASAR si requireAuth se ejecutó justo antes.
				// Si pasa, indica un error en la cadena de middlewares o en requireAuth.
				log.Println("❌ requireRole: Error interno. Usuario autenticado no encontrado en contexto.")
				http.Error(w, "Error de autorización interna.", http.StatusInternalServerError) // 500
				return
			}

			// Verificar si el rol del usuario está en la lista de roles permitidos
			allowed := false
			for _, role := range roles {
				if user.Rol == role {
					allowed = true
					break // Rol encontrado, podemos salir del bucle
				}
			}

			if !allowed {
				log.Printf("❌ requireRole: Acceso denegado para usuario '%s' (Rol: %s). Se requiere uno de los roles: %v.", user.NombreUsuario, user.Rol, roles)
				http.Error(w, fmt.Sprintf("Permiso denegado. Se requiere uno de los siguientes roles: %v", roles), http.StatusForbidden) // 403 Forbidden
				return
			}

			// Si el rol es permitido, llamar al siguiente handler en la cadena
			log.Printf("✅ requireRole: Permiso concedido para usuario '%s' (Rol: %s).", user.NombreUsuario, user.Rol)
			next.ServeHTTP(w, r) // Pasa la petición (que aún tiene el usuario en contexto) al siguiente
		}
	}
}

// Función auxiliar para setear una cookie expirada (útil en middleware y logout)
func setExpiredCookie(w http.ResponseWriter) {
	expiredCookie := http.Cookie{
		Name:     cookieNombreSesion,
		Value:    "",
		Expires:  time.Now().Add(-24 * time.Hour), // Un día en el pasado
		Path:     "/",
		HttpOnly: true,
		Secure:   false,                // AJUSTAR para producción con HTTPS
		SameSite: http.SameSiteLaxMode, // Ajustar si usaste StrictMode
	}
	http.SetCookie(w, &expiredCookie)
}

// Función auxiliar para configurar cabeceras CORS (Llamar al inicio de CADA handler API)
func configurarCORS(w http.ResponseWriter) {
	// !!! IMPORTANTE: Reemplaza "http://localhost:5500" con el ORIGEN REAL de tu frontend
	// Si usas VS Code Live Server, suele ser 5500. Si es otro puerto, ajusta.
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
	w.Header().Set("Access-Control-Allow-Credentials", "true")                        // Necesario para que el navegador envíe cookies (sesión)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Métodos permitidos para el frontend
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // Cabeceras permitidas en la petición (Authorization por si usas tokens después, Content-Type para JSON)

	// Handle preflight OPTIONS requests - CORS middleware real haría esto de otra forma
	// Para esta evaluación, puedes añadir una verificación básica si OPTIONS es un problema.
	// if r.Method == http.MethodOptions {
	//      w.WriteHeader(http.StatusOK)
	//      return // Terminar la petición OPTIONS aquí
	// }
	// Esta verificación del método OPTIONS debe ir en el envoltorio de cada handler,
	// o idealmente en un middleware CORS dedicado que envuelva todo el mux API.
	// Por ahora, confiar en que configurarCORS setea las cabeceras correctas para la respuesta
	// y que el navegador las respeta puede ser suficiente para la evaluación.
}
