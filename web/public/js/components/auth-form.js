// web/public/js/components/auth-form.js

// Define la plantilla HTML para el componente
const authFormTemplate = document.createElement('template');
authFormTemplate.innerHTML = `
    <style>
        /* Mantén los estilos que ya añadiste */
        .auth-container {
            max-width: 400px;
            margin: 20px auto;
            padding: 20px;
            border: 1px solid #ccc;
            border-radius: 5px;
            box-shadow: 2px 2px 10px rgba(0, 0, 0, 0.1);
            text-align: center;
        }
        h2 {
            text-align: center;
            margin-bottom: 20px;
            color: #333;
        }
        .form-group {
            margin-bottom: 15px;
            text-align: left;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: #555;
        }
        input[type="text"],
        input[type="password"] {
            width: calc(100% - 22px);
            padding: 10px;
            border: 1px solid #ccc;
            border-radius: 4px;
            box-sizing: border-box;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 10px 15px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            width: 100%;
            margin-top: 10px;
            transition: background-color 0.3s ease;
        }
        button:hover {
            background-color: #0056b3;
        }
        .toggle-mode {
            margin-top: 15px;
            font-size: 0.9em;
        }
        .toggle-mode a {
            color: #007bff;
            text-decoration: none;
            cursor: pointer;
        }
        .toggle-mode a:hover {
             text-decoration: underline;
        }
         .error-message {
            color: red;
            font-size: 0.9em;
            margin-top: 10px;
            text-align: center;
        }

    </style>

    <div class="auth-container">
        <h2 id="form-title">Iniciar Sesión</h2>
        <div id="auth-form">
             </div>
        <p id="error-display" class="error-message" style="display: none;"></p>
        <p class="toggle-mode">
            <span id="toggle-text"></span> <a href="#" id="toggle-link"></a>.
        </p>
    </div>
`;

// Define la clase Custom Element
class AuthForm extends HTMLElement {
    constructor() {
        super();

        this.attachShadow({ mode: 'open' });
        this.shadowRoot.appendChild(authFormTemplate.content.cloneNode(true));

        // Referencias a elementos clave
        this.$formTitle = this.shadowRoot.getElementById('form-title');
        this.$authForm = this.shadowRoot.getElementById('auth-form');
        this.$toggleText = this.shadowRoot.getElementById('toggle-text');
        this.$toggleLink = this.shadowRoot.getElementById('toggle-link');
        this.$errorDisplay = this.shadowRoot.getElementById('error-display');

        this._mode = 'login'; // Estado inicial

        this._renderForm();
        this._attachEventListeners();

        console.log("AuthForm constructor ejecutado.");
    }

    // --- Ciclo de Vida del Componente ---
    connectedCallback() {
        console.log("AuthForm añadido al DOM.");
    }
    disconnectedCallback() {
        console.log("AuthForm removido del DOM.");
    }
    // observedAttributes, attributeChangedCallback (si decides usar atributos para el modo)


    // --- Métodos Internos ---

    _renderForm() {
        this.$authForm.innerHTML = '';

        let usernameField = `
             <div class="form-group">
                <label for="username">Nombre de Usuario:</label>
                <input type="text" id="username" name="username" required autocomplete="${this._mode === 'login' ? 'current-username' : 'new-username'}">
            </div>
        `;
        let passwordField = `
             <div class="form-group">
                <label for="password">Contraseña:</label>
                <input type="password" id="password" name="password" required autocomplete="${this._mode === 'login' ? 'current-password' : 'new-password'}">
            </div>
        `;
        let submitButton;

        if (this._mode === 'login') {
            this.$formTitle.textContent = 'Iniciar Sesión';
            submitButton = '<button id="submit-btn">Iniciar Sesión</button>';
            this.$toggleText.textContent = '¿No tienes cuenta?';
            this.$toggleLink.textContent = 'Regístrate aquí';
        } else { // register mode
            this.$formTitle.textContent = 'Registrarse';
            submitButton = '<button id="submit-btn">Registrarse</button>';
            this.$toggleText.textContent = '¿Ya tienes cuenta?';
            this.$toggleLink.textContent = 'Inicia sesión aquí';
        }

        this.$authForm.innerHTML = usernameField + passwordField + submitButton;
        this.hideError();
    }

    _attachEventListeners() {
        this.$toggleLink.addEventListener('click', (e) => {
            e.preventDefault();
            this.toggleMode();
        });

        // Listener para el botón de submit dentro del formulario
        // Es mejor añadir el listener al formulario y escuchar el evento 'submit'
         // o al contenedor del formulario y escuchar 'click' en el botón
         // Ya que el botón se re-renderiza, el listener en $authForm (el contenedor) es válido
        this.$authForm.addEventListener('click', (e) => {
            if (e.target && e.target.id === 'submit-btn') {
                 // Llamamos a _handleSubmit cuando se hace clic en el botón
                this._handleSubmit(e);
            }
        });
    }

    // Marcar _handleSubmit como asíncrono porque llamará a funciones asíncronas (fetch)
    async _handleSubmit(event) {
         event.preventDefault();

         this.hideError();

         const usernameInput = this.shadowRoot.getElementById('username');
         const passwordInput = this.shadowRoot.getElementById('password');

         const username = usernameInput.value.trim();
         const password = passwordInput.value.trim();

         if (!username || !password) {
             this.showError("Nombre de usuario y contraseña no pueden estar vacíos.");
             return;
         }

         console.log(`Intentando ${this._mode} con usuario: ${username}`);

         // Deshabilitar el botón mientras se procesa la petición (buena UX)
         const submitBtn = this.shadowRoot.getElementById('submit-btn');
         if(submitBtn) submitBtn.disabled = true;


         try {
             if (this._mode === 'login') {
                 await this._callLoginApi(username, password);
             } else { // register
                 await this._callRegisterApi(username, password);
             }
         } catch (error) {
             console.error("Error en la petición de autenticación:", error);
             // Manejar errores de red u otros errores que no vengan de la API con un status no-OK
             this.showError("Error de comunicación con el servidor.");
         } finally {
             // Habilitar el botón de nuevo después de la petición
             if(submitBtn) submitBtn.disabled = false;
         }
    }

    // --- Métodos Públicos ---

    toggleMode() {
        this._mode = this._mode === 'login' ? 'register' : 'login';
        this._renderForm();
        console.log(`Modo de formulario cambiado a: ${this._mode}`);
    }

    showError(message) {
        this.$errorDisplay.textContent = message;
        this.$errorDisplay.style.display = 'block';
    }

    hideError() {
        this.$errorDisplay.textContent = '';
        this.$errorDisplay.style.display = 'none';
    }

    // Método para emitir un evento personalizado cuando la autenticación es exitosa
    _dispatchAuthSuccessEvent(userData) {
        // Creamos un CustomEvent con un nombre descriptivo
        const event = new CustomEvent('auth-success', {
            bubbles: true, // El evento "burbujea" por el DOM, puede ser escuchado por padres
            composed: true, // El evento puede pasar el límite del Shadow DOM
            detail: { // Datos personalizados adjuntos al evento
                user: userData
            }
        });
        this.dispatchEvent(event); // Disparamos el evento desde este componente
        console.log("Evento 'auth-success' disparado.");
    }


    // --- Implementación de las llamadas a la API con fetch ---

    // Llamada a la API de Login
    async _callLoginApi(username, password) {
         console.log(`Llamando a POST /api/auth/login para ${username}...`);

         const url = '/api/auth/login'; // Ruta relativa a la raíz del servidor
         const method = 'POST';
         const headers = {
             'Content-Type': 'application/json' // Indicamos que enviamos JSON
         };
         const body = JSON.stringify({ username, password }); // Convertimos el objeto JS a string JSON

         try {
             const response = await fetch(url, {
                 method: method,
                 headers: headers,
                 body: body
                 // No necesitamos especificar 'credentials: "include"' si el frontend y backend
                 // están en el mismo origen (http://localhost:8080). Si están en orígenes diferentes
                 // (ej: frontend en 5500, backend en 8080), y configuraste CORS con
                 // Access-Control-Allow-Credentials: true en el backend, el navegador enviará/recibirá
                 // la cookie automáticamente. Si tienes problemas de cookies con CORS, podrías
                 // intentar añadir 'credentials: "include"' aquí, pero normalmente no es necesario
                 // si el backend CORS está bien configurado para ello.
             });

             // La API de Go devuelve JSON tanto en éxito (200) como en errores (401, 409, etc.)
             // Intentamos siempre leer el cuerpo como JSON.
             const responseData = await response.json(); // Leemos el cuerpo como JSON

             if (response.ok) { // response.ok es true para códigos 200-299
                 console.log("Login exitoso:", responseData);
                 this.showError("Login exitoso!"); // Mensaje temporal de éxito
                 // Ocultar el formulario o navegar a otra vista
                 // Podemos disparar un evento para que el componente padre maneje el éxito
                 this._dispatchAuthSuccessEvent(responseData);

             } else {
                 // Manejar errores del backend (401, 409, 500, etc.)
                 console.error(`Error en Login (${response.status}):`, responseData);
                 // Mostrar el mensaje de error proporcionado por la API (si existe)
                 const errorMessage = responseData.message || `Error ${response.status}: ${response.statusText}`;
                 this.showError(`Error al iniciar sesión: ${errorMessage}`);
             }

         } catch (error) {
             // Manejar errores de red (servidor caído, conexión rechazada, etc.)
             console.error("Error de red durante la petición de Login:", error);
             this.showError("Error de conexión. Asegúrate que el servidor esté corriendo.");
         }
    }

    // Llamada a la API de Registro
    async _callRegisterApi(username, password) {
        console.log(`Llamando a POST /api/auth/register para ${username}...`);

        const url = '/api/auth/register'; // Ruta relativa
        const method = 'POST';
        const headers = {
            'Content-Type': 'application/json'
        };
        const body = JSON.stringify({ username, password });

        try {
            const response = await fetch(url, {
                method: method,
                headers: headers,
                body: body
            });

            const responseData = await response.json();

            if (response.ok) { // 201 Created es ok
                console.log("Registro exitoso:", responseData);
                this.showError("Registro exitoso! Ahora puedes iniciar sesión."); // Mensaje de éxito

                // Opcional: Cambiar al modo login automáticamente después del registro exitoso
                // this.toggleMode();

            } else {
                 // Manejar errores del backend (409 Conflict, 400 Bad Request, 500 Internal Server Error)
                console.error(`Error en Registro (${response.status}):`, responseData);
                 const errorMessage = responseData.message || `Error ${response.status}: ${response.statusText}`;
                 this.showError(`Error al registrarse: ${errorMessage}`);
            }

        } catch (error) {
             // Manejar errores de red
            console.error("Error de red durante la petición de Registro:", error);
            this.showError("Error de conexión. Asegúrate que el servidor esté corriendo.");
        }
    }


    // Puedes añadir métodos públicos para, por ejemplo, establecer el modo desde fuera
    // set mode(value) {
    //    if (value === 'login' || value === 'register') {
    //         this._mode = value;
    //         this._renderForm();
    //     } else {
    //         console.error("Modo inválido para AuthForm. Debe ser 'login' o 'register'.");
    //     }
    // }
    // get mode() { return this._mode; }

}

// 4. Definir el Custom Element en el registro del navegador
customElements.define('auth-form', AuthForm);

console.log("Componente auth-form.js cargado y definido.");