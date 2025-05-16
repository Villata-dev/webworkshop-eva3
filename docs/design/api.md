# Evaluación 3 - API RESTful Completa y Cliente Web

## Descripción General

Este proyecto consiste en la implementación de una API RESTful utilizando Go con el paquete estándar `net/http` y un cliente web básico desarrollado con Web Components (HTML, CSS, JavaScript) para consumir dicha API.

La API gestiona una colección de **Productos** (CRUD) y cuenta con un sistema básico de autenticación (registro, login, logout) basado en cookies de sesión y autorización mediante roles (requiriendo autenticación para la mayoría de las operaciones CRUD y el rol 'admin' para la eliminación de productos). El almacenamiento de datos (productos y sesiones) se realiza en memoria para propósitos de esta evaluación.

El cliente web permite a los usuarios interactuar con la API para registrarse, iniciar/cerrar sesión, visualizar la lista de productos, y (dependiendo de sus permisos) crear, editar y eliminar productos a través de una interfaz gráfica sencilla.

## Endpoints CRUD

Aquí se detallan los endpoints de la API RESTful para la gestión de la entidad Producto (`/api/v1/productos`).

| Método | Ruta                       | Descripción                                     | Parámetros (URL/Path) | Cuerpo Petición (Body)                                  | Ejemplo Petición (curl)                                                                                               | Respuesta Éxito (Body)                                  | Errores Posibles (Códigos HTTP)                               |
|--------|----------------------------|-------------------------------------------------|-----------------------|---------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------|---------------------------------------------------------------|
| GET    | `/api/v1/productos`        | Obtiene una lista de todos los productos.       | Ninguno               | Ninguno                                                 | `curl http://localhost:8080/api/v1/productos`                                                                           | `[ { "id": "...", "nombre": "...", ... }, ... ]` (JSON Array de Producto) | 405 Method Not Allowed                                        |
| POST   | `/api/v1/productos`        | Crea un nuevo producto. **Requiere Auth.** | Ninguno               | Objeto Producto (JSON): `{ "nombre": "string", "descripcion": "string", "precio": float64, "stock": int }` | `curl -X POST -H "Content-Type: application/json" -d '{"nombre": "Ejemplo", "precio": 100}' http://localhost:8080/api/v1/productos` | Objeto Producto creado (JSON)                           | 400 Bad Request, 401 Unauthorized, 405 Method Not Allowed, 500 Internal Server Error |
| GET    | `/api/v1/productos/{id}`   | Obtiene un producto específico por su ID.       | `id` (string)         | Ninguno                                                 | `curl http://localhost:8080/api/v1/productos/123`                                                                       | Objeto Producto (JSON): `{ "id": "...", "nombre": "...", ... } ` | 404 Not Found, 405 Method Not Allowed                         |
| PUT    | `/api/v1/productos/{id}`   | Actualiza un producto específico por su ID. **Requiere Auth.** | `id` (string)         | Objeto Producto (JSON): `{ "id": "string", "nombre": "...", ... }` (el `id` en body es opcional, se usa el de la ruta) | `curl -X PUT -H "Content-Type: application/json" -d '{"nombre": "Actualizado", "precio": 200}' http://localhost:8080/api/v1/productos/123` | Objeto Producto actualizado (JSON)                        | 400 Bad Request, 401 Unauthorized, 404 Not Found, 405 Method Not Allowed, 500 Internal Server Error |
| DELETE | `/api/v1/productos/{id}` | Elimina un producto específico por su ID. **Requiere Auth (Rol Admin).** | `id` (string)         | Ninguno                                                 | `curl -X DELETE http://localhost:8080/api/v1/productos/123`                                                             | Respuesta vacía (Status 204 No Content)                 | 401 Unauthorized, 403 Forbidden, 404 Not Found, 405 Method Not Allowed, 500 Internal Server Error |

---

## Endpoints de Autenticación

Endpoints para el registro, inicio y cierre de sesión de usuarios (`/api/auth/`).

| Método | Ruta                     | Descripción                                        | Cuerpo Petición (Body)                    | Cookies (Envía/Recibe)                                  | Ejemplo Petición (curl)                                                                                                                               | Respuesta Éxito (Body)                                      | Errores Posibles (Códigos HTTP)                      |
|--------|--------------------------|----------------------------------------------------|-------------------------------------------|---------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------|------------------------------------------------------|
| POST   | `/api/auth/register`     | Registra un nuevo usuario en el sistema.           | `{ "username": "string", "password": "string" }` | Ninguna                                                 | `curl -v -H "Content-Type: application/json" -d '{"username": "nuevo", "password": "pass"}' http://localhost:8080/api/auth/register`                     | `{ "message": "...", "id": "...", "username": "..." }`    | 400 Bad Request, 409 Conflict, 500 Internal Server Error |
| POST   | `/api/auth/login`        | Inicia sesión para un usuario existente.           | `{ "username": "string", "password": "string" }` | Recibe `Set-Cookie: session_id=...`                     | `curl -v -c cookiejar.txt -H "Content-Type: application/json" -d '{"username": "user", "password": "user123"}' http://localhost:8080/api/auth/login`        | `{ "message": "...", "username": "...", "id": "...", "rol": "..." }` | 400 Bad Request, 401 Unauthorized, 500 Internal Server Error |
| POST   | `/api/auth/logout`       | Cierra la sesión activa del usuario actual.        | Ninguno                                   | Envía `Cookie: session_id=...`, Recibe `Set-Cookie: session_id=...; Expires=(past)` | `curl -v -b cookiejar.txt -X POST http://localhost:8080/api/auth/logout`                                                                                | Respuesta vacía (Status 204 No Content)                   | 204 No Content (si no había sesión activa), 500 Internal Server Error |


## Middleware y Permisos

La API utiliza un enfoque basado en middleware para manejar la autenticación y autorización en las rutas protegidas. Los middlewares se aplican a los handlers en la función `main` al momento de registrar las rutas.

-   **`requireAuth`:** Este middleware es el primero en la cadena para las rutas que requieren que el usuario esté logueado.
    1.  Intenta obtener la cookie de sesión (`session_id`) de la petición entrante.
    2.  Si no encuentra la cookie, responde inmediatamente con `401 Unauthorized`.
    3.  Si encuentra la cookie, obtiene el `session_id` de su valor.
    4.  **Valida el `session_id`** consultando el almacenamiento de sesiones (gestionado en el paquete `usuario`). Verifica si ese ID de sesión existe y es válido.
    5.  Si el `session_id` no es válido (no existe en el mapa de sesiones, quizás porque el servidor se reinició o la sesión fue eliminada), responde con `401 Unauthorized` y setea una cookie expirada en la respuesta para que el navegador la elimine.
    6.  Si el `session_id` es válido, recupera el objeto `Usuario` autenticado asociado a ese ID de sesión (consultando el almacenamiento de usuarios en el paquete `usuario`).
    7.  Adjunta el objeto `Usuario` autenticado al [contexto](https://pkg.go.dev/context) de la petición (`r.WithContext`) bajo una clave específica (`ContextKeyUsuarioAutenticado`).
    8.  Llama al siguiente handler en la cadena (pasándole la petición con el contexto modificado).

-   **`requireRole`:** Esta es una fábrica de middleware que se usa DESPUÉS de `requireAuth` para rutas que requieren un rol específico. Toma uno o más nombres de roles permitidos como argumento (ej: `requireRole("admin")`).
    1.  Recupera el objeto `Usuario` del contexto de la petición (asume que `requireAuth` ya lo puso allí). Si no lo encuentra (lo cual indicaría un error en la cadena de middlewares), responde con `500 Internal Server Error`.
    2.  Verifica si el `Rol` del `Usuario` autenticado coincide con alguno de los roles pasados como argumento a la fábrica (`roles ...string`).
    3.  Si el rol del usuario NO está entre los roles permitidos, responde con `403 Forbidden`.
    4.  Si el rol del usuario SÍ está permitido, llama al siguiente handler en la cadena (el handler final de la ruta).

**Rutas Protegidas y Cadena de Middlewares:**

-   `GET /api/v1/productos` (Listar): No protegida. Cadena: `apiHandler` -> `listarProductosHandler`.
-   `POST /api/v1/productos` (Crear): Protegida (cualquier usuario logueado). Cadena: `apiHandler` -> `requireAuth` -> `crearProductoHandler`.
-   `GET /api/v1/productos/{id}` (Obtener por ID): No protegida. Cadena: `apiHandler` -> `obtenerProductoHandler`.
-   `PUT /api/v1/productos/{id}` (Actualizar por ID): Protegida (cualquier usuario logueado). Cadena: `apiHandler` -> `requireAuth` -> `actualizarProductoHandler`.
-   `DELETE /api/v1/productos/{id}` (Eliminar por ID): Protegida (solo Admin). Cadena: `apiHandler` -> `requireAuth` -> `requireRole("admin")` -> `eliminarProductoHandler`.
-   Rutas de Autenticación (`/api/auth/*`): No necesitan `requireAuth` ni `requireRole` ya que son para gestionar la sesión misma. Cadena: `apiHandler` -> `[handler de auth]`.

---

## 🚀 Cómo Ejecutar el Servidor Go

Este servidor fue desarrollado en Go. Necesitas tener Go instalado en tu sistema (versión 1.22 o superior es recomendada por el uso de `r.PathValue` en el routing y otras características).

1.  **Clonar/Descargar el Proyecto:** Obtén el código fuente del proyecto (ej: `git clone tu_repo_aqui`).
2.  **Abrir Terminal:** Navega en tu terminal a la raíz del directorio del proyecto (`web-workshop-eval3-producto/` o como hayas nombrado la carpeta).
3.  **Descargar Dependencias:** El proyecto utiliza algunas librerías externas (`bcrypt`, `uuid`). Descárgalas con el siguiente comando. Este comando lee el archivo `go.mod` y descarga las dependencias listadas.
    ```bash
    go mod download
    ```
    Asegúrate de tener conexión a internet para este paso.
4.  **Compilar la Aplicación:** Go compila el código fuente en un archivo ejecutable binario. Usa el siguiente comando. La bandera `-o` especifica el nombre del archivo de salida (`main.server` en este caso).
    ```bash
    go build -o main.server web/main.server.go
    ```
    Esto creará un archivo llamado `main.server` (en Linux/macOS) o `main.server.exe` (en Windows) en la raíz del proyecto. Si hay errores de compilación, la terminal te lo indicará. Debes resolverlos antes de continuar.
5.  **Ejecutar el Servidor:** Ejecuta el archivo compilado.
   
   - go run web/main.server.go

    El servidor se iniciará y **deberías ver en la terminal mensajes de log similar a esto:**
    ```
    ⏳ Inicializando datos de ejemplo de productos...
    ✅ Inicializados X productos de ejemplo.
    ⏳ Registrando usuarios de ejemplo 'admin' y 'user'...
    ✅ Usuario de ejemplo 'admin' registrado.
    ✅ Usuario de ejemplo 'user' registrado.
    ✅ Usuarios de ejemplo registrados.
    🚀 Servidor iniciado en http://localhost:8080
    ```
    **Mantén esta terminal abierta y ejecutándose** mientras pruebas la API o usas el cliente web. El servidor escuchará peticiones HTTP en `http://localhost:8080`.


## 🧪 Pruebas de Backend con cURL

Una vez que el servidor Go está ejecutándose (ver sección anterior), puedes usar la herramienta de línea de comandos `curl` para enviar peticiones HTTP directamente a la API y verificar su correcto funcionamiento antes de usar el cliente web. `curl` suele estar preinstalado en la mayoría de los sistemas operativos o es fácil de instalar.

Abre una **segunda terminal** (manteniendo el servidor corriendo en la primera) para ejecutar los siguientes comandos.

**Nota:** Usaremos el flag `-v` en `curl` para ver detalles completos de la petición y la respuesta (incluyendo cabeceras importantes como `Set-Cookie`, `Cookie`, `Access-Control-...` y los códigos de estado HTTP). Usaremos `-c <archivo.txt>` para indicar a `curl` que **guarde** las cookies recibidas en la respuesta en el archivo especificado (ej: `-c cookiejar.txt`). Usaremos `-b <archivo.txt>` para indicar a `curl` que **lea** las cookies del archivo especificado y las **envíe** en la cabecera `Cookie` de la petición, simulando el comportamiento del navegador.

1.  **Verificar Servidor y CORS (GET /api/v1/productos):** Prueba una petición básica para confirmar que el servidor responde y setea las cabeceras CORS correctamente para el origen de tu frontend (asumiendo que configuras CORS para responder a `http://localhost:5500`).
    ```bash
    curl -v -H "Origin: http://localhost:5500" http://localhost:8080/api/v1/productos
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 200 OK`, la cabecera `Access-Control-Allow-Origin: http://localhost:5500` y un cuerpo JSON con la lista de productos de ejemplo.

2.  **Registrar un Usuario:**
    ```bash
    curl -v -H "Content-Type: application/json" -H "Origin: http://localhost:5500" -d '{"username": "testuser", "password": "securepassword123"}' http://localhost:8080/api/auth/register
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 201 Created`, `Access-Control-Allow-Origin: http://localhost:5500` y un cuerpo JSON de éxito (`"message": "Usuario registrado exitosamente", ...`).
    * *Prueba Adicional:* Intenta registrar el mismo usuario de nuevo. Deberías obtener `HTTP/1.1 409 Conflict` y un mensaje de error.

3.  **Iniciar Sesión (Guardar Cookie para 'user'):** Usaremos el usuario de ejemplo 'user' (`user123`). Guarda la cookie en `user_cookiejar.txt`.
    ```bash
    curl -v -c user_cookiejar.txt -H "Content-Type: application/json" -H "Origin: http://localhost:5500" -d '{"username": "user", "password": "user123"}' http://localhost:8080/api/auth/login
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 200 OK`, un cuerpo JSON de éxito (sin `session_id`), `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials: true` y **la cabecera `Set-Cookie: session_id=...`**. Verifica que se creó o actualizó el archivo `user_cookiejar.txt`.

4.  **Iniciar Sesión como Admin (Guardar Cookie para 'admin'):** Usa el usuario de ejemplo 'admin' (`admin123`). Guarda la cookie en `admin_cookiejar.txt`.
    ```bash
    curl -v -c admin_cookiejar.txt -H "Content-Type: application/json" -H "Origin: http://localhost:5500" -d '{"username": "admin", "password": "admin123"}' http://localhost:8080/api/auth/login
    ```
    * **Verificar:** `HTTP/1.1 200 OK`, JSON de éxito, `Set-Cookie`, `Access-Control-Allow-Origin`, `Access-Control-Allow-Credentials: true`. Archivo `admin_cookiejar.txt` creado/actualizado.

5.  **Acceso a Ruta Protegida (POST Crear Producto) Sin Autenticar:** Intenta crear un producto sin enviar ninguna cookie (simulando un usuario no logueado).
    ```bash
    curl -v -H "Content-Type: application/json" -H "Origin: http://localhost:5500" -d '{"nombre": "Producto Bloqueado", "precio": 1.0}' http://localhost:8080/api/v1/productos
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 401 Unauthorized` y un mensaje de error ("Autenticación requerida..."). Los logs del servidor deberían mostrar que `requireAuth` bloqueó la petición.

6.  **Acceso a Ruta Protegida (POST Crear Producto) Autenticado (Usuario 'user'):** Usa el archivo de cookie del usuario normal (`-b user_cookiejar.txt`).
    ```bash
    curl -v -b user_cookiejar.txt -H "Content-Type: application/json" -H "Origin: http://localhost:5500" -d '{"nombre": "Producto Creado por User", "descripcion": "Demo", "precio": 10.0, "stock": 5}' http://localhost:8080/api/v1/productos
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 201 Created` y el JSON del producto creado. Los logs deberían mostrar que `requireAuth` pasó y `crearProductoHandler` se ejecutó.

7.  **Acceso a Ruta Protegida (PUT Actualizar Producto) Autenticado (Usuario 'user'):** Actualiza uno de los productos de ejemplo o uno creado (ej: ID `1`). Usa el archivo de cookie del usuario normal.
    ```bash
    curl -v -b user_cookiejar.txt -H "Content-Type: application/json" -H "Origin: http://localhost:5500" -d '{"nombre": "Laptop Gamer UPD user", "precio": 1250.00}' http://localhost:8080/api/v1/productos/1
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 200 OK` y el JSON del producto actualizado. Logs mostrando que `requireAuth` pasó.

8.  **Acceso a Ruta Protegida (DELETE Eliminar Producto) Autenticado (Usuario 'user'):** Intenta eliminar un producto (ej: ID `1`). Usa el archivo de cookie del usuario normal.
    ```bash
    curl -v -b user_cookiejar.txt -H "Origin: http://localhost:5500" -X DELETE http://localhost:8080/api/v1/productos/1
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 403 Forbidden` y un mensaje de error ("Permiso denegado."). Logs mostrando que `requireAuth` pasó, pero `requireRole` bloqueó la petición.

9.  **Acceso a Ruta Protegida (DELETE Eliminar Producto) Autenticado (Usuario 'admin'):** Usa el archivo de cookie del usuario admin (`-b admin_cookiejar.txt`). Elimina un producto (ej: ID 2).
    ```bash
    curl -v -b admin_cookiejar.txt -H "Origin: http://localhost:5500" -X DELETE http://localhost:8080/api/v1/productos/2
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 204 No Content`. Logs mostrando que `requireAuth` y `requireRole` pasaron.

10. **Cerrar Sesión (para usuario 'user'):** Usa el archivo de cookie del usuario normal.
    ```bash
    curl -v -b user_cookiejar.txt -H "Origin: http://localhost:5500" -X POST http://localhost:8080/api/auth/logout
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 204 No Content`. En las cabeceras (`-v`), busca `Set-Cookie: session_id=; Expires=(una fecha pasada)`. Los logs deberían mostrar que la sesión fue eliminada del almacenamiento en memoria.

11. **Cerrar Sesión (Sin Cookie):** Elimina el archivo `user_cookiejar.txt` o usa un nombre diferente en `-b`.
    ```bash
    curl -v -b non_existent_cookiejar.txt -H "Origin: http://localhost:5500" -X POST http://localhost:8080/api/auth/logout
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 204 No Content`. Los logs deberían mostrar "Intento de logout sin cookie de sesión activa.".

12. **Acceso a Ruta Protegida Después de Cerrar Sesión:** Intenta una petición protegida (ej: POST crear) usando el archivo `user_cookiejar.txt` después de haber hecho logout.
    ```bash
    curl -v -b user_cookiejar.txt -H "Content-Type: application/json" -H "Origin: http://localhost:5500" -d '{"nombre": "Post Logout", "precio": 1.0}' http://localhost:8080/api/v1/productos
    ```
    * **Verificar:** La respuesta incluye `HTTP/1.1 401 Unauthorized`.

