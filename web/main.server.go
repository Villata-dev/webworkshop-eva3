package main

import (
	"context"
	"encoding/json" // Importar fmt si se usa para Printf, etc.
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	// Remover "sync" si mueves sesiones fuera de main
	"time"

	"web-workshop-eval3/web/modules/producto"
	"web-workshop-eval3/web/modules/usuario" // Asegúrate que la ruta es correcta y que incluye la lógica de sesiones

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Constantes para el manejo de sesiones (mantener aquí)
const (
	cookieNombreSesion = "session_id"
	duracionSesion     = 24 * time.Hour
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
	// Inicializar el mux
	mux := http.NewServeMux()

	// Configurar el servidor de archivos estáticos
	fs := http.FileServer(http.Dir("web/public"))
	mux.Handle("/", http.StripPrefix("/", fs))

	// Rutas públicas
	mux.HandleFunc("/api/auth/login", loginHandler)
	mux.HandleFunc("/api/auth/logout", logoutHandler)
	mux.HandleFunc("/api/auth/register", registrarUsuarioHandler)

	// Rutas protegidas
	mux.HandleFunc("/api/v1/productos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			requireAuth(crearProductoHandler)(w, r)
		} else {
			requireAuth(listarProductosHandler)(w, r)
		}
	})

	// Rutas protegidas que requieren rol admin
	mux.HandleFunc("/api/v1/productos/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			requireAuth(requireRole("admin")(eliminarProductoHandler))(w, r)
		} else {
			requireAuth(manejarProducto)(w, r)
		}
	})

	// Inicializar el servidor
	log.Println("🚀 Servidor iniciando en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

// --- Handlers de la API para Productos ---
// Mantener estos handlers. ASEGURARSE de que obtienen el ID usando r.PathValue("id")
// donde aplica (obtenerProductoHandler, actualizarProductoHandler, eliminarProductoHandler).
// Remover la lógica de Method check dentro de estos handlers si se registran con metodo especifico (ej: "GET /...").
// Remover la obtención del ID del contexto en estos handlers si se usa r.PathValue.

func listarProductosHandler(w http.ResponseWriter, r *http.Request) {
	configurarCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Estructura para la respuesta paginada
	type PaginatedResponse struct {
		Items      []producto.Producto `json:"items"`
		TotalItems int                 `json:"totalItems"`
		Page       int                 `json:"page"`
		PerPage    int                 `json:"perPage"`
	}

	// Obtener parámetros de paginación
	page := 1
	perPage := 5

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if perPageStr := r.URL.Query().Get("perPage"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}

	// Calcular índices
	start := (page - 1) * perPage
	end := start + perPage

	// Obtener todos los productos y paginar
	var items []producto.Producto
	totalItems := len(producto.Productos)

	// Ajustar el final si excede el total
	if end > totalItems {
		end = totalItems
	}

	// Extraer slice de productos para la página actual
	keys := make([]string, 0, len(producto.Productos))
	for k := range producto.Productos {
		keys = append(keys, k)
	}
	// Ordenar los keys para que la paginación sea consistente (opcional)
	// sort.Strings(keys)

	for i := start; i < end && i < len(keys); i++ {
		items = append(items, *producto.Productos[keys[i]])
	}

	// Preparar respuesta paginada
	response := PaginatedResponse{
		Items:      items,
		TotalItems: totalItems,
		Page:       page,
		PerPage:    perPage,
	}

	// Asegurarse de encodificar la respuesta correctamente
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error al encodificar respuesta JSON: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	log.Println("✅ listarProductosHandler completado")
}

func obtenerProductoHandler(w http.ResponseWriter, r *http.Request) {
	// configurarCORS(w) // Ya llamada en el main func wrap
	log.Println("📝 Ejecutando obtenerProductoHandler")
	// Remover Method check
	// --- Obtener ID del patrón de la ruta (CORRECCIÓN) ---
	// Ya NO se obtiene del contexto si usas registro con {id}
	// Extraer el ID del producto desde la URL manualmente (Go <1.22)
	// Espera rutas tipo /api/v1/productos/{id}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/productos/"), "/")
	id := pathParts[0]

	if id == "" || id == r.URL.Path || strings.Contains(id, "/") {
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
	// Extraer el ID del producto desde la URL manualmente (Go <1.22)
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/productos/"), "/")
	idProductoAActualizar := pathParts[0]

	if idProductoAActualizar == "" || idProductoAActualizar == r.URL.Path || strings.Contains(idProductoAActualizar, "/") {
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
	// Extraer el ID del producto desde la URL manualmente (Go <1.22)
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/productos/"), "/")
	idProductoAEliminar := pathParts[0]

	if idProductoAEliminar == "" || idProductoAEliminar == r.URL.Path || strings.Contains(idProductoAEliminar, "/") {
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
	configurarCORS(w)
	log.Println("📝 Iniciando proceso de login")

	var credenciales struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credenciales); err != nil {
		log.Printf("❌ Error decodificando credenciales: %v", err)
		http.Error(w, "Error al decodificar credenciales", http.StatusBadRequest)
		return
	}

	log.Printf("👤 Intentando autenticar usuario: %s", credenciales.Username)

	// Usar la función del paquete usuario
	usuarioEncontrado, existe := usuario.ObtenerUsuarioPorNombre(credenciales.Username)
	if !existe {
		log.Printf("❌ Usuario no encontrado: %s", credenciales.Username)
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword(usuarioEncontrado.HashContraseña,
		[]byte(credenciales.Password)); err != nil {
		log.Printf("❌ Contraseña incorrecta para usuario: %s", credenciales.Username)
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	log.Printf("✅ Usuario autenticado exitosamente: %s", credenciales.Username)

	// Crear sesión
	sessionID := uuid.New().String()
	usuario.CrearSesion(sessionID)

	// Establecer cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNombreSesion, // Usar la constante en lugar del string literal
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Enviar respuesta
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"id":       usuarioEncontrado.ID,
		"username": usuarioEncontrado.NombreUsuario,
		"rol":      usuarioEncontrado.Rol,
	}

	log.Printf("📤 Enviando respuesta: %+v", response)
	json.NewEncoder(w).Encode(response)
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

// requireAuth es un middleware que verifica si el usuario está autenticado mediante la cookie de sesión.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Configurar headers CORS para permitir peticiones del frontend
		configurarCORS(w)

		// Si es una petición OPTIONS, retornar inmediatamente
		if r.Method == http.MethodOptions {
			return
		}

		// Buscar la cookie de sesión
		cookie, err := r.Cookie(cookieNombreSesion)
		if err != nil {
			// Si no hay cookie, el usuario no está autenticado
			log.Printf("❌ Error de autenticación: No se encontró la cookie de sesión")
			http.Error(w, "No autorizado", http.StatusUnauthorized)
			return
		}

		// Obtener el usuario asociado a la sesión
		user, ok := obtenerSesion(cookie.Value)
		if !ok {
			// Si la sesión no existe o es inválida
			log.Printf("❌ Error de autenticación: Sesión inválida")
			http.Error(w, "Sesión inválida", http.StatusUnauthorized)
			return
		}

		log.Printf("✅ Usuario autenticado: %s", user.NombreUsuario)
		// Agregar el usuario al contexto de la petición
		ctx := contextWithUsuario(r.Context(), user)
		// Llamar al siguiente handler con el contexto actualizado
		next(w, r.WithContext(ctx))
	}
}

// contextWithUsuario agrega el usuario autenticado al contexto.
func contextWithUsuario(ctx context.Context, user *usuario.Usuario) context.Context {
	return context.WithValue(ctx, ContextKeyUsuarioAutenticado, user)
}

// requireRole es un middleware que verifica si el usuario autenticado tiene el rol requerido.
func requireRole(rol string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(ContextKeyUsuarioAutenticado).(*usuario.Usuario)
			if !ok || user == nil || user.Rol != rol {
				http.Error(w, "No autorizado: se requiere rol "+rol, http.StatusForbidden)
				return
			}
			next(w, r)
		}
	}
}

// configurarCORS agrega las cabeceras necesarias para permitir CORS en las respuestas HTTP.
func configurarCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// Agregar estas funciones que faltan:
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		configurarCORS(w)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	iniciarSesionHandler(w, r)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		configurarCORS(w)
		return
	}
	cerrarSesionHandler(w, r)
}

func manejarProducto(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		obtenerProductoHandler(w, r)
	case http.MethodPut:
		actualizarProductoHandler(w, r)
	case http.MethodDelete:
		eliminarProductoHandler(w, r)
	case http.MethodOptions:
		configurarCORS(w)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func obtenerSesion(sessionID string) (*usuario.Usuario, bool) {
	userID, existe := usuario.ObtenerUsuarioIDPorSesion(sessionID)
	if !existe {
		return nil, false
	}

	user, existe := usuario.ObtenerUsuarioPorID(userID)
	if !existe {
		return nil, false
	}

	return user, true
}
